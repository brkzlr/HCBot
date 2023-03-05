package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var commands = make(map[string]func(s *discordgo.Session, m *discordgo.Message))

func init() {
	commands["+help"] = func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("help", m.Author.Username)
		helpField := discordgo.MessageEmbedField{
			Name: "Commands Help",
			Value: `+gamertag (or +gt) "gamertag" - sets your gamertag
			+count - shows number of users with corresponding completion role
			+riddle - for those that fancy a riddle

			***Only after setting your gamertag once:***
			+mcc - checks if you're eligible for MCC role
			+infinite - checks if you're eligible for Halo Infinite role
			+legacy - checks if you're eligible for Legacy Completionist role
			+modern - checks if you're eligible for Modern Completionist role
			+hc - checks if you're eligible for Halo Completionist role`,
			Inline: true,
		}
		embed := discordgo.MessageEmbed{
			Color:  0x5662f6,
			Fields: []*discordgo.MessageEmbedField{&helpField},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	}

	commands["+count"] = func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("count", m.Author.Username)
		rolesToCheck := []string{mccRoleID, infiniteRoleID, modernRoleID, legacyRoleID, lasochistRoleID, mccMasterRoleID, hcRoleID, fcRoleID}

		rolesCount := make(map[string]int)
		for _, roleID := range rolesToCheck {
			rolesCount[roleID] = 0
		}

		guildMembers := GetAllGuildMembers(s, m.GuildID)
		for _, member := range guildMembers {
			rolesMap := HasRoles(member, rolesToCheck)
			for roleID, hasRole := range rolesMap {
				if hasRole {
					rolesCount[roleID]++
				}
			}
		}

		resultStr := "Number of users with each role:\n" +
			"MCC:  **%d**\n" +
			"Infinite:  **%d**\n" +
			"Modern Completionist:  **%d**\n" +
			"Legacy Completionist:  **%d**\n" +
			"Lasochist:  **%d**\n" +
			"MCC Master:  **%d**\n" +
			"Halo Completionist:  **%d**\n" +
			"Franchise Completionist:  **%d**\n"

		ReplyToMsg(s, m, fmt.Sprintf(resultStr,
			rolesCount[mccRoleID],
			rolesCount[infiniteRoleID],
			rolesCount[modernRoleID],
			rolesCount[legacyRoleID],
			rolesCount[lasochistRoleID],
			rolesCount[mccMasterRoleID],
			rolesCount[hcRoleID],
			rolesCount[fcRoleID]))
	}

	gtFunc := func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("gt", m.Author.Username)
		contentSpl := strings.Split(m.Content, " ")[1:]
		if len(contentSpl) == 0 {
			ReactFail(s, m)
			ReplyToMsg(s, m, "Please write your gamertag after the command.")
			return
		}
		s.MessageReactionAdd(m.ChannelID, m.ID, "⚙️")

		gTag := strings.Join(contentSpl, " ")
		gTag = strings.Trim(gTag, "\"")
		gTag = strings.Trim(gTag, "“")
		gTag = strings.Trim(gTag, "”")
		xuid, err := RequestPlayerGT(gTag)
		if err != nil {
			ReactFail(s, m)
			ReplyToMsg(s, m, err.Error())
			return
		}

		AddGamertagToDB(m.Author.ID, xuid)
		ReactSuccess(s, m)
		ReplyToMsg(s, m, fmt.Sprintf("Gamertag set to \"%s\".", gTag))
	}
	commands["+gt"] = gtFunc
	commands["+gamertag"] = gtFunc

	commands["+mcc"] = func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("mcc", m.Author.Username)
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}

		rolesMap := HasRoles(member, []string{mccRoleID, mccMasterRoleID, modernRoleID, hcRoleID, fcRoleID})

		if rolesMap[fcRoleID] {
			ReplyToMsg(s, m, "You've already finished Franchise Completionist, which requires MCC.")
			return
		}
		if rolesMap[hcRoleID] {
			ReplyToMsg(s, m, "You've already finished Halo Completionist, which replaces MCC.")
			return
		}
		if rolesMap[modernRoleID] {
			ReplyToMsg(s, m, "You've already finished Modern Completionist, which requires MCC.")
			return
		}
		if rolesMap[mccMasterRoleID] {
			ReplyToMsg(s, m, "You've already finished MCC Master, which requires more than the MCC 100% role.")
			return
		}
		if rolesMap[mccRoleID] {
			ReplyToMsg(s, m, "You've already finished MCC.")
			return
		}

		s.MessageReactionAdd(m.ChannelID, m.ID, "⚙️")
		games, err := RequestPlayerAchievements(m.Author.ID)
		if err != nil {
			ReactFail(s, m)
			ReplyToMsg(s, m, err.Error())
			return
		}

		for _, game := range games {
			if game.TitleID == "1144039928" {
				if game.Stats.CurrentGScore == game.Stats.TotalGScore {
					ReactSuccess(s, m)
					ReplyToMsg(s, m, fmt.Sprintf("Hey everyone! %s finished MCC! Congrats!", m.Author.Username))
					s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, mccRoleID)
					return
				} else {
					ReactFail(s, m)
					ReplyToMsg(s, m, "Sorry, you haven't finished MCC yet.")
					return
				}
			}
		}

		ReactFail(s, m)
		ReplyToMsg(s, m, "You haven't played MCC before.")
	}

	commands["+infinite"] = func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("infinite", m.Author.Username)
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}

		rolesMap := HasRoles(member, []string{infiniteRoleID, modernRoleID, hcRoleID, fcRoleID})

		if rolesMap[fcRoleID] {
			ReplyToMsg(s, m, "You've already finished Franchise Completionist, which requires Halo Infinite.")
			return
		}
		if rolesMap[hcRoleID] {
			ReplyToMsg(s, m, "You've already finished Halo Completionist, which requires Halo Infinite.")
			return
		}
		if rolesMap[modernRoleID] {
			ReplyToMsg(s, m, "You've already finished Modern Completionist, which requires Halo Infinite.")
			return
		}
		if rolesMap[infiniteRoleID] {
			ReplyToMsg(s, m, "You've already finished Halo Infinite.")
			return
		}

		s.MessageReactionAdd(m.ChannelID, m.ID, "⚙️")
		games, err := RequestPlayerAchievements(m.Author.ID)
		if err != nil {
			ReactFail(s, m)
			ReplyToMsg(s, m, err.Error())
			return
		}

		for _, game := range games {
			if game.Name == "Halo Infinite" {
				if game.Stats.CurrentGScore == game.Stats.TotalGScore {
					ReactSuccess(s, m)
					ReplyToMsg(s, m, fmt.Sprintf("Hey everyone! %s finished Halo Infinite! Congrats!", m.Author.Username))
					s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, infiniteRoleID)
					return
				} else {
					ReactFail(s, m)
					ReplyToMsg(s, m, "Sorry, you haven't finished Halo Infinite yet.")
					return
				}
			}
		}

		ReactFail(s, m)
		ReplyToMsg(s, m, "You haven't played Halo Infinite before.")
	}

	commands["+legacy"] = func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("legacy", m.Author.Username)
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}

		rolesMap := HasRoles(member, []string{legacyRoleID, fcRoleID})

		if rolesMap[fcRoleID] {
			ReplyToMsg(s, m, "You've already finished Franchise Completionist, which replaces Legacy Completionist.")
			return
		}
		if rolesMap[legacyRoleID] {
			ReplyToMsg(s, m, "You've already finished Legacy Completionist.")
			return
		}

		s.MessageReactionAdd(m.ChannelID, m.ID, "⚙️")
		games, err := RequestPlayerAchievements(m.Author.ID)
		if err != nil {
			ReactFail(s, m)
			ReplyToMsg(s, m, err.Error())
			return
		}

		completionMap := map[string]bool{
			"Halo: Combat Evolved Anniversary": false,
			"Halo 3":                           false,
			"Halo Wars":                        false,
			"Halo 3: ODST Campaign Edition":    false,
			"Halo: Reach":                      false,
			"Halo 4":                           false,
		}
		for _, game := range games {
			if _, exists := completionMap[game.Name]; exists {
				completionMap[game.Name] = (game.Stats.CurrentGScore == game.Stats.TotalGScore)
			}
		}

		for _, isDone := range completionMap {
			if !isDone {
				ReactFail(s, m)
				failMsg := `Here's your progress on the Legacy Completionist games:
**Halo Combat Evolved Anniversary** : %s
**Halo 3** : %s
**Halo Wars** : %s
**Halo 3 ODST** : %s
**Halo Reach** : %s
**Halo 4** : %s`
				failMsg = fmt.Sprintf(failMsg,
					GetCompletionSymbol(completionMap["Halo: Combat Evolved Anniversary"]),
					GetCompletionSymbol(completionMap["Halo 3"]),
					GetCompletionSymbol(completionMap["Halo Wars"]),
					GetCompletionSymbol(completionMap["Halo 3: ODST Campaign Edition"]),
					GetCompletionSymbol(completionMap["Halo: Reach"]),
					GetCompletionSymbol(completionMap["Halo 4"]),
				)
				ReplyToMsg(s, m, failMsg)
				return
			}
		}

		ReactSuccess(s, m)
		ReplyToMsg(s, m, fmt.Sprintf("Hey everyone! %s finished Legacy Completionist! Congrats!", m.Author.Username))
		s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, legacyRoleID)
	}

	commands["+modern"] = func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("modern", m.Author.Username)
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}

		rolesMap := HasRoles(member, []string{modernRoleID, hcRoleID, fcRoleID})

		if rolesMap[fcRoleID] {
			ReplyToMsg(s, m, "You've already finished Franchise Completionist, which replaces Modern Completionist.")
			return
		}
		if rolesMap[hcRoleID] {
			ReplyToMsg(s, m, "You've already finished Halo Completionist, which replaces Modern Completionist.")
			return
		}
		if rolesMap[modernRoleID] {
			ReplyToMsg(s, m, "You've already finished Modern Completionist.")
			return
		}

		s.MessageReactionAdd(m.ChannelID, m.ID, "⚙️")
		games, err := RequestPlayerAchievements(m.Author.ID)
		if err != nil {
			ReactFail(s, m)
			ReplyToMsg(s, m, err.Error())
			return
		}

		completionMap := map[string]bool{
			"Halo: The Master Chief Collection":  false,
			"Halo 5: Guardians":                  false,
			"Halo Wars: Definitive Edition (PC)": false,
			"Halo Wars 2":                        false,
			"Halo Infinite":                      false,
		}
		for _, game := range games {
			if completion, exists := completionMap[game.Name]; exists && !completion {
				// Some games like MCC & MCC China have the same XBL name so we don't want to replace
				// a true completion with a false completion which is why we check !completion
				completionMap[game.Name] = (game.Stats.CurrentGScore == game.Stats.TotalGScore)
			}
		}

		for _, isDone := range completionMap {
			if !isDone {
				ReactFail(s, m)
				failMsg := `Here's your progress on the Modern Completionist games:
**Halo MCC** : %s
**Halo 5 Guardians** : %s
**Halo Wars Definitive Edition** : %s
**Halo Wars 2** : %s
**Halo Infinite** : %s

Note: **If you finished everything and played any game on a non-XBL platform, please ping a staff member with screenshot proof in <#984079675385077820>.**`
				failMsg = fmt.Sprintf(failMsg,
					GetCompletionSymbol(completionMap["Halo: The Master Chief Collection"]),
					GetCompletionSymbol(completionMap["Halo 5: Guardians"]),
					GetCompletionSymbol(completionMap["Halo Wars: Definitive Edition (PC)"]),
					GetCompletionSymbol(completionMap["Halo Wars 2"]),
					GetCompletionSymbol(completionMap["Halo Infinite"]),
				)
				ReplyToMsg(s, m, failMsg)
				return
			}
		}

		ReactSuccess(s, m)
		ReplyToMsg(s, m, fmt.Sprintf("Hey everyone! %s finished Modern Completionist! Congrats!", m.Author.Username))
		s.GuildMemberRoleRemove(m.GuildID, m.Author.ID, mccRoleID)
		s.GuildMemberRoleRemove(m.GuildID, m.Author.ID, infiniteRoleID)
		s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, modernRoleID)
	}

	commands["+hc"] = func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("hc", m.Author.Username)
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}

		rolesMap := HasRoles(member, []string{hcRoleID, fcRoleID})

		if rolesMap[fcRoleID] {
			ReplyToMsg(s, m, "You've already finished Franchise Completionist, which replaces Halo Completionist.")
			return
		}
		if rolesMap[hcRoleID] {
			ReplyToMsg(s, m, "You've already finished Halo Completionist.")
			return
		}

		s.MessageReactionAdd(m.ChannelID, m.ID, "⚙️")
		games, err := RequestPlayerAchievements(m.Author.ID)
		if err != nil {
			ReactFail(s, m)
			ReplyToMsg(s, m, err.Error())
			return
		}

		completionMap := map[string]bool{
			"Halo: The Master Chief Collection":  false,
			"Halo Wars":                          false,
			"Halo Wars: Definitive Edition (PC)": false,
			"Halo Wars 2":                        false,
			"Halo: Spartan Strike":               false,
			"Halo: Spartan Assault":              false,
			"Halo Infinite":                      false,
			"Halo 5: Guardians":                  false,
			"Halo 5: Forge":                      false,
		}
		classicCompletionMap := map[string]bool{
			"Halo: Combat Evolved Anniversary": false,
			"Halo 2 (PC)":                      false,
			"Halo 3":                           false,
			"Halo 3: ODST Campaign Edition":    false,
			"Halo: Reach":                      false,
			"Halo 4":                           false,
		}

		for _, game := range games {
			if completion, exists := completionMap[game.Name]; exists && !completion {
				isBugged := game.Stats.TotalGScore == 0 // Some games like SS and SA are bugged

				// Some games like MCC & MCC China (or SS/SA in this case)
				// have the same XBL name so we don't want to replace
				// a true completion with a false completion which is why we check !completion
				completionMap[game.Name] = (game.Stats.CurrentGScore == game.Stats.TotalGScore) && !isBugged
			}

			if _, exists := classicCompletionMap[game.Name]; exists {
				classicCompletionMap[game.Name] = (game.Stats.CurrentGScore == game.Stats.TotalGScore)
			}
		}

		classicDone := true
		for _, isDone := range classicCompletionMap {
			if !isDone {
				classicDone = false
				break
			}
		}

		for gameName, isDone := range completionMap {
			if !isDone {
				if gameName == "Halo: The Master Chief Collection" && classicDone {
					continue
				}
				if gameName == "Halo Wars" || gameName == "Halo Wars: Definitive Edition (PC)" {
					if completionMap["Halo Wars"] || completionMap["Halo Wars: Definitive Edition (PC)"] {
						//Just one of these is necessary
						continue
					}
				}
				ReactFail(s, m)
				failMsg := `Here's your progress on the Halo Completionist games:
**Halo MCC** : %s
**Halo Wars Definitive Edition** : %s  or  **Halo Wars** : %s
**Halo 5 Guardians** : %s
**Halo 5 Forge** : %s
**Halo Spartan Assault** : %s
**Halo Spartan Strike** : %s
**Halo Wars 2** : %s
**Halo Infinite** : %s

Instead of **Halo MCC**, you can do:
**Halo Combat Evolved Anniversary** : %s
**Halo 2 (Vista)** : %s
**Halo 3** : %s
**Halo 3 ODST** : %s
**Halo Reach** : %s
**Halo 4** : %s

Note 1: **If your SA/SS completion is not correct, those achievements might be bugged. Ping a staff member with screenshot proof in <#984079675385077820> if it blocks you from obtaining the role**
Note 2: **If you finished everything and played any game on a non-XBL platform, please ping a staff member with screenshot proof in <#984079675385077820>.**`
				failMsg = fmt.Sprintf(failMsg,
					GetCompletionSymbol(completionMap["Halo: The Master Chief Collection"]),
					GetCompletionSymbol(completionMap["Halo Wars: Definitive Edition (PC)"]),
					GetCompletionSymbol(completionMap["Halo Wars"]),
					GetCompletionSymbol(completionMap["Halo 5: Guardians"]),
					GetCompletionSymbol(completionMap["Halo 5: Forge"]),
					GetCompletionSymbol(completionMap["Halo: Spartan Assault"]),
					GetCompletionSymbol(completionMap["Halo: Spartan Strike"]),
					GetCompletionSymbol(completionMap["Halo Wars 2"]),
					GetCompletionSymbol(completionMap["Halo Infinite"]),
					GetCompletionSymbol(classicCompletionMap["Halo: Combat Evolved Anniversary"]),
					GetCompletionSymbol(classicCompletionMap["Halo 2 (PC)"]),
					GetCompletionSymbol(classicCompletionMap["Halo 3"]),
					GetCompletionSymbol(classicCompletionMap["Halo 3: ODST Campaign Edition"]),
					GetCompletionSymbol(classicCompletionMap["Halo: Reach"]),
					GetCompletionSymbol(classicCompletionMap["Halo 4"]),
				)
				ReplyToMsg(s, m, failMsg)
				return
			}
		}

		ReactSuccess(s, m)
		ReplyToMsg(s, m, fmt.Sprintf("Hey everyone! %s finished Halo Completionist! Congrats!", m.Author.Username))
		s.GuildMemberRoleRemove(m.GuildID, m.Author.ID, mccRoleID)
		s.GuildMemberRoleRemove(m.GuildID, m.Author.ID, infiniteRoleID)
		s.GuildMemberRoleRemove(m.GuildID, m.Author.ID, modernRoleID)
		s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, hcRoleID)
	}
	commands["+riddle"] = func(s *discordgo.Session, m *discordgo.Message) {
		LogCommand("riddle", m.Author.Username)

		riddle, err := GetRiddle()
		if err != nil {
			ReplyToMsg(s, m, "Whoops, encountered an error while trying to find a riddle. Sorry!")
			fmt.Println(err)
			return
		}

		msgToPrint := riddle.Question + "\n\nAnswer will be revealed in one minute."
		riddleMsg, _ := ReplyToMsg(s, m, msgToPrint)

		riddleTimer := time.NewTimer(1 * time.Minute)
		<-riddleTimer.C
		ReplyToMsg(s, riddleMsg, riddle.Answer)
	}

}
