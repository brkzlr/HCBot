package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func AddGamertagToDB(discordID, xblID string) {
	databaseLock.Lock()
	databaseMap[discordID] = xblID
	databaseLock.Unlock()
	dirtyDatabase = true
}

func AddRoleCheckCooldown(discordID string) {
	cooldownLock.Lock()
	cooldownMap[discordID] = time.Now().Add(time.Hour * 1)
	cooldownLock.Unlock()
}

func CheckCooldown(discordID string) (bool, time.Duration) {
	cooldownLock.Lock()
	expirationTime, exists := cooldownMap[discordID]
	cooldownLock.Unlock()
	if exists {
		return expirationTime.After(time.Now()), expirationTime.Sub(time.Now())
	}
	return false, 0
}

// This function will query Spartan Assault & Spartan Strike legacy platform achievements
// as they can be bugged if we check the profile with RequestPlayerAchievements
// while querying individual title IDs will return the correct state
func CheckLegacyAssaultStrikeAchievements(discordID string) (map[string]GameStatus, error) {
	// We'll skip over Spartan Assault X1/XSX because this endpoint can only detect older versions
	gamesToCheck := map[string]GameStatus{
		hsaTitleID:    NOT_FOUND, // Windows version
		hsa360TitleID: NOT_FOUND,
		hsaWPTitleID:  NOT_FOUND,
		hsaIOSTitleID: NOT_FOUND,

		hssTitleID:    NOT_FOUND,
		hssWPTitleID:  NOT_FOUND,
		hssIOSTitleID: NOT_FOUND,
	}
	// Hardcode the achievement count for these game to reduce the number of calls to the API as they're unlikely to change
	// We have to check 7 games (8th being X1/XSX which can be found by the other endpoint)
	// but this endpoint will only give the achievements the player has.
	// This means we either check against hardcoded numbers or call another endpoint which will give
	// the total amount of achievements, which would double the number of calls per user to 14 just for SS/SA
	const (
		AssaultAchievCount    = 25 // Windows/WP
		AssaultIOSAchievCount = 20 // 5 unobtainable achievements in iOS
		Assault360AchievCount = 28
		StrikeAchievCount     = 20 // Spartan Strike has 20 achievements on all 3 platforms
	)

	// This function should never be called before RequestPlayerAchievements so we'll skip the checks
	xbID, _ := GetGamertagID(discordID)

	for titleID := range gamesToCheck {
		url := "https://xbl.io/api/v2/achievements/x360/" + xbID + "/title/" + titleID
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("X-Authorization", tokens.OpenXBL)
		req.Header.Add("Accept", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
		}

		var playerAchievsInfo AchievListInfo
		err = json.Unmarshal(body, &playerAchievsInfo)
		if err != nil {
			return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
		}

		achievCountToCheck := 0
		switch titleID {
		case hsaTitleID:
			fallthrough
		case hsaWPTitleID:
			achievCountToCheck = AssaultAchievCount
		case hsaIOSTitleID:
			achievCountToCheck = AssaultIOSAchievCount
		case hsa360TitleID:
			achievCountToCheck = Assault360AchievCount

		case hssTitleID:
			fallthrough
		case hssWPTitleID:
			fallthrough
		case hssIOSTitleID:
			achievCountToCheck = StrikeAchievCount
		}

		if playerAchievsInfo.PagingInfo.TotalRecords == achievCountToCheck {
			gamesToCheck[titleID] = COMPLETED
		} else if playerAchievsInfo.PagingInfo.TotalRecords == 0 {
			gamesToCheck[titleID] = NOT_FOUND
		} else {
			gamesToCheck[titleID] = NOT_COMPLETED
		}
	}

	return gamesToCheck, nil
}

