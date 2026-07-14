package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func AddGamertagToDB(discordID, xblID string) {
	_, err := database.Exec(
		`INSERT INTO users (discordID, xuid) VALUES (?,?)
		ON CONFLICT(discordID) DO UPDATE SET xuid=?;`, discordID, xblID, xblID,
	)
	if err != nil {
		log.Printf("Failed to save gamertag for user %s: %s", discordID, err)
	}
}

func AddRoleCheckCooldown(discordID string) {
	unixTimeCD := time.Now().Add(time.Hour * 1).Unix()
	_, err := database.Exec(
		`INSERT INTO moderation (discordID, command_cooldown) VALUES (?,?)
		ON CONFLICT(discordID) DO UPDATE SET command_cooldown=?;`, discordID, unixTimeCD, unixTimeCD,
	)
	if err != nil {
		log.Printf("Failed to set rolecheck cooldown for user %s: %s", discordID, err)
	}
}

func AppendRoleName(rolesString *string, roleName string) {
	if rolesString == nil {
		return
	}
	if *rolesString == "" {
		*rolesString = roleName
	} else {
		*rolesString += fmt.Sprintf(", %s", roleName)
	}
}

func CheckCooldown(discordID string) (bool, time.Duration) {
	row := database.QueryRow(
		`SELECT command_cooldown FROM moderation WHERE discordID=?`, discordID,
	)

	var rowValue int64
	if err := row.Scan(&rowValue); err != nil {
		return false, 0
	}
	expirationTime := time.Unix(rowValue, 0)
	return expirationTime.After(time.Now()), expirationTime.Sub(time.Now())
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
		AssaultIOSAchievCount = 20 // 5 unobtainable achievements in iOS, though I developed a patch to enable the missing 5, so we'll do an additional check below.
		Assault360AchievCount = 28
		StrikeAchievCount     = 20 // Spartan Strike has 20 achievements on all 3 platforms
	)

	// This function should never be called before RequestPlayerAchievements so we'll skip the checks
	xbID, _ := GetGamertagID(discordID)

	for titleID := range gamesToCheck {
		url := "https://api.xbl.io/v2/achievements/x360/" + xbID + "/title/" + titleID
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("X-Authorization", tokens.OpenXBL)
		req.Header.Add("Accept", "application/json")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error in x360 achievements GET request: ", err)
			return nil, errors.New("Error trying to contact the server! Please try again later.")
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			log.Println("Received http error status code in x360 achievements GET response, status code: ", resp.StatusCode)
			return nil, errors.New("OpenXBL responded with an error status code. Please try again later.")
		}

		var achievsResp X360AchievsResp
		err = json.NewDecoder(resp.Body).Decode(&achievsResp)
		resp.Body.Close()
		if err != nil {
			log.Println("Error decoding x360 achievements GET response: ", err)
			return nil, errors.New("OpenXBL sent an invalid response. Please try again later.")
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

		switch achievsResp.Content.PagingInfo.TotalRecords {
		case achievCountToCheck:
			gamesToCheck[titleID] = COMPLETED
		case 0:
			gamesToCheck[titleID] = NOT_FOUND
		default:
			if titleID == hsaIOSTitleID && achievsResp.Content.PagingInfo.TotalRecords == 25 {
				// With my iOS SA patcher fix, people can now get all 25 achievements instead of 20.
				// More info at https://github.com/brkzlr/SASSFix
				gamesToCheck[titleID] = COMPLETED
			} else {
				gamesToCheck[titleID] = NOT_COMPLETED
			}
		}
	}

	return gamesToCheck, nil
}

func CheckTimedAchievs(session *discordgo.Session) {
	// Invoked once per day at ~9 AM UTC by the scheduler in main.
	today := time.Now().UTC()

	baseText := "Remember to grab your <@&%d> achievement today! %s\n\n***If this message helped you get the achievement, make sure to react with <:pepeok:1117969363627159622> so I can remove the role from you!***"

	if timedRole, exists := timedAchievRoles[today.Day()]; exists {
		specificText := fmt.Sprintf("Simply start up a mission or load into a multiplayer game in %s", timedRole.Game)
		session.ChannelMessageSend(generalMccChannelID, fmt.Sprintf(baseText, timedRole.ID, specificText))
	}
	for _, date := range destinationVacationDates {
		if today.Day() == date.Day && today.Month() == date.Month {
			session.ChannelMessageSend(generalMccChannelID, fmt.Sprintf(baseText, 990602317575368724, "Simply load up a Custom Game on Halo 2 Classic Zanzibar, go to the beach and look at the sign next to the water!"))
			break
		}
	}
	for _, date := range elderSignsDates {
		if today.Day() == date.Day && today.Month() == date.Month {
			session.ChannelMessageSend(generalMccChannelID, fmt.Sprintf(baseText, 990602348659363850, "Simply load up a Custom Game on Halo 3 Valhalla and look at the Sigil on the wall. Remember you need to have looked at 2 different ones for it to unlock!"))
			break
		}
	}
}

