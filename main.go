package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func saveDatabase() {
	if !dirtyDatabase {
		return
	}

	dbFile, err := os.Open("database.json")
	if err != nil {
		fmt.Println("Error opening database to save!")
		return
	}
	defer dbFile.Close()

	jsonMap, err := json.Marshal(databaseMap)
	if err != nil {
		fmt.Println("Error marshaling database!")
		return
	}

	os.WriteFile("database.json", jsonMap, 0644)
	fmt.Println("Saved database successfully!")
	dirtyDatabase = false
}

func init() {
	// Grab Discord and OpenXBL Tokens
	jsonFile, err := os.Open("tokens.json")
	if err != nil {
		fmt.Println("Error opening tokens.json! Aborting!")
		return
	}
	defer jsonFile.Close()

	fileByte, _ := io.ReadAll(jsonFile)
	json.Unmarshal(fileByte, &tokens)

	// Grab guild users' xuids
	dbFile, err := os.Open("database.json")
	if err != nil {
		fmt.Println("Error opening database.json! Aborting!")
		return
	}
	defer dbFile.Close()

	fileByte, _ = io.ReadAll(dbFile)
	json.Unmarshal(fileByte, &databaseMap)

	go KeepAliveRequest() //Do a simple request to OpenXBL so token is authenticated
}

func main() {
	discord, err := discordgo.New("Bot " + tokens.Discord)
	if err != nil {
		fmt.Println("Error creating Discord session!", err)
		return
	}

	discord.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers | discordgo.IntentMessageContent
	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening connection!", err)
		return
	}
	defer discord.Close()
	fmt.Println("Bot is now running!")

	InitCommands(discord)
	fmt.Println("Commands initialised!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	ticker := time.NewTicker(1 * time.Hour)
	achievTicker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	defer achievTicker.Stop()

	for loop := true; loop; {
		select {
		case <-ticker.C:
			saveDatabase()
			go KeepAliveRequest()
		case <-achievTicker.C:
			CheckTimedAchievs(discord)
		case <-sc:
			loop = false
		}
	}

	// V this might not be even needed V
	discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, hcGuildID, nil) // Delete all application (slash) commands
	saveDatabase()
}
