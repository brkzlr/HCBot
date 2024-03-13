package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func InitCommands(s *discordgo.Session) {
	// Register each slash command to Discord
	slashCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "count",
			Description: "Show the number of users of each completion role",
		},
		{
			Name:        "gamertag",
			Description: "Set your gamertag for subsequent completion role commands",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "gamertag",
					Description: "Your gamertag",
					Required:    true,
				},
			},
		},
		{
			Name:        "mcc",
			Description: "Check if you're eligible for the MCC 100% completion role",
		},
		{
			Name:        "mcc-cn",
			Description: "Check if you're eligible for the MCC CN \"100%\" completion role",
		},
		{
			Name:        "infinite",
			Description: "Check if you're eligible for the Infinite 100% completion role",
		},
		{
			Name:        "legacy",
			Description: "Check if you're eligible for the Legacy Completionist role",
		},
		{
			Name:        "modern",
			Description: "Check if you're eligible for the Modern Completionist role",
		},
		{
			Name:        "hc",
			Description: "Check if you're eligible for the Halo Completionist role",
		},
		{
			Name:        "riddle",
			Description: "Get a random riddle from the internet",
		},
		{
			Name:        "timestamp",
			Description: "Generates a Unix timestamped message for you. Input time/date must be in UTC!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "hour",
					Description: "Specify the hour in UTC.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "minute",
					Description: "Specify the minute in UTC.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "day",
					Description: "Day of the date. If unspecified, it will be the current day at UTC time.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "month",
					Description: "Month of the date. If unspecified, it will be the current month at UTC time.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "year",
					Description: "Year of the date. If unspecified, it will be the current year at UTC time.",
					Required:    false,
				},
			},
		},
		{
			Name:        "timestamp-relative",
			Description: "Generates a Unix timestamped message for you. Input time is relative from your current time.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "days",
					Description: "How many days from this moment.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "hours",
					Description: "How many hours from this moment.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "minutes",
					Description: "How many minutes from this moment.",
					Required:    true,
				},
			},
		},
	}
	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, hcGuildID, slashCommands)
	if err != nil {
		fmt.Println(err)
	}

	// Create the handler for each slash command
	slashCommandsHandlers["count"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("count", i.Member.User.Username)
		rolesToCheck := []string{mccRoleID, infiniteRoleID, modernRoleID, legacyRoleID, lasochistRoleID, mccMasterRoleID, hcRoleID, fcRoleID}

		rolesCount := make(map[string]int)
		for _, roleID := range rolesToCheck {
			rolesCount[roleID] = 0
		}

		guildMembers := GetAllGuildMembers(s, i.GuildID)
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

		resultMsg := fmt.Sprintf(resultStr,
			rolesCount[mccRoleID],
			rolesCount[infiniteRoleID],
			rolesCount[modernRoleID],
			rolesCount[legacyRoleID],
			rolesCount[lasochistRoleID],
			rolesCount[mccMasterRoleID],
			rolesCount[hcRoleID],
			rolesCount[fcRoleID])

		RespondToInteraction(s, i.Interaction, resultMsg)
	}

	slashCommandsHandlers["gamertag"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("gamertag", i.Member.User.Username)
		if i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("This command is usable only in <#%s>!", botChannelID))
			return
		}

		gTag := i.ApplicationCommandData().Options[0].StringValue()
		xuid, err := RequestPlayerGT(gTag)
		if err != nil {
			RespondToInteraction(s, i.Interaction, err.Error())
			return
		}

		AddGamertagToDB(i.Member.User.ID, xuid)
		RespondToInteraction(s, i.Interaction, fmt.Sprintf("Gamertag set to \"%s\".", gTag))
	}

	slashCommandsHandlers["mcc"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("mcc", i.Member.User.Username)
		if i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("This command is usable only in <#%s>!", botChannelID))
			return
		}

		rolesMap := HasRoles(i.Member, []string{mccRoleID, mccMasterRoleID, modernRoleID, hcRoleID, fcRoleID})
		if rolesMap[fcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Franchise Completionist, which requires MCC.")
			return
		}
		if rolesMap[hcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Halo Completionist, which replaces MCC.")
			return
		}
		if rolesMap[modernRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Modern Completionist, which requires MCC.")
			return
		}
		if rolesMap[mccMasterRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished MCC Master, which requires more than the MCC 100% role.")
			return
		}
		if rolesMap[mccRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished MCC.")
			return
		}
		RespondACKToInteraction(s, i.Interaction)

		games, err := RequestPlayerAchievements(i.Member.User.ID)
		if err != nil {
			RespondFollowUpToInteraction(s, i.Interaction, err.Error())
			return
		}

		for _, game := range games {
			if game.TitleID == "1144039928" {
				if game.Stats.CurrentGScore == game.Stats.TotalGScore {
					RespondFollowUpToInteraction(s, i.Interaction, fmt.Sprintf("Hey everyone! %s finished MCC! Congrats!", i.Member.DisplayName()))
					s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, mccRoleID)
					return
				} else {
					RespondFollowUpToInteraction(s, i.Interaction, "Sorry, you haven't finished MCC yet.")
					return
				}
			}
		}

		RespondFollowUpToInteraction(s, i.Interaction, "You haven't played MCC before.")
	}

	slashCommandsHandlers["mcc-cn"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("mcc-cn", i.Member.User.Username)
		if i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("This command is usable only in <#%s>!", botChannelID))
			return
		}

		if HasRole(i.Member, mccChinaRoleID) {
			RespondToInteraction(s, i.Interaction, "You've already \"finished\" MCC China.")
			return
		}
		RespondACKToInteraction(s, i.Interaction)

		games, err := RequestPlayerAchievements(i.Member.User.ID)
		if err != nil {
			RespondFollowUpToInteraction(s, i.Interaction, err.Error())
			return
		}

		for _, game := range games {
			if game.TitleID == "812065290" {
				if game.Stats.CurrentGScore == (game.Stats.TotalGScore - 90) { // MCC CN has 4 unobtainable achievements that are 90G in total
					RespondFollowUpToInteraction(s, i.Interaction, fmt.Sprintf("Hey everyone! %s \"finished\" MCC China! Congrats!", i.Member.DisplayName()))
					s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, mccChinaRoleID)
					return
				} else {
					RespondFollowUpToInteraction(s, i.Interaction, "Sorry, you haven't \"finished\" MCC China yet.")
					return
				}
			}
		}

		RespondFollowUpToInteraction(s, i.Interaction, "You haven't played MCC China before.")
	}

	slashCommandsHandlers["infinite"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("infinite", i.Member.User.Username)
		if i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("This command is usable only in <#%s>!", botChannelID))
			return
		}

		rolesMap := HasRoles(i.Member, []string{infiniteRoleID, modernRoleID, hcRoleID, fcRoleID})
		if rolesMap[fcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Franchise Completionist, which requires Halo Infinite.")
			return
		}
		if rolesMap[hcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Halo Completionist, which requires Halo Infinite.")
			return
		}
		if rolesMap[modernRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Modern Completionist, which requires Halo Infinite.")
			return
		}
		if rolesMap[infiniteRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Halo Infinite.")
			return
		}
		RespondACKToInteraction(s, i.Interaction)

		games, err := RequestPlayerAchievements(i.Member.User.ID)
		if err != nil {
			RespondFollowUpToInteraction(s, i.Interaction, err.Error())
			return
		}

		for _, game := range games {
			if game.Name == "Halo Infinite" {
				if game.Stats.CurrentGScore == game.Stats.TotalGScore {
					RespondFollowUpToInteraction(s, i.Interaction, fmt.Sprintf("Hey everyone! %s finished Halo Infinite! Congrats!", i.Member.DisplayName()))
					s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, infiniteRoleID)
					return
				} else {
					RespondFollowUpToInteraction(s, i.Interaction, "Sorry, you haven't finished Halo Infinite yet.")
					return
				}
			}
		}

		RespondFollowUpToInteraction(s, i.Interaction, "You haven't played Halo Infinite before.")
	}

	slashCommandsHandlers["legacy"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("legacy", i.Member.User.Username)
		if i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("This command is usable only in <#%s>!", botChannelID))
			return
		}

		rolesMap := HasRoles(i.Member, []string{legacyRoleID, fcRoleID})
		if rolesMap[fcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Franchise Completionist, which replaces Legacy Completionist.")
			return
		}
		if rolesMap[legacyRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Legacy Completionist.")
			return
		}
		RespondACKToInteraction(s, i.Interaction)

		games, err := RequestPlayerAchievements(i.Member.User.ID)
		if err != nil {
			RespondFollowUpToInteraction(s, i.Interaction, err.Error())
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
				failMsg := `**You're not eligible for the role**. Retry the command once you have all of the achievements in the listed games.

You can find your current progress on the Legacy Completionist games below, **but please note that the following info is not saved, only checked in the moment**:
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
				RespondFollowUpToInteraction(s, i.Interaction, failMsg)
				return
			}
		}

		RespondFollowUpToInteraction(s, i.Interaction, fmt.Sprintf("Hey everyone! %s finished Legacy Completionist! Congrats!", i.Member.DisplayName()))
		s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, legacyRoleID)
	}

	slashCommandsHandlers["modern"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("modern", i.Member.User.Username)
		if i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("This command is usable only in <#%s>!", botChannelID))
			return
		}

		rolesMap := HasRoles(i.Member, []string{modernRoleID, hcRoleID, fcRoleID})
		if rolesMap[fcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Franchise Completionist, which replaces Modern Completionist.")
			return
		}
		if rolesMap[hcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Halo Completionist, which replaces Modern Completionist.")
			return
		}
		if rolesMap[modernRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Modern Completionist.")
			return
		}
		RespondACKToInteraction(s, i.Interaction)

		games, err := RequestPlayerAchievements(i.Member.User.ID)
		if err != nil {
			RespondFollowUpToInteraction(s, i.Interaction, err.Error())
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
				failMsg := `**You're not eligible for the role**. Retry the command once you have all of the achievements in the listed games.

You can find your current progress on the Modern Completionist games below, **but please note that the following info is not saved, only checked in the moment**:
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
				RespondFollowUpToInteraction(s, i.Interaction, failMsg)
				return
			}
		}

		RespondFollowUpToInteraction(s, i.Interaction, fmt.Sprintf("Hey everyone! %s finished Modern Completionist! Congrats!", i.Member.DisplayName()))
		s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, mccRoleID)
		s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, infiniteRoleID)
		s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, modernRoleID)
	}

	slashCommandsHandlers["hc"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("hc", i.Member.User.Username)
		if i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("This command is usable only in <#%s>!", botChannelID))
			return
		}

		rolesMap := HasRoles(i.Member, []string{hcRoleID, fcRoleID})
		if rolesMap[fcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Franchise Completionist, which replaces Halo Completionist.")
			return
		}
		if rolesMap[hcRoleID] {
			RespondToInteraction(s, i.Interaction, "You've already finished Halo Completionist.")
			return
		}
		RespondACKToInteraction(s, i.Interaction)

		games, err := RequestPlayerAchievements(i.Member.User.ID)
		if err != nil {
			RespondFollowUpToInteraction(s, i.Interaction, err.Error())
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
				failMsg := `**You're not eligible for the role**. Retry the command once you have all of the achievements in the listed games.

You can find your current progress on the Halo Completionist games below, **but please note that the following info is not saved, only checked in the moment**:
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
				RespondFollowUpToInteraction(s, i.Interaction, failMsg)
				return
			}
		}

		RespondFollowUpToInteraction(s, i.Interaction, fmt.Sprintf("Hey everyone! %s finished Halo Completionist! Congrats!", i.Member.DisplayName()))
		s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, mccRoleID)
		s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, infiniteRoleID)
		s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, modernRoleID)
		s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, hcRoleID)
	}

	slashCommandsHandlers["riddle"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("riddle", i.Member.User.Username)

		riddle, err := GetRiddle()
		if err != nil {
			RespondToInteraction(s, i.Interaction, "Whoops, encountered an error while trying to find a riddle. Sorry!")
			fmt.Println(err)
			return
		}

		RespondToInteraction(s, i.Interaction, riddle.Question+"\n\nAnswer will be revealed in one minute.")

		time.Sleep(1 * time.Minute)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: riddle.Answer,
		})
	}

	slashCommandsHandlers["timestamp"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("timestamp", i.Member.User.Username)

		hour := int(i.ApplicationCommandData().Options[0].IntValue())
		minute := int(i.ApplicationCommandData().Options[1].IntValue())

		day := time.Now().UTC().Day()
		month := time.Now().UTC().Month()
		year := time.Now().UTC().Year()
		for index, option := range i.ApplicationCommandData().Options {
			if index < 2 {
				continue
			}
			switch option.Name {
			case "day":
				day = int(option.IntValue())
			case "month":
				month = time.Month(option.IntValue())
			case "year":
				year = int(option.IntValue())
			}
		}

		unixTimestamp := time.Date(year, month, day, hour, minute, 0, 0, time.UTC).Unix()
		RespondToInteraction(s, i.Interaction, fmt.Sprintf("%s's specified time is <t:%d> which is <t:%d:R>.", i.Member.DisplayName(), unixTimestamp, unixTimestamp))
	}

	slashCommandsHandlers["timestamp-relative"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("relative timestamp", i.Member.User.Username)

		days := int(i.ApplicationCommandData().Options[0].IntValue())
		hours := i.ApplicationCommandData().Options[1].IntValue()
		minutes := i.ApplicationCommandData().Options[2].IntValue()

		currentTime := time.Now().UTC()
		hoursDuration, _ := time.ParseDuration(fmt.Sprintf("%dh%dm", hours, minutes))
		futureTime := currentTime.Add(hoursDuration).AddDate(0, 0, days)

		unixTimestamp := futureTime.Unix()
		RespondToInteraction(s, i.Interaction, fmt.Sprintf("%s's specified time is <t:%d> which is <t:%d:R>.", i.Member.DisplayName(), unixTimestamp, unixTimestamp))
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, exists := slashCommandsHandlers[i.ApplicationCommandData().Name]; exists {
			handler(s, i)
		}
	})
}
