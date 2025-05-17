package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var isTimerActive bool

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	message, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	if err != nil {
		log.Println("Error retrieving message for reaction checking!")
		return
	}

	if m.Emoji.ID != "1117969363627159622" {
		return
	}
	if message.Author.ID != s.State.User.ID {
		return
	}
	if !strings.Contains(message.Content, "Remember to grab your") {
		return
	}

	roleID := roleRegex.FindStringSubmatch(message.Content)[1]

	err = s.GuildMemberRoleRemove(guildID, m.UserID, roleID)
	if err != nil {
		log.Printf("Error removing timed role (ID: %s) from user %s\n", roleID, m.Member.User.Username)
		return
	}
	infoLog.Printf("Successfully removed timed role (ID: %s) from user %s\n", roleID, m.Member.User.Username)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID == dropsChannelID && m.Author.ID == "1287938268423524352" && !isTimerActive {
		isTimerActive = true
		time.AfterFunc(10*time.Minute, func() {
			replyStr := fmt.Sprintf("Hey <@&%s>! Check out this ^ drop/giveaway.", dropsRoleID)
			ReplyToMsg(s, m.Message, replyStr)
			isTimerActive = false
		})
		return
	}
	if m.Author.Bot {
		return
	}
	if m.Member == nil {
		// Ignore messages not coming from a guild
		return
	}

	msg := strings.ToLower(m.Content)

	/////////////////////////////////////////////////////////////////////////////////////////////////////
	///////////////////////////////////// Check for discord invites /////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////
	if !IsStaff(m.Member) &&
		(strings.Contains(msg, "discord.gg/") || strings.Contains(msg, "discord.com/invite/")) {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		return
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////
	///////////////////////// Check for proof-of-completion channel misuse //////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////
	if m.ChannelID == proofChannelID && !IsStaff(m.Member) && !strings.Contains(msg, "manual check: modern") && !strings.Contains(msg, "manual check: halo") {
		if (strings.Contains(msg, "mcc") && !strings.Contains(msg, "master")) ||
			strings.Contains(msg, "chief collection") ||
			strings.Contains(msg, "china") || strings.Contains(msg, "cn") ||
			strings.Contains(msg, "infinite") ||
			strings.Contains(msg, "legacy") ||
			strings.Contains(msg, "modern") ||
			strings.Contains(msg, "halo completionist") ||
			strings.Contains(msg, "hc") {

			str := fmt.Sprintf("<@%s> You can obtain that role **only** by using me in <#%s>! Use /rolecheck in there to begin.", m.Author.ID, botChannelID)
			s.ChannelMessageSend(proofChannelID, str)
			s.ChannelMessageDelete(proofChannelID, m.ID)
			return
		} else {
			saidAcceptedWords := func(msg string) bool {
				return strings.Contains(msg, "franchise") ||
					strings.Contains(msg, "lasochist") ||
					strings.Contains(msg, "mcc master") ||
					strings.Contains(msg, "ice") ||
					strings.Contains(msg, "fire") ||
					strings.Contains(msg, "jacker") ||
					strings.Contains(msg, "jackal")
			}
			for _, attach := range m.Attachments {
				if attach != nil && attach.Height != 0 && !saidAcceptedWords(msg) { // Posted image but didn't say an acceptable proof role or keyword
					str := fmt.Sprintf("<@%s> Please state the exact name of the vanity role you wish to obtain in the same message as the image!", m.Author.ID)
					s.ChannelMessageSend(proofChannelID, str)
					s.ChannelMessageDelete(proofChannelID, m.ID)
					return
				}
			}
		}
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////
	//////////////////////////// Check for wrong channel reach messages /////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////
	if (m.ChannelID == reachChannelID || m.ChannelID == generalMccChannelID) && !IsStaff(m.Member) {
		if strings.Contains(msg, "skunked") || strings.Contains(msg, "invasion") || strings.Contains(msg, "headhunter") || (strings.Contains(msg, "negative") && strings.Contains(msg, "ghostrider")) {
			str := fmt.Sprintf("<@%s> That achievement is a multiplayer achievement. You should use <#984080289242427413> for multiplayer achievements and not the campaign channels!\n\n***Make sure to check the pinned posts in that channel before posting!***", m.Author.ID)
			s.ChannelMessageSend(m.ChannelID, str)
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////
	//////////////////////////// Check for missing platform in nickname /////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////
	if (m.ChannelID != "984080289242427413" && m.ChannelID != "984080200054734849" && m.ChannelID != "984080232904527872" && m.ChannelID != "984080266135994418") && !IsStaff(m.Member) {
		// Channel IDs in order: MCC Multiplayer, Halo 3 MCC, ODST MCC, MCC China

		var platform string
		if m.Member.Nick != "" {
			platform = platformRegex.FindString(strings.ToLower(m.Member.Nick))
		} else {
			platform = platformRegex.FindString(strings.ToLower(m.Author.GlobalName))
		}

		if platform == "" {
			channel, err := s.Channel(m.ChannelID)
			if err != nil {
				log.Printf("Error retrieving channel for MCC platform checking! Error: %s", err)
				return
			}

			if channel.ParentID == "984080152088698920" { // Master Chief Collection category
				currentRoles := HasRoles(m.Member, []string{mccRoleID, mccMasterRoleID, modernRoleID, hcRoleID, fcRoleID})
				for _, hasRole := range currentRoles {
					if hasRole == true {
						return
					}
				}

				str := fmt.Sprintf("<@%s> You're trying to talk in MCC channels but you're missing a platform tag in your name!\nMCC does not support cross-platform play except for Halo 3, ODST and Multiplayer modes ***so you must have a platform tag in your display name/server nickname***.\n\nFor more information and examples of platform tags, check <#1046457435277242470> ***which is mandatory reading***.", m.Author.ID)
				s.ChannelMessageSend(m.ChannelID, str)
				s.ChannelMessageDelete(m.ChannelID, m.ID)
			}
		}

	}
}
