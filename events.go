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
	//Explain users that ping the bot how to use the bot.
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID && m.MessageReference == nil { //Check for MessageReference since Mentions include replies.
			s.MessageReactionAdd(m.ChannelID, m.ID, "🤡")
			str := fmt.Sprintf("Don't ping me. Type `+help` in <#%s>", botChannelID)
			ReplyToMsg(s, m.Message, str)
			return
		}
	}

	if m.ChannelID == proofChannelID && !IsStaff(m.Member) {
		msg := strings.ToLower(m.Content)
		if (strings.Contains(msg, "mcc") && !strings.Contains(msg, "master")) ||
			strings.Contains(msg, "infinite") ||
			strings.Contains(msg, "legacy") ||
			strings.Contains(msg, "modern") {

			str := fmt.Sprintf("<@%s> Read the description of the channel please!", m.Author.ID)
			s.ChannelMessageSend(proofChannelID, str)
			s.ChannelMessageDelete(proofChannelID, m.ID)
			return
		} else if strings.Contains(msg, "halo completionist") || strings.Contains(msg, "hc") {
			str := fmt.Sprintf("Make sure you used the `+hc` command in <#%s> beforehand. This channel is only if SA/SS is bugged for you!", botChannelID)
			ReplyToMsg(s, m.Message, str)
			return
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
