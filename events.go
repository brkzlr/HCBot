package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
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
