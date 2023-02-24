package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Author.Bot {
		return
	}
	// Explain users that ping the bot how to use the bot.
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID && m.MessageReference == nil { // Check for MessageReference since Mentions include replies.
			s.MessageReactionAdd(m.ChannelID, m.ID, "🤡")
			str := fmt.Sprintf("Don't ping me. Type `+help` in <#%s>", botChannelID)
			ReplyToMsg(s, m.Message, str)
			return
		}
	}

	if m.ChannelID == proofChannelID && !IsStaff(m.Member) {
		msg := strings.ToLower(m.Content)
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
			saidProofRole := func(msg string) bool {
				return strings.Contains(msg, "franchise") ||
					strings.Contains(msg, "lasochist") ||
					strings.Contains(msg, "mcc master") ||
					strings.Contains(msg, "ice") ||
					strings.Contains(msg, "fire") ||
					strings.Contains(msg, "jacker") || strings.Contains(msg, "jackal")
			}
			for _, attach := range m.Attachments {
				if attach != nil && attach.Height != 0 && !saidProofRole(msg) { // Posted image but didn't say an acceptable proof role
					str := fmt.Sprintf("<@%s> Please state the exact name of the vanity role you wish to obtain in the same message as the image!", m.Author.ID)
					s.ChannelMessageSend(proofChannelID, str)
					s.ChannelMessageDelete(proofChannelID, m.ID)
					return
				}
			}
		}
	}
	if !strings.HasPrefix(m.Content, "+") {
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