func GetAsPtr[T any](v T) *T {
	return &v
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
	row := database.QueryRow(
		`SELECT xuid FROM users WHERE discordID=?`, discordID,
	)

	err := row.Scan(&xuid)
	exists = (err == nil)
	return
}

func HasAnySpecifiedRoles(member *discordgo.Member, rolesID []string) bool {
	for _, id := range member.Roles {
		if slices.Contains(rolesID, id) {
			return true
		}
	}

	return false
}

func HasRole(member *discordgo.Member, roleID string) bool {
	return slices.Contains(member.Roles, roleID)
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
	// Sentinel (Admin) / Oracle (Mod) / Guardian (Admin) / Founder roles
	staffRoles := []string{"984083352367800410", "984081125657964664", "984080972108668959", "1075504782023852102"}
	return HasAnySpecifiedRoles(member, staffRoles)
}

func LogCommand(cmdName, author string) {
	infoLog.Println(cmdName + " command used by " + author)
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

func RespondACKToInteractionEphemeral(s *discordgo.Session, i *discordgo.Interaction) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func RespondFollowUpToInteraction(s *discordgo.Session, i *discordgo.Interaction, respondMsg string) (*discordgo.Message, error) {
	return s.FollowupMessageCreate(i, true, &discordgo.WebhookParams{
		Content:         respondMsg,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	})
}

func RespondFollowUpToInteractionEphemeral(s *discordgo.Session, i *discordgo.Interaction, respondMsg string) (*discordgo.Message, error) {
	return s.FollowupMessageCreate(i, true, &discordgo.WebhookParams{
		Content:         respondMsg,
		Flags:           discordgo.MessageFlagsEphemeral,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	})
}

func RespondToInteraction(s *discordgo.Session, i *discordgo.Interaction, respondMsg string) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         respondMsg,
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}

func RespondToInteractionEphemeral(s *discordgo.Session, i *discordgo.Interaction, respondMsg string) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         respondMsg,
			Flags:           discordgo.MessageFlagsEphemeral,
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}

func RequestPlayerAchievements(discordID string) (AchievementsResp, error) {
	xbID, ok := GetGamertagID(discordID)
	if !ok {
		return AchievementsResp{}, errors.New("Your gamertag is missing from the database! Please set your gamertag first using the `/gamertag` command.")
	}

	url := "https://api.xbl.io/v2/achievements/player/" + xbID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error in creating achievements GET request: ", err)
		return AchievementsResp{}, errors.New("Whoops! Sorry, internal error. Please try again.")
	}

	req.Header.Add("X-Authorization", tokens.OpenXBL)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in achievements GET request: ", err)
		return AchievementsResp{}, errors.New("Error trying to contact the server! Please try again later.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Received http error status code in achievements GET response, status code: ", resp.StatusCode)
		return AchievementsResp{}, errors.New("OpenXBL responded with an error status code. Please try again later.")
	}

	var achievsResp AchievementsResp
	if err := json.NewDecoder(resp.Body).Decode(&achievsResp); err != nil {
		log.Println("Error decoding achievements GET response: ", err)
		return AchievementsResp{}, errors.New("OpenXBL sent an invalid response. Please try again later.")
	}

	if len(achievsResp.Content.Titles) == 0 {
		return AchievementsResp{}, errors.New("You have either not played any games or your Xbox profile is private.")
	}

	return achievsResp, nil
}

func RequestPlayerGT(gamerTag string) (string, error) {
	// Gamertags with a suffix should not include the hashtag
	urlTag := strings.ReplaceAll(gamerTag, "#", "")
	urlTag = strings.ReplaceAll(urlTag, " ", "%20")
	url := "https://api.xbl.io/v2/friends/search?gt=" + urlTag

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error in creating gamertag GET request: ", err)
		return "", errors.New("Whoops! Sorry, internal error. Please try again.")
	}

	req.Header.Add("X-Authorization", tokens.OpenXBL)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in gamertag GET request: ", err)
		return "", errors.New("Error trying to contact the server! Please try again later.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Received http error status code in gamertag GET response, status code: ", resp.StatusCode)
		return "", errors.New("OpenXBL responded with an error status code. Please try again later.")
	}

	var jsonResp GTResp
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		log.Println("Error decoding gamertag GET response: ", err)
		return "", errors.New("OpenXBL sent an invalid response. Please try again later.")
	}

	if jsonResp.StatusCode == 404 {
		str := fmt.Sprintf("I couldn't find any valid \"**%s**\" gamertag! Please make sure you typed the gamertag correctly.", gamerTag)
		return "", errors.New(str)
	}
	if len(jsonResp.Content.ProfileUsers) == 0 {
		log.Println("Gamertag GET response contains no profileUsers despite no 404 code")
		return "", errors.New("OpenXBL sent an invalid response. Please try again later.")
	}

	return jsonResp.Content.ProfileUsers[0].ID, nil
}
