package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			s.MessageReactionAdd(m.ChannelID, m.ID, "🤡")
			ReplyToMsg(s, m.Message, "Don't ping.")
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
