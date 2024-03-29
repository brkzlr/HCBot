package main

import (
	"time"
)

type AchievListInfo struct {
	PagingInfo struct {
		TotalRecords int `json:"totalRecords"`
	} `json:"pagingInfo"`
}

type GameStatsResp struct {
	Stats struct {
		CurrentGScore int `json:"currentGamerscore"`
		TotalGScore   int `json:"totalGamerscore"`
	} `json:"achievement"`
	TitleID string `json:"titleId"`
}

// "Enum" type
type GameStatus int

const (
	NOT_FOUND     GameStatus = 0
	NOT_COMPLETED GameStatus = 1
	COMPLETED     GameStatus = 2
)

///////////////

type GTResp struct {
	ID string `json:"id"`
}

type Riddle struct {
	Question string `json:"riddle"`
	Answer   string `json:"answer"`
}

type RoleDate struct {
	Day   int
	Month time.Month
}

type TimedRoles struct {
	ID   int
	Game string
}
