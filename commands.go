package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func InitCommands(s *discordgo.Session) error {
	// Register each slash command to Discord
	appCommands = []*discordgo.ApplicationCommand{
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
			Name:                     "bulkdelete",
			Description:              "Mass delete messages from the current channel using filtering options",
			DefaultMemberPermissions: GetAsPtr[int64](discordgo.PermissionManageGuildExpressions),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "limit",
					Description: "Amount of messages to filter through (default 50)",
					MinValue:    GetAsPtr(0.0),
					MaxValue:    100,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user_id",
					Description: "If provided, only delete messages for this user",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "before_message_id",
					Description: "If provided, all messages deleted will be before given message ID",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "after_message_id",
					Description: "If provided, all messages deleted will be after given message ID",
				},
			},
		},
		{
			Name:                     "muteinchannel",
			Description:              "Remove a user's ability from talking in the current channel",
			DefaultMemberPermissions: GetAsPtr[int64](discordgo.PermissionManageGuildExpressions),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user_id",
					Description: "The user ID that will lose the ability to send messages in this channel",
					Required:    true,
				},
			},
		},
		{
			Name:                     "unmuteinchannel",
			Description:              "Restore a user's ability from talking in the current channel",
			DefaultMemberPermissions: GetAsPtr[int64](discordgo.PermissionManageGuildExpressions),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user_id",
					Description: "The user ID that will regain the ability to send messages in this channel",
					Required:    true,
				},
			},
		},
		{
			Name:                     "warnwrongchannel",
			Description:              "Warn a user for posting in the wrong channel",
			DefaultMemberPermissions: GetAsPtr[int64](discordgo.PermissionManageGuildExpressions),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message_id",
					Description: "The message that should be warned",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channel_id",
					Description: "The ID of the correct channel where the message should've been posted",
					Required:    true,
				},
			},
		},
		{
			Name:                     "removereactions",
			DefaultMemberPermissions: GetAsPtr[int64](discordgo.PermissionManageGuildExpressions),
			Type:                     discordgo.MessageApplicationCommand,
		},
	}
	if _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, guildID, appCommands); err != nil {
		return err
	}

	// Create the handler for each slash command
	appCommandsHandlers["count"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	appCommandsHandlers["gamertag"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	appCommandsHandlers["rolecheck"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

		eligibleRoles := ""
		currentRoles := HasRoles(i.Member, []string{mccRoleID, mccMasterRoleID, mccChinaRoleID, infiniteRoleID, modernRoleID, legacyRoleID, hcRoleID, fcRoleID})

		// Give non-stackable roles first if eligible
		legacyDone := true
		for _, gameStatus := range legacyCompletionMap {
			if gameStatus != COMPLETED {
				legacyDone = false
				break
			}
		}
		if legacyDone && !currentRoles[legacyRoleID] {
			s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, legacyRoleID)
			AppendRoleName(&eligibleRoles, "**Legacy Completionist**")
		}
		if gameStatus := miscCompletionMap[mccChinaTitleID]; gameStatus == COMPLETED && !currentRoles[mccChinaRoleID] {
			s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, mccChinaRoleID)
			AppendRoleName(&eligibleRoles, "**MCC CN \"100%\"**")
		}
		///////////////////////////////////////////

		// FC/HC holders don't need the checks below
		if !currentRoles[fcRoleID] && !currentRoles[hcRoleID] {
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
					if !currentRoles[hcRoleID] {
						// Grant HC
						s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, mccRoleID)
						s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, infiniteRoleID)
						s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, modernRoleID)
						s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, hcRoleID)
						AppendRoleName(&eligibleRoles, "**Halo Completionist**")
					}
				} else if !currentRoles[modernRoleID] {
					// Grant Modern
					s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, mccRoleID)
					s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, infiniteRoleID)
					s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, modernRoleID)
					AppendRoleName(&eligibleRoles, "**Modern Completionist**")
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
					if !currentRoles[hcRoleID] {
						s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, mccRoleID)
						s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, infiniteRoleID)
						s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, modernRoleID)
						s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, hcRoleID)
						AppendRoleName(&eligibleRoles, "**Halo Completionist**")
					}
				} else if !currentRoles[infiniteRoleID] {
					// Grant Infinite as we have it done and couldn't give HC
					s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, infiniteRoleID)
					AppendRoleName(&eligibleRoles, "**Infinite 100%**")
				}
			} else {
				// Check MCC and Infinite eligibility
				if modernCompletionMap[mccTitleID] == COMPLETED && !currentRoles[mccRoleID] {
					s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, mccRoleID)
					AppendRoleName(&eligibleRoles, "**MCC 100%**")
				}
				if modernCompletionMap[infiniteTitleID] == COMPLETED && !currentRoles[infiniteRoleID] {
					s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, infiniteRoleID)
					AppendRoleName(&eligibleRoles, "**Infinite 100%**")
				}
			}
		}

		if eligibleRoles == "" {
			warnMsg := "You're not eligible for any new role! **Please only use this command once you fulfill the requirements for a new role**. You can check role requirements in <#984078260671483945>."
			RespondFollowUpToInteraction(s, i.Interaction, warnMsg)
		} else {
			progressMsg := `Role check done! You have been granted the following role(s): %s

You can also check other role requirements in <#984078260671483945>. Here's your current progress on the Halo games:
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
				eligibleRoles,
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
		}
	}

	appCommandsHandlers["bulkdelete"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("bulkdelete", i.Member.User.Username)

		limit := 0
		userID := ""
		beforeID := ""
		afterID := ""

		for _, option := range i.ApplicationCommandData().Options {
			if option == nil {
				continue
			}

			switch option.Name {
			case "limit":
				limit = int(option.IntValue())
			case "user_id":
				userID = option.StringValue()
			case "before_message_id":
				beforeID = option.StringValue()
			case "after_message_id":
				afterID = option.StringValue()
			}
		}
		RespondACKPing(s, i.Interaction)

		messages, err := s.ChannelMessages(i.ChannelID, limit, beforeID, afterID, "")
		if err != nil {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("Error while grabbing channel messages! (Error: %s)", err))
			return
		}

		if limit == 0 {
			// Discord returns 50 messages by default if limit is not supplied.
			limit = 50
		}

		messagesToDelete := make([]string, 0, limit)
		for _, message := range messages {
			if message == nil {
				continue
			}
			if userID != "" && userID != message.Author.ID {
				continue
			}
			messagesToDelete = append(messagesToDelete, message.ID)
		}

		err = s.ChannelMessagesBulkDelete(i.ChannelID, messagesToDelete)
		if err != nil {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("Error while deleting channel messages! (Error: %s)", err))
			return
		}
		RespondToInteractionEphemeral(s, i.Interaction, "Successfully deleted messages!")
	}

	appCommandsHandlers["muteinchannel"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("muteinchannel", i.Member.User.Username)

		userID := i.ApplicationCommandData().Options[0].StringValue()
		RespondACKPing(s, i.Interaction)

		err := s.ChannelPermissionSet(i.ChannelID, userID, discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionSendMessages)
		if err != nil {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("Error while changing channel permission overwrite! (Error: %s)", err))
			return
		}
		RespondToInteractionEphemeral(s, i.Interaction, "Successfully removed user's ability to message in this channel!")
	}

	appCommandsHandlers["unmuteinchannel"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("unmuteinchannel", i.Member.User.Username)

		userID := i.ApplicationCommandData().Options[0].StringValue()
		RespondACKPing(s, i.Interaction)

		err := s.ChannelPermissionDelete(i.ChannelID, userID)
		if err != nil {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("Error while deleting channel permission overwrite! (Error: %s)", err))
			return
		}
		RespondToInteractionEphemeral(s, i.Interaction, "Successfully restored user's ability to message in this channel!")
	}

	appCommandsHandlers["warnwrongchannel"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("warnwrongchannel", i.Member.User.Username)

		messageID := ""
		correctChannelID := ""

		for _, option := range i.ApplicationCommandData().Options {
			if option == nil {
				continue
			}

			switch option.Name {
			case "message_id":
				messageID = option.StringValue()
			case "channel_id":
				correctChannelID = option.StringValue()
			}
		}
		RespondACKPing(s, i.Interaction)

		message, err := s.ChannelMessage(i.ChannelID, messageID)
		if err != nil {
			RespondToInteractionEphemeral(s, i.Interaction, fmt.Sprintf("Error while retrieving message! (Error: %s)", err))
			return
		}

		replyStr := fmt.Sprintf("<@%s> The topic of your message is not fit for this channel!\nIt seems you might've not read <#1046457435277242470> fully, ***which is mandatory reading***.\n\nAs a tip, you should use <#%s> for this topic, but please be more mindful of the channel you're typing in next time.", message.Author.ID, correctChannelID)
		ReplyToMsg(s, message, replyStr)
		s.ChannelMessageDelete(i.ChannelID, messageID)
		RespondToInteractionEphemeral(s, i.Interaction, "Successfully warned the user!")
	}

	appCommandsHandlers["removereactions"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		LogCommand("removereactions", i.Member.User.Username)

		message := i.ApplicationCommandData().Resolved.Messages[i.ApplicationCommandData().TargetID]
		if message == nil {
			return
		}
		RespondACKPing(s, i.Interaction)

		s.MessageReactionsRemoveAll(message.ChannelID, message.ID)
		RespondToInteractionEphemeral(s, i.Interaction, "Successfully removed all reactions from the message!")
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, exists := appCommandsHandlers[i.ApplicationCommandData().Name]; exists {
			handler(s, i)
		}
	})
	return nil
}
