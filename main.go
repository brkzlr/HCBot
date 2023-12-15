package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"
)

func processChatRequests(session *discordgo.Session) {
	chatRequestLock.Lock()
	requestQueue := slices.Clone(chatRequestQueue)
	chatRequestQueue = nil
	chatRequestLock.Unlock()

	for _, request := range requestQueue {
		contextMessages = append(contextMessages,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: request.Prompt,
			})
		if len(contextMessages) > 51 {
			contextMessages = slices.Delete(contextMessages, 1, 3)
		}

		resp, err := aiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    "ft:gpt-3.5-turbo-1106:hc::8J4SHaOm",
				Messages: contextMessages,
			})

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			ReplyToMsg(session, request.MessageEvent.Message, "Sorry! I encountered a ChatCompletion error! Please retry later.")
			contextMessages = contextMessages[:len(contextMessages)-1] // delete last entry
			return
		}

		respContent := resp.Choices[0].Message.Content
		ReplyToMsg(session, request.MessageEvent.Message, respContent)

		contextMessages = append(contextMessages,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: respContent,
			})
		dirtyContextMsg = true
	}
}

func saveChatContext() {
	if !dirtyContextMsg {
		return
	}

	contextFile, err := os.Open("context.json")
	if err != nil {
		fmt.Println("Error opening context.json to save!")
		return
	}
	defer contextFile.Close()

	jsonArr, err := json.Marshal(contextMessages)
	if err != nil {
		fmt.Println("Error marshaling context messages!")
		return
	}

	os.WriteFile("context.json", jsonArr, 0644)
	fmt.Println("Saved context messages successfully!")
	dirtyContextMsg = false
}

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

	// Grab GPT AI context messages
	contextFile, err := os.Open("context.json")
	if err != nil {
		fmt.Println("Error opening context.json! Aborting!")
		return
	}
	defer contextFile.Close()

	fileByte, _ = io.ReadAll(contextFile)
	json.Unmarshal(fileByte, &contextMessages)

	go KeepAliveRequest() //Do a simple request to OpenXBL so token is authenticated
}

func main() {
	discord, err := discordgo.New("Bot " + tokens.Discord)
	if err != nil {
		fmt.Println("Error creating Discord session!", err)
		return
	}

	aiClient = openai.NewClient(tokens.OpenAI)

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
			saveChatContext()
			saveDatabase()
			go KeepAliveRequest()
		case <-achievTicker.C:
			CheckTimedAchievs(discord)
		case <-sc:
			loop = false
		default:
			processChatRequests(discord)
		}
	}

	// V this might not be even needed V
	discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, hcGuildID, nil) // Delete all application (slash) commands
	saveChatContext()
	saveDatabase()
}
