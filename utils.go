package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

type Riddle struct {
	Question string `json:"riddle"`
	Answer   string `json:"answer"`
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
	//Gamertags with a suffix should not include the hashtag
	urlTag := strings.ReplaceAll(gamerTag, "#", "")
	urlTag = strings.ReplaceAll(urlTag, " ", "%20")
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

func GetRiddle() (Riddle, error) {
	resp, err := http.Get("https://riddles-api.vercel.app/random")
	if err != nil {
		return Riddle{"", ""}, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var riddleResp Riddle
	if err = json.Unmarshal(body, &riddleResp); err != nil {
		return Riddle{"", ""}, err
	}

	return riddleResp, err
}

// Workaround until OpenXBL API changed for official Xbox API
func KeepAliveRequest() {
	req, _ := http.NewRequest("GET", "https://xbl.io/api/v2/account", nil)
	req.Header.Add("X-Authorization", Tokens.OpenXBL)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	client.Do(req)
	fmt.Println("Sent KeepAlive to OpenXBL!")
	//We don't care about the result, we just want to do a GET request on OpenXBL
	//so our token gets refreshed and future requests after idling won't fail
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

func IsStaff(member *discordgo.Member) bool {
	// Pillar / Oracle (Mod) / Guardian (Admin) roles
	staffRoles := []string{"987989822813650974", "984081125657964664", "984080972108668959"}
	result := HasRoles(member, staffRoles)
	for _, roleID := range staffRoles {
		if result[roleID] {
			return true
		}
	}

	return false
}

func GetAllGuildMembers(s *discordgo.Session, guildID string) []*discordgo.Member {
	guildMembers, _ := s.GuildMembers(guildID, "", 1000)

	//Discord API can only return a maximum of 1000 members.
	//To get all the members for guilds that have more than this limit
	//we check the length of the returned slice and if it's 1000 we try to grab
	//the next 1000 starting from the last member in the previous request using it's ID,
	//repeating this until we get a slice with less than 1000.
	gotAll := len(guildMembers) < 1000
	for !gotAll {
		lastID := guildMembers[len(guildMembers)-1].User.ID
		tempGMembers, _ := s.GuildMembers(guildID, lastID, 1000)
		if len(tempGMembers) < 1000 {
			gotAll = true
		}
		guildMembers = append(guildMembers, tempGMembers...)
	}

	return guildMembers
}

func GetCompletionSymbol(gameCompl bool) string {
	if gameCompl {
		return "✅"
	} else {
		return "❌"
	}
}

func ReplyToMsg(s *discordgo.Session, m *discordgo.Message, replyMsg string) (*discordgo.Message, error) {
	return s.ChannelMessageSendReply(m.ChannelID, replyMsg, &discordgo.MessageReference{
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
