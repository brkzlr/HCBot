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

var (
	Tokens struct {
		Discord string `json:"discordToken"`
		OpenXBL string `json:"xblToken"`
	}

	DatabaseMap   map[string]string
	DirtyDatabase bool
)

func saveDatabase() {
	if !DirtyDatabase {
		return
	}

	dbFile, err := os.Open("database.json")
	if err != nil {
		fmt.Println("Error opening database to save!")
		return
	}
	defer dbFile.Close()

	jsonMap, err := json.Marshal(DatabaseMap)
	if err != nil {
		fmt.Println("Error marshaling database!")
		return
	}

	os.WriteFile("database.json", jsonMap, 0644)
	fmt.Println("Saved database successfully!")
}

func checkTimedAchievs(session *discordgo.Session) {
	today := time.Now().UTC()
	if today.Hour() != 9 || today.Minute() != 0 { //@todo: Find a better way to activate at 9 AM UTC
		return
	}

	targetChannelID := "984160204440633454"
	if timedRole, exists := timedAchievRoles[today.Day()]; exists {
		session.ChannelMessageSend(targetChannelID,
			fmt.Sprintf("Remember to grab your <@&%d> achievement today! Simply start up a mission or load into a multiplayer game in %s", timedRole.ID, timedRole.Game))
	}

	for _, date := range destinationVacationDates {
		if today.Day() == date.Day && today.Month() == date.Month {
			session.ChannelMessageSend(targetChannelID,
				"Remember to grab your <@&990602317575368724> Achievement today! Simply load up a Custom Game on Halo 2 Classic Zanzibar, go to the beach and look at the sign next to the water!")
			break
		}
	}

	for _, date := range elderSignsDates {
		if today.Day() == date.Day && today.Month() == date.Month {
			session.ChannelMessageSend(targetChannelID,
				"Remember to grab your <@&990602348659363850> Achievement today! Simply load up a Custom Game on Halo 3 Valhalla and look at the Sigil on the wall. Remember you need to have looked at 2 different ones for it to unlock!")
			break
		}
	}
}

func init() {
	jsonFile, err := os.Open("tokens.json")
	if err != nil {
		fmt.Println("Error opening tokens.json! Aborting!")
		return
	}
	defer jsonFile.Close()

	fileByte, _ := io.ReadAll(jsonFile)
	json.Unmarshal(fileByte, &Tokens)

	dbFile, err := os.Open("database.json")
	if err != nil {
		fmt.Println("Error opening database.json! Aborting!")
		return
	}
	defer dbFile.Close()

	fileByte, _ = io.ReadAll(dbFile)
	json.Unmarshal(fileByte, &DatabaseMap)

	go KeepAliveRequest() //Do a simple request to OpenXBL so token is authenticated
}

func main() {
	discord, err := discordgo.New("Bot " + Tokens.Discord)
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
			checkTimedAchievs(discord)
		case <-sc:
			loop = false
		}
	}
	saveDatabase()
}
