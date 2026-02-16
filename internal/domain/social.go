package domain

import "time"

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Online   bool   `json:"online"`
}

type RoomSummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CurrentPeople int    `json:"currentPeople"`
	Capacity      int    `json:"capacity"`
	Mode          string `json:"mode"`
}

type MatchRecord struct {
	ID           string         `json:"id"`
	RoomID       string         `json:"roomId"`
	Participants []string       `json:"participants"`
	Winners      []string       `json:"winners"`
	Settlement   map[string]int `json:"settlement"`
	PlayedAt     time.Time      `json:"playedAt"`
}

type HeadToHeadStats struct {
	Me           string  `json:"me"`
	Opponent     string  `json:"opponent"`
	TotalRounds  int     `json:"totalRounds"`
	MeWins       int     `json:"meWins"`
	OpponentWins int     `json:"opponentWins"`
	Draws        int     `json:"draws"`
	MeWinRate    float64 `json:"meWinRate"`
}
