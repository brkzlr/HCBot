package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func clearCooldowns() {
	var idsToRemove []string
	cooldownLock.Lock()
	for discordID, expirationTime := range cooldownMap {
		if expirationTime.Before(time.Now()) {
			idsToRemove = append(idsToRemove, discordID)
		}
	}
	for _, ids := range idsToRemove {
		delete(cooldownMap, ids)
	}
	cooldownLock.Unlock()
}

func saveDatabase() {
	if !dirtyDatabase {
		return
	}

	dbFile, err := os.Open("database.json")
	if err != nil {
		log.Println("Error opening database to save! ", err)
		return
	}
	defer dbFile.Close()

	jsonMap, err := json.Marshal(databaseMap)
	if err != nil {
		log.Println("Error marshaling database! ", err)
		return
	}

	os.WriteFile("database.json", jsonMap, 0644)
	infoLog.Println("Saved database successfully!")
	dirtyDatabase = false
}

var isTest bool
var guildID string
var infoLog log.Logger

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

	// Grab guild users' xuids
	dbFile, err := os.Open("database.json")
	if err != nil {
		log.Fatal("Error opening database.json! Aborting!")
	}
	defer dbFile.Close()

	fileByte, _ = io.ReadAll(dbFile)
	json.Unmarshal(fileByte, &databaseMap)

	go KeepAliveRequest() //Do a simple request to OpenXBL so token is authenticated
}

func main() {
	discord, err := discordgo.New("Bot " + tokens.Discord)
	if err != nil {
		log.Fatal("Error creating Discord session! ", err)
	}

	discord.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers | discordgo.IntentMessageContent
	discord.AddHandler(messageCreate)

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
			clearCooldowns()
			saveDatabase()
			go KeepAliveRequest()
		case <-achievTicker.C:
			CheckTimedAchievs(discord)
		case <-sc:
			break MainLoop
		}
	}

	discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, guildID, nil) // Delete all application (slash) commands
	saveDatabase()
}