func CheckTimedAchievs(session *discordgo.Session) {
	today := time.Now().UTC()
	if today.Hour() != 9 || today.Minute() != 0 { // TODO: Find a better way to activate at 9 AM UTC
		return
	}

	targetChannelID := "984160204440633454" // general-mcc channel ID
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

func GetCompletionSymbol(status GameStatus) string {
	switch status {
	case NOT_FOUND:
		return "❌"
	case NOT_COMPLETED:
		return "🔶"
	case COMPLETED:
		return "✅"
	}

	return "❔"
}

func GetGamertagID(discordID string) (xuid string, exists bool) {
	databaseLock.Lock()
	xuid, exists = databaseMap[discordID]
	databaseLock.Unlock()
	return
}

func GetRiddle() (Riddle, error) {
	resp, err := http.Get("https://riddles-api.vercel.app/random")
	if err != nil {
		return Riddle{"", ""}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var riddleResp Riddle
	if err = json.Unmarshal(body, &riddleResp); err != nil {
		return Riddle{"", ""}, err
	}

	return riddleResp, err
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
	// Pillar / Oracle (Mod) / Guardian (Admin) / Founder roles
	staffRoles := []string{"987989822813650974", "984081125657964664", "984080972108668959", "1075504782023852102"}
	result := HasRoles(member, staffRoles)
	for _, roleID := range staffRoles {
		if result[roleID] {
			return true
		}
	}

	return false
}

func LogCommand(cmdName, author string) {
	infoLog.Println(cmdName + " command used by " + author)
}

func KeepAliveRequest() {
	req, _ := http.NewRequest("GET", "https://xbl.io/api/v2/account", nil)
	req.Header.Add("X-Authorization", tokens.OpenXBL)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	//We don't care about the result, we just want to do a GET request on OpenXBL
	//so our token gets refreshed and future requests after idling won't fail
}

func ReplyToMsg(s *discordgo.Session, m *discordgo.Message, replyMsg string) (*discordgo.Message, error) {
	return s.ChannelMessageSendReply(m.ChannelID, replyMsg, &discordgo.MessageReference{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
	})
}

func RespondACKToInteraction(s *discordgo.Session, i *discordgo.Interaction) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func RespondFollowUpToInteraction(s *discordgo.Session, i *discordgo.Interaction, respondMsg string) (*discordgo.Message, error) {
	return s.FollowupMessageCreate(i, true, &discordgo.WebhookParams{
		Content: respondMsg,
	})
}

func RespondToInteraction(s *discordgo.Session, i *discordgo.Interaction, respondMsg string) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: respondMsg,
		},
	})
}

func RespondToInteractionEphemeral(s *discordgo.Session, i *discordgo.Interaction, respondMsg string) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: respondMsg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func RequestPlayerAchievements(discordID string) ([]GameStatsResp, error) {
	xbID, ok := GetGamertagID(discordID)
	if !ok {
		return nil, errors.New("Please set your gamertag first using the `/gamertag` command")
	}

	url := "https://xbl.io/api/v2/achievements/player/" + xbID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Authorization", tokens.OpenXBL)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var objMap map[string]json.RawMessage
	err = decoder.Decode(&objMap)
	if err != nil {
		return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
	}

	var gamesStats []GameStatsResp
	err = json.Unmarshal(objMap["titles"], &gamesStats)
	if err != nil {
		return nil, errors.New("Whoops! Server responded with an error! Apologies, please try again!")
	}

	if len(gamesStats) == 0 {
		return nil, errors.New("You have either not played any games or your Xbox profile is private.")
	}

	return gamesStats, nil
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

	req.Header.Add("X-Authorization", tokens.OpenXBL)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("Whoops! Server responded with an error! Apologies, please try again!")
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var objMap map[string]json.RawMessage
	err = decoder.Decode(&objMap)
	if err != nil {
		return "", errors.New("Server responded with garbage! Not your fault. Please try again now!")
	}

	var respID []GTResp
	err = json.Unmarshal(objMap["profileUsers"], &respID)
	if err != nil {
		return "", errors.New("Hmm, that gamertag didn't work! Please make sure you typed the gamertag correctly.")
	}

	return respID[0].ID, nil
}
