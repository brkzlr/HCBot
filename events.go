package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func checkComboBreaker(s *discordgo.Session, m *discordgo.MessageCreate) {
	if comboMsg, exists := currentComboMsgs[m.ChannelID]; exists {
		if comboMsg != m.Content {
			str := fmt.Sprintf("<@%s> You broke the combo!! You absolute fucking buffoon!!", m.Author.ID)
			ReplyToMsg(s, m.Message, str)
			s.MessageReactionAdd(m.ChannelID, m.ID, "🤡")
			delete(currentComboMsgs, m.ChannelID)
		}
	} else {
		messages, err := s.ChannelMessages(m.ChannelID, 2, m.ID, "", "")
		if err != nil {
			return
		}

		for _, message := range messages {
			if message.Content != m.Content {
				return
			}
		}

		currentComboMsgs[m.ChannelID] = m.Content
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Author.Bot {
		return
	}

	// Should probably move these blocks below into their own function
	/////////////////////////////////////////////////////////////////////////////////////////////////////
	///////////////////////////////////// Check for discord invites /////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////
	msg := strings.ToLower(m.Content)
	if !IsStaff(m.Member) &&
		(strings.Contains(msg, "discord.gg/") || strings.Contains(msg, "discord.com/invite/")) {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		return
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////
	///////////////////////// Check for proof-of-completion channel misuse //////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////////////////////////
	if m.ChannelID == proofChannelID && !IsStaff(m.Member) {
		if (strings.Contains(msg, "mcc") && !strings.Contains(msg, "master")) ||
			strings.Contains(msg, "chief collection") ||
			strings.Contains(msg, "infinite") ||
			strings.Contains(msg, "legacy") ||
			strings.Contains(msg, "modern") {

			str := fmt.Sprintf("<@%s> You can only obtain that role by using me in <#%s>! Type `+help` in there to begin.", m.Author.ID, botChannelID)
			s.ChannelMessageSend(proofChannelID, str)
			s.ChannelMessageDelete(proofChannelID, m.ID)
			return
		} else if strings.Contains(msg, "halo completionist") || strings.Contains(msg, "hc") {
			str := fmt.Sprintf("Make sure you used the `+hc` command in <#%s> beforehand. This channel is only if SA/SS is bugged for you!", botChannelID)
			ReplyToMsg(s, m.Message, str)
			return
		} else {
			saidAcceptedWords := func(msg string) bool {
				return strings.Contains(msg, "franchise") ||
					strings.Contains(msg, "lasochist") ||
					strings.Contains(msg, "mcc master") ||
					strings.Contains(msg, "ice") ||
					strings.Contains(msg, "fire") ||
					strings.Contains(msg, "jacker") ||
					strings.Contains(msg, "jackal") ||
					strings.Contains(msg, "manual check")
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

	if !strings.HasPrefix(m.Content, "+") {
		checkComboBreaker(s, m)
		return
	}
	cmdContent := strings.Split(m.Content, " ")
	cmdName := strings.ToLower(cmdContent[0])

	if commandFnc, ok := commands[cmdName]; ok {
		go commandFnc(s, m.Message)
	} else {
		ReactFail(s, m.Message)
		s.ChannelMessageSend(m.ChannelID, "That command doesn't exist! Try `+help`")
	}
}
