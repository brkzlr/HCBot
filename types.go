package main

import "time"

type X360AchievsResp struct {
	StatusCode int `json:"code"`
	Content    struct {
		Achievements []struct {
			ID                int       `json:"id"`
			TitleID           int       `json:"titleId"`
			Name              string    `json:"name"`
			Sequence          int       `json:"sequence"`
			Flags             int       `json:"flags"`
			UnlockedOnline    bool      `json:"unlockedOnline"`
			Unlocked          bool      `json:"unlocked"`
			IsSecret          bool      `json:"isSecret"`
			Platform          int       `json:"platform"`
			Gamerscore        int       `json:"gamerscore"`
			ImageID           int       `json:"imageId"`
			Description       string    `json:"description"`
			LockedDescription string    `json:"lockedDescription"`
			Type              int       `json:"type"`
			IsRevoked         bool      `json:"isRevoked"`
			TimeUnlocked      time.Time `json:"timeUnlocked"`
			Rarity            struct {
				CurrentCategory   string  `json:"currentCategory"`
				CurrentPercentage float64 `json:"currentPercentage"`
			} `json:"rarity"`
		} `json:"achievements"`
		Version    time.Time `json:"version"`
		PagingInfo struct {
			ContinuationToken any `json:"continuationToken"`
			TotalRecords      int `json:"totalRecords"`
		} `json:"pagingInfo"`
	} `json:"content"`
}

type AchievementsResp struct {
	StatusCode int `json:"code"`
	Content    struct {
		Titles []struct {
			TitleID       string `json:"titleId"`
			Name          string `json:"name"`
			DisplayImage  string `json:"displayImage"`
			ModernTitleID string `json:"modernTitleId"`
			IsBundle      bool   `json:"isBundle"`
			Achievement   struct {
				CurrentAchievements int     `json:"currentAchievements"`
				TotalAchievements   int     `json:"totalAchievements"`
				CurrentGamerscore   int     `json:"currentGamerscore"`
				TotalGamerscore     int     `json:"totalGamerscore"`
				ProgressPercentage  float64 `json:"progressPercentage"`
			} `json:"achievement"`
			Stats    any `json:"stats"`
			GamePass any `json:"gamePass"`
			Images   []struct {
				URL  string `json:"url"`
				Type string `json:"type"`
			} `json:"images"`
			TitleHistory struct {
				LastTimePlayed time.Time `json:"lastTimePlayed"`
				Visible        bool      `json:"visible"`
				CanHide        bool      `json:"canHide"`
			} `json:"titleHistory"`
			Detail            any `json:"detail"`
			FriendsWhoPlayed  any `json:"friendsWhoPlayed"`
			AlternateTitleIds any `json:"alternateTitleIds"`
			ContentBoards     any `json:"contentBoards"`
		} `json:"titles"`
	} `json:"content"`
}

type GTResp struct {
	StatusCode int `json:"code"`
	Content    struct {
		ProfileUsers []struct {
			ID string `json:"id"`
		} `json:"profileUsers"`
	} `json:"content"`
}

type RoleDate struct {
	Day   int
	Month time.Month
}

type TimedRoles struct {
	ID   int
	Game string
}

// "Enum" type
type GameStatus int

const (
	NOT_FOUND     GameStatus = 0
	NOT_COMPLETED GameStatus = 1
	COMPLETED     GameStatus = 2
)
