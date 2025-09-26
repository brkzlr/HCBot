package main

import (
	"database/sql"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Constants
const (
	mccRoleID       = "985327507274350612"
	mccChinaRoleID  = "1206765698869887016"
	infiniteRoleID  = "985327648605614140"
	modernRoleID    = "985327740112760873"
	legacyRoleID    = "985327809566232626"
	lasochistRoleID = "985644631423320064"
	mccMasterRoleID = "985644874051231825"
	hcRoleID        = "985327939161849857"
	fcRoleID        = "985328007088590918"

	coopPCRole   = "984090494067949620"
	coopXboxRole = "984090416150364180"

	botChannelID            = "984075793904848916"
	chinaMccChannelID       = "984080266135994418"
	generalMccChannelID     = "984160204440633454"
	mccCategoryID           = "984080152088698920"
	proofChannelID          = "984079675385077820"
	multiplayerMccChannelID = "984080289242427413"

	dropsChannelID = "984078138332028969"
	dropsRoleID    = "984088663946326018"

	hcGuildID = "984075026816991252"

	// Title IDs
	mccTitleID      = "1144039928"
	mccChinaTitleID = "812065290"
	hwdeTitleID     = "1030018025"
	h5TitleID       = "219630713"
	h5ForgeTitleID  = "766737132"
	hw2TitleID      = "1667928394"
	infiniteTitleID = "2043073184"
	hceaTitleID     = "1297287601"
	h3TitleID       = "1297287142"
	hwTitleID       = "1297287176"
	odstTitleID     = "1297287287"
	reachTitleID    = "1297287259"
	h4TitleID       = "1297287449"
	h2TitleID       = "1297287183"

	hsaTitleID     = "1297292157"
	hsaXboxTitleID = "682562723"
	hsa360TitleID  = "1480659986"
	hsaWPTitleID   = "1297290378"
	hsaIOSTitleID  = "1297291180"

	hssTitleID    = "1297292194"
	hssWPTitleID  = "1297290417"
	hssIOSTitleID = "1297291181"
)

// Global variables
var (
	tokens struct {
		Discord string `json:"discordToken"`
		OpenXBL string `json:"xblToken"`
	}

	appCommandsHandlers = make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))

	appCommands []*discordgo.ApplicationCommand
	database    *sql.DB

	multiplayerAchievRegex     = regexp.MustCompile("skunked|invasion|headhunter|(negative.*ghostrider)|skull(a)?manjaro|grand theft halo|double down|blown out.*sky|decorated warrior|ninja redux|flaming ninja anniversary|legend slayer|invaders repelled|put up your dukes|red vs\\.? blue|shield(s)? up|shook the hornet('s|s)? nest|skeet shooter|goose is loose|top gungoose|bombing run|cold as ice|cold fusion|counter(-|\\s)?snipe|triple threat|wetwork|roadkill rampage|rock and coil|splatter|environmentalist|requiescat|easy to overlook")
	multiplayerSoloAchievRegex = regexp.MustCompile("tour of duty|eminent domain|gate your thirst|high altitude thirst|rule your thirst|thirst locked down|bloody thirsty|worship your thirst|blastacular|power play|stayin(g|')? alive")
	platformRegex              = regexp.MustCompile("(?:\\(|\\[).*(pc|xbox).*(?:\\)|\\])")
	roleRegex                  = regexp.MustCompile("<@&(\\d+)>")
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
