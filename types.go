package main

import (
	"time"
)

type Game struct {
	Name  string `json:"name"`
	Stats struct {
		CurrentGScore int `json:"currentGamerscore"`
		TotalGScore   int `json:"totalGamerscore"`
	} `json:"achievement"`
	TitleID string `json:"titleId"`
}

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
