package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type GTResp struct {
	ID string `json:"id"`
}

type AchievementsStats struct {
	CurrentGScore int `json:"currentGamerscore"`
	TotalGScore   int `json:"totalGamerscore"`
}

type Game struct {
	Stats   AchievementsStats `json:"achievement"`
	TitleID string            `json:"titleId"`
	Name    string            `json:"name"`
}

func RequestPlayerAchievements(discordID string) ([]Game, error) {
	xbID, ok := DatabaseMap[discordID]
	if !ok {
		return nil, errors.New("Please set your gamertag first using the command `+gt \"gamertag\"`")
	}

	url := "https://xbl.io/api/v2/achievements/player/" + xbID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Authorization", Tokens.OpenXBL)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
	}

	decoder := json.NewDecoder(resp.Body)
	var objMap map[string]json.RawMessage
	err = decoder.Decode(&objMap)
	if err != nil {
		return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
	}

	var games []Game
	err = json.Unmarshal(objMap["titles"], &games)
	if err != nil {
		return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
	}

	if len(games) == 0 {
		return nil, errors.New("You have either not played any games or your Xbox profile is private.")
	}

	return games, nil
}

func RequestPlayerGT(gamerTag string) (string, error) {
	urlTag := strings.ReplaceAll(gamerTag, " ", "%20")
	url := "https://xbl.io/api/v2/friends/search?gt=" + urlTag

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("X-Authorization", Tokens.OpenXBL)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("Whoops! Server responded with an error! Apologies, please try again!")
	}

	decoder := json.NewDecoder(resp.Body)
	var objMap map[string]json.RawMessage
	err = decoder.Decode(&objMap)
	if err != nil {
		return "", errors.New("Server responded with garbage! Not your fault. Please try again now!")
	}

	var respID []GTResp
	err = json.Unmarshal(objMap["profileUsers"], &respID)
	if err != nil {
		return "", errors.New("Hmm, that gamertag didn't work! For names using the new gamertag system, please do not put a hashtag before the numbers.")
	}

	return respID[0].ID, nil
}

func AddGamertagToDB(discordID, xblID string) {
	GlobalLock.Lock()
	DatabaseMap[discordID] = xblID
	GlobalLock.Unlock()
}

func HasRole(member *discordgo.Member, roleID string) bool {
	for _, id := range member.Roles {
		if id == roleID {
			return true
		}
	}

	return false
}

func HasRoles(member *discordgo.Member, rolesID []string) map[string]bool {
	rolesMap := make(map[string]bool)
	for _, searchID := range rolesID {
		rolesMap[searchID] = false
	}

	for _, id := range member.Roles {
		if _, exists := rolesMap[id]; exists {
			rolesMap[id] = true
		}
	}

	return rolesMap
}

func GetCompletionSymbol(gameCompl bool) string {
	if gameCompl {
		return "✅"
	} else {
		return "❌"
	}
}

func ReplyToMsg(s *discordgo.Session, m *discordgo.Message, replyMsg string) {
	s.ChannelMessageSendReply(m.ChannelID, replyMsg, &discordgo.MessageReference{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
	})
}

func ReactSuccess(s *discordgo.Session, m *discordgo.Message) {
	s.MessageReactionsRemoveEmoji(m.ChannelID, m.ID, "⚙️")
	s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
}

func ReactFail(s *discordgo.Session, m *discordgo.Message) {
	s.MessageReactionsRemoveEmoji(m.ChannelID, m.ID, "⚙️")
	s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
}

func LogCommand(cmdName, author string) {
	fmt.Println(cmdName + " command used - " + author)
}
