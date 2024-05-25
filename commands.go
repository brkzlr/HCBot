package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func InitCommands(s *discordgo.Session) error {
	// Register each slash command to Discord
	slashCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "count",
			Description: "Show the number of users of each completion role",
		},
		{
			Name:        "gamertag",
			Description: "Set your gamertag for the /rolecheck command",
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
			Name:        "rolecheck",
			Description: "Check and receive roles according to your Halo games achievements",
		},
		{
			Name:        "riddle",
			Description: "Get a random riddle from the internet",
		},
		{
			Name:        "timestamp",
			Description: "Generate a Unix timestamped message from the bot. Input time/date must be in UTC!",
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
			Description: "Generate a Unix timestamped message from the bot. Input time is relative from your current time.",
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
	if _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, guildID, slashCommands); err != nil {
		return err
	}

	// Create the handler for each slash command
	slashCommandsHandlers["count"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("count", i.Member.User.Username)
		rolesToCheck := []string{mccRoleID, mccChinaRoleID, infiniteRoleID, modernRoleID, legacyRoleID, lasochistRoleID, mccMasterRoleID, hcRoleID, fcRoleID}

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
			"MCC China:  **%d**\n" +
			"Infinite:  **%d**\n" +
			"Modern Completionist:  **%d**\n" +
			"Legacy Completionist:  **%d**\n" +
			"Lasochist:  **%d**\n" +
			"MCC Master:  **%d**\n" +
			"Halo Completionist:  **%d**\n" +
			"Franchise Completionist:  **%d**\n"

		resultMsg := fmt.Sprintf(resultStr,
			rolesCount[mccRoleID],
			rolesCount[mccChinaRoleID],
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
		if !isTest && i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
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

	slashCommandsHandlers["rolecheck"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("rolecheck", i.Member.User.Username)
		if !isTest && i.ChannelID != botChannelID && i.ChannelID != "1026542051287892009" {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("This command is usable only in <#%s>!", botChannelID))
			return
		}
		RespondACKToInteraction(s, i.Interaction)

		if onCooldown, duration := CheckCooldown(i.Member.User.ID); onCooldown {
			RespondFollowUpToInteraction(s, i.Interaction, fmt.Sprintf("Sorry! You're on cooldown for this command. Remaining duration: %s", duration.String()))
			return
		}

		gamesStats, err := RequestPlayerAchievements(i.Member.User.ID)
		if err != nil {
			RespondFollowUpToInteraction(s, i.Interaction, err.Error())
			return
		}
		spartanGamesStats, err := CheckLegacyAssaultStrikeAchievements(i.Member.User.ID)
		if err != nil {
			RespondFollowUpToInteraction(s, i.Interaction, err.Error())
			return
		}

		if !isTest && !IsStaff(i.Member) {
			AddRoleCheckCooldown(i.Member.User.ID)
		}

		modernCompletionMap := map[string]GameStatus{
			mccTitleID:      NOT_FOUND,
			h5TitleID:       NOT_FOUND,
			hwdeTitleID:     NOT_FOUND,
			hw2TitleID:      NOT_FOUND,
			infiniteTitleID: NOT_FOUND,
		}
		legacyCompletionMap := map[string]GameStatus{
			hceaTitleID:  NOT_FOUND,
			h3TitleID:    NOT_FOUND,
			hwTitleID:    NOT_FOUND,
			odstTitleID:  NOT_FOUND,
			reachTitleID: NOT_FOUND,
			h4TitleID:    NOT_FOUND,
		}
		miscCompletionMap := map[string]GameStatus{
			h2TitleID:       NOT_FOUND,
			mccChinaTitleID: NOT_FOUND,
			h5ForgeTitleID:  NOT_FOUND,
		}

		// These 2 little shits need their own maps due to the vast amount of title IDs they have
		assaultCompletionMap := map[string]GameStatus{
			hsaTitleID:     NOT_FOUND,
			hsaXboxTitleID: NOT_FOUND,
			hsa360TitleID:  NOT_FOUND,
			hsaWPTitleID:   NOT_FOUND,
			hsaIOSTitleID:  NOT_FOUND,
		}
		strikeCompletionMap := map[string]GameStatus{
			hssTitleID:    NOT_FOUND,
			hssWPTitleID:  NOT_FOUND,
			hssIOSTitleID: NOT_FOUND,
		}
		/////////////////////////////////////////////////////////////////////////////////////////

		for _, game := range gamesStats {
			isDone := (game.Stats.CurrentGScore == game.Stats.TotalGScore) && (game.Stats.TotalGScore != 0)
			if _, exists := modernCompletionMap[game.TitleID]; exists {
				if isDone {
					modernCompletionMap[game.TitleID] = COMPLETED
				} else {
					modernCompletionMap[game.TitleID] = NOT_COMPLETED
				}
			} else if _, exists := legacyCompletionMap[game.TitleID]; exists {
				if isDone {
					legacyCompletionMap[game.TitleID] = COMPLETED
				} else {
					legacyCompletionMap[game.TitleID] = NOT_COMPLETED
				}
			} else if _, exists := miscCompletionMap[game.TitleID]; exists {
				if game.TitleID == mccChinaTitleID {
					// MCC CN has 4 unobtainable achievements that are 90G in total
					isDone = (game.Stats.CurrentGScore == (game.Stats.TotalGScore - 90))
				}
				if isDone {
					miscCompletionMap[game.TitleID] = COMPLETED
				} else {
					miscCompletionMap[game.TitleID] = NOT_COMPLETED
				}
			} else if game.TitleID == hsaXboxTitleID {
				// We'll grab X1/XSX version of Spartan Assault here but check the other versions below
				if isDone {
					assaultCompletionMap[hsaXboxTitleID] = COMPLETED
				} else {
					assaultCompletionMap[hsaXboxTitleID] = NOT_COMPLETED
				}
			}
		}
		for titleID := range assaultCompletionMap {
			if titleID != hsaXboxTitleID {
				assaultCompletionMap[titleID] = spartanGamesStats[titleID]
			}
		}
		for titleID := range strikeCompletionMap {
			strikeCompletionMap[titleID] = spartanGamesStats[titleID]
		}

		progressMsg := `Role check done! If you fulfill any of the role requirements, the role has been assigned to you. You can check role requirements in <#984078260671483945>.

You can find your current progress on the Halo games below, **but please note that the following info is not saved, only checked in the moment when you use this command**.
Legend:
- ❌: Game not found on your profile
- 🔶: Not all obtainable achievements were earned
- ✅: All obtainable achievements earned

Modern games:
- **Halo: The Master Chief Collection**: %s
- **Halo 5: Guardians**: %s
- **Halo Wars: Definitive Edition**: %s
- **Halo Wars 2**: %s
- **Halo Infinite**: %s

Legacy games:
- **Halo: Combat Evolved Anniversary**: %s
- **Halo 3**: %s
- **Halo Wars**: %s
- **Halo 3: ODST**: %s
- **Halo: Reach**: %s
- **Halo 4**: %s

Spartan Assault versions:
- **Windows**: %s
- **Xbox One/Series**: %s
- **Xbox 360**: %s
- **Windows Phone**: %s
- **iOS**: %s

Spartan Strike versions:
- **Windows**: %s
- **Windows Phone**: %s
- **iOS**: %s

Forgotten by the world:
- **Halo 2 (Vista)**: %s
- **Halo: The Master Chief Collection (China)**: %s
- **Halo 5 Forge**: %s

Note: **If you fulfill the requirements for the Modern/Halo Completionist role but finished HWDE/SA/SS on Steam, post screenshot proof in <#984079675385077820> with "Manual check: (role name)" text attached to the pictures, all in the same post.**`
		progressMsg = fmt.Sprintf(progressMsg,
			GetCompletionSymbol(modernCompletionMap[mccTitleID]),
			GetCompletionSymbol(modernCompletionMap[h5TitleID]),
			GetCompletionSymbol(modernCompletionMap[hwdeTitleID]),
			GetCompletionSymbol(modernCompletionMap[hw2TitleID]),
			GetCompletionSymbol(modernCompletionMap[infiniteTitleID]),
			GetCompletionSymbol(legacyCompletionMap[hceaTitleID]),
			GetCompletionSymbol(legacyCompletionMap[h3TitleID]),
			GetCompletionSymbol(legacyCompletionMap[hwTitleID]),
			GetCompletionSymbol(legacyCompletionMap[odstTitleID]),
			GetCompletionSymbol(legacyCompletionMap[reachTitleID]),
			GetCompletionSymbol(legacyCompletionMap[h4TitleID]),
			GetCompletionSymbol(assaultCompletionMap[hsaTitleID]),
			GetCompletionSymbol(assaultCompletionMap[hsaXboxTitleID]),
			GetCompletionSymbol(assaultCompletionMap[hsa360TitleID]),
			GetCompletionSymbol(assaultCompletionMap[hsaWPTitleID]),
			GetCompletionSymbol(assaultCompletionMap[hsaIOSTitleID]),
			GetCompletionSymbol(strikeCompletionMap[hssTitleID]),
			GetCompletionSymbol(strikeCompletionMap[hssWPTitleID]),
			GetCompletionSymbol(strikeCompletionMap[hssIOSTitleID]),
			GetCompletionSymbol(miscCompletionMap[h2TitleID]),
			GetCompletionSymbol(miscCompletionMap[mccChinaTitleID]),
			GetCompletionSymbol(miscCompletionMap[h5ForgeTitleID]),
		)
		RespondFollowUpToInteraction(s, i.Interaction, progressMsg)

		// Give non-stackable roles first if eligible
		legacyDone := true
		for _, gameStatus := range legacyCompletionMap {
			if gameStatus != COMPLETED {
				legacyDone = false
				break
			}
		}
		if legacyDone {
			s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, legacyRoleID)
		}
		if gameStatus := miscCompletionMap[mccChinaTitleID]; gameStatus == COMPLETED {
			s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, mccChinaRoleID)
		}
		///////////////////////////////////////////

		// Early exit as FC/HC holders don't need the checks below
		if HasRole(i.Member, fcRoleID) || HasRole(i.Member, hcRoleID) {
			return
		}

		// Check Assault and Strike individually as we only care about one version completion
		assaultDone := false
		for _, gameStatus := range assaultCompletionMap {
			if gameStatus == COMPLETED {
				assaultDone = true
				break
			}
		}
		strikeDone := false
		for _, gameStatus := range strikeCompletionMap {
			if gameStatus == COMPLETED {
				strikeDone = true
				break
			}
		}

		// Check role eligibility in the order of priority: HC -> Modern -> Infinite & MCC
		modernDone := true
		modernPartiallyDone := true
		for titleID, gameStatus := range modernCompletionMap {
			if gameStatus != COMPLETED {
				modernDone = false

				// We can still use modern as a base for HC if we're only missing MCC and HWDE
				if titleID != mccTitleID && titleID != hwdeTitleID {
					modernPartiallyDone = false
				}
			}
		}
		if modernDone {
			// Check SS, SA and Forge for HC eligibility
			if miscCompletionMap[h5ForgeTitleID] == COMPLETED && assaultDone && strikeDone {
				// Grant HC
				s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, mccRoleID)
				s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, infiniteRoleID)
				s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, modernRoleID)
				s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, hcRoleID)
			} else {
				// Grant Modern
				s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, mccRoleID)
				s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, infiniteRoleID)
				s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, modernRoleID)
			}
		} else if modernPartiallyDone && miscCompletionMap[h5ForgeTitleID] == COMPLETED && assaultDone && strikeDone {
			// We can still grant HC if we replace MCC with classic halos and HWDE with classic HW
			grantHC := false
			if legacyDone && miscCompletionMap[h2TitleID] == COMPLETED {
				grantHC = true
			} else {
				// If we don't have all of the alternative games completed, we have to manually check to make sure we're still eligibile for HC
				if modernCompletionMap[mccTitleID] != COMPLETED && modernCompletionMap[hwdeTitleID] == COMPLETED {
					grantHC = (legacyCompletionMap[hceaTitleID] == COMPLETED) &&
						(legacyCompletionMap[h3TitleID] == COMPLETED) &&
						(legacyCompletionMap[odstTitleID] == COMPLETED) &&
						(legacyCompletionMap[reachTitleID] == COMPLETED) &&
						(legacyCompletionMap[h4TitleID] == COMPLETED) &&
						(miscCompletionMap[h2TitleID] == COMPLETED)

				} else if modernCompletionMap[mccTitleID] == COMPLETED && modernCompletionMap[hwdeTitleID] != COMPLETED {
					grantHC = (legacyCompletionMap[hwTitleID] == COMPLETED)
				}
				// If both are not completed, then we don't have to check anymore as we need all legacy and h2v which is checked in the first if above
			}

			if grantHC {
				s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, mccRoleID)
				s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, infiniteRoleID)
				s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, modernRoleID)
				s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, hcRoleID)
			} else {
				// Grant Infinite as we have it done and couldn't give HC
				s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, infiniteRoleID)
			}
		} else {
			// Check MCC and Infinite eligibility
			if modernCompletionMap[mccTitleID] == COMPLETED {
				s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, mccRoleID)
			}
			if modernCompletionMap[infiniteTitleID] == COMPLETED {
				s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, infiniteRoleID)
			}
		}
	}

	slashCommandsHandlers["riddle"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("riddle", i.Member.User.Username)

		riddle, err := GetRiddle()
		if err != nil {
			RespondToInteraction(s, i.Interaction, "Whoops, encountered an error while trying to find a riddle. Sorry!")
			log.Println("Error while trying to obtain a riddle! ", err)
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
	return nil
}
