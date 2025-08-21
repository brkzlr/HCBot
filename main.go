package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

var isTest bool
var guildID string
var infoLog log.Logger

func initDatabase(dbPath string) error {
	var err error
	database, err = sql.Open("sqlite3", dbPath)
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
		log.Fatal("Error opening tokens.json! Aborting!")
	}
	defer jsonFile.Close()

	fileByte, _ := io.ReadAll(jsonFile)
	json.Unmarshal(fileByte, &tokens)

	err = initDatabase("database.db")
	if err != nil {
		log.Fatalf("Error opening/creating database! Error: %s", err)
	}

	go KeepAliveRequest() // Do a simple request to OpenXBL so token is authenticated
}

func main() {
	discord, err := discordgo.New("Bot " + tokens.Discord)
	if err != nil {
		log.Fatal("Error creating Discord session! ", err)
	}

	discord.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers | discordgo.IntentMessageContent
	discord.AddHandler(messageCreate)
	discord.AddHandler(messageReactionAdd)

	err = discord.Open()
	if err != nil {
		log.Fatal("Error opening connection! ", err)
	}
	defer discord.Close()
	infoLog.Println("Bot is now running!")

	err = InitCommands(discord)
	if err != nil {
		log.Fatal("Error initializing commands! ", err)
	}
	infoLog.Println("Commands initialised!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	ticker := time.NewTicker(1 * time.Hour)
	achievTicker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	defer achievTicker.Stop()

MainLoop:
	for {
		select {
		case <-ticker.C:
			go KeepAliveRequest()
		case <-achievTicker.C:
			CheckTimedAchievs(discord)
		case <-sc:
			break MainLoop
		}
	}

	discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, guildID, nil) // Delete all application (slash) commands
}
