package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var isTimerActive bool

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.Emoji.ID != "1117969363627159622" {
		return
	}
	message, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	if err != nil {
		log.Println("Error retrieving message for reaction checking!")
		return
	}
	if message.Author.ID != s.State.User.ID {
		return
	}
	if !strings.Contains(message.Content, "Remember to grab your") {
		return
	}

	regex := regexp.MustCompile("<@&(\\d+)>")
	roleID := regex.FindStringSubmatch(message.Content)[1]

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
	////////////////////////////// Check for wrong reach-mcc messages ///////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////
	if m.ChannelID == reachChannelID && !IsStaff(m.Member) {
		if strings.Contains(msg, "skunked") || (strings.Contains(msg, "negative") && strings.Contains(msg, "ghostrider")) {
			str := fmt.Sprintf("<@%s> That achievement is a multiplayer achievement. You should use <#984080289242427413> for multiplayer achievements!\n\n***Make sure to check pinned posts in the respective channel beforehand!***.", m.Author.ID)
			s.ChannelMessageSend(reachChannelID, str)
			s.ChannelMessageDelete(reachChannelID, m.ID)
			return
		}
	}
}
