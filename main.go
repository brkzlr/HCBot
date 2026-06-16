package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

const disconnectGracePeriod = 2 * time.Minute

var isTest bool
var guildID string
var infoLog log.Logger

func initDatabase(dbPath string) error {
	var err error
	database, err = sql.Open("sqlite3", dbPath+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return err
	}

	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS users (
			discordID TEXT PRIMARY KEY,
			xuid TEXT NOT NULL
		);`,
	)
	if err != nil {
		return err
	}

	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS moderation (
			discordID TEXT PRIMARY KEY,
			command_cooldown INTEGER
		);`,
	)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	// Setup logging
	log.SetPrefix("ERROR: ")
	infoLog = *log.New(os.Stdout, "INFO: ", log.LstdFlags)

	// Parse flags
	flag.BoolVar(&isTest, "test", false, "Flag to set testing mode, disabling stuff like cooldown functionality.")
	flag.StringVar(&guildID, "guild", hcGuildID, "Flag to override default guild ID set to HC server.")
	flag.Parse()

	// Grab Discord and OpenXBL Tokens
	jsonFile, err := os.Open("tokens.json")
	if err != nil {
		log.Fatalf("Error opening tokens.json! Error: %s", err)
	}
	defer jsonFile.Close()

	fileByte, _ := io.ReadAll(jsonFile)
	if err := json.Unmarshal(fileByte, &tokens); err != nil {
		log.Fatalf("Error parsing tokens.json! Error: %s", err)
	}

	err = initDatabase("database.db")
	if err != nil {
		log.Fatalf("Error opening/creating database! Error: %s", err)
	}
}

func nextNineAM(now time.Time) time.Time {
	now = now.UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func main() {
	discord, err := discordgo.New("Bot " + tokens.Discord)
	if err != nil {
		log.Fatal("Error creating Discord session! ", err)
	}

	discord.Dialer.HandshakeTimeout = 15 * time.Second
	discord.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers | discordgo.IntentMessageContent
	discord.AddHandler(messageCreate)
	discord.AddHandler(messageReactionAdd)

	// Register the watchdog before opening so it sees the first Connect event and starts tracking gateway health immediately.
	restartCh := startConnectionWatchdog(discord)

	err = discord.Open()
	if err != nil {
		log.Fatal("Error opening connection! ", err)
	}
	infoLog.Println("Bot is now running!")

	err = InitCommands(discord)
	if err != nil {
		log.Fatal("Error initializing commands! ", err)
	}
	infoLog.Println("Commands initialised!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	achievTimer := time.NewTimer(time.Until(nextNineAM(time.Now())))
	defer achievTimer.Stop()

	for {
		select {
		case <-achievTimer.C:
			go CheckTimedAchievs(discord)
			achievTimer.Reset(time.Until(nextNineAM(time.Now())))
		case <-sc:
			infoLog.Println("Received shutdown signal, exiting...")
			gracefulExit(discord, 0)
		case <-restartCh:
			log.Printf("Watchdog: gateway down for over %s, exiting for supervisor restart...", disconnectGracePeriod)
			gracefulExit(discord, 1)
		}
	}
}

func startConnectionWatchdog(s *discordgo.Session) <-chan struct{} {
	restartCh := make(chan struct{}, 1)

	var mu sync.Mutex
	timer := time.AfterFunc(disconnectGracePeriod, func() {
		select {
		case restartCh <- struct{}{}:
		default:
		}
	})
	timer.Stop()

	s.AddHandler(func(_ *discordgo.Session, _ *discordgo.Connect) {
		mu.Lock()
		timer.Stop()
		mu.Unlock()
	})
	s.AddHandler(func(_ *discordgo.Session, _ *discordgo.Disconnect) {
		log.Println("Gateway disconnected, starting reconnect grace period...")
		mu.Lock()
		timer.Reset(disconnectGracePeriod)
		mu.Unlock()
	})

	return restartCh
}

func gracefulExit(discord *discordgo.Session, code int) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, guildID, nil,
			discordgo.WithContext(ctx), discordgo.WithRetryOnRatelimit(false)); err != nil {
			log.Printf("Could not clear slash commands on shutdown: %s", err)
		}
		discord.Close()
	}()

	select {
	case <-done:
	case <-time.After(8 * time.Second):
		log.Println("Shutdown cleanup timed out, forcing exit...")
	}
	os.Exit(code)
}
