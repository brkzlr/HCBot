package main

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	openai "github.com/sashabaranov/go-openai"
)

// Constants, mostly channel or role IDs
const (
	mccRoleID       = "985327507274350612"
	infiniteRoleID  = "985327648605614140"
	modernRoleID    = "985327740112760873"
	legacyRoleID    = "985327809566232626"
	lasochistRoleID = "985644631423320064"
	mccMasterRoleID = "985644874051231825"
	hcRoleID        = "985327939161849857"
	fcRoleID        = "985328007088590918"

	botChannelID   = "984075793904848916"
	proofChannelID = "984079675385077820"

	dropsChannelID = "984078138332028969"
	dropsRoleID    = "984088663946326018"

	hcGuildID = "984075026816991252"
)

// Global variables
var (
	tokens struct {
		Discord string `json:"discordToken"`
		OpenXBL string `json:"xblToken"`
		OpenAI  string `json:"aiToken"`
	}

	slashCommandsHandlers = make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))

	databaseMap   map[string]string
	dirtyDatabase bool
	databaseLock  sync.Mutex
	slashCommands []*discordgo.ApplicationCommand

	// AI stuff
	aiClient         *openai.Client
	contextMessages  []openai.ChatCompletionMessage
	dirtyContextMsg  bool
	chatRequestQueue []ChatRequest
	chatRequestLock  sync.Mutex
)

// Timed achievements variables
var (
	timedAchievRoles = map[int]TimedRoles{
		15: {990601630237982750, "Halo: Combat Evolved"},
		9:  {990601760961875998, "Halo 2 Classic"},
		25: {990601846391463967, "Halo 3"},
		6:  {990602246888779837, "Halo 4"},
		22: {990601924703297586, "Halo 3: ODST"},
		14: {990602198184501268, "Halo: Reach"},
	}
	destinationVacationDates = [...]RoleDate{
		{1, time.January},
		{7, time.July},
		{31, time.October},
		{25, time.December},
	}
	elderSignsDates = [...]RoleDate{
		{1, time.January},
		{22, time.April},
		{5, time.May},
		{4, time.July},
		{7, time.July},
		{31, time.October},
		{11, time.November},
		{25, time.December},
	}
)
