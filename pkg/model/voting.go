package model

import "time"

type Voting struct {
	ID        string      `json:"id"`
	CreatorID string      `json:"creator_id"`
	Question  string      `json:"question"`
	ChannelID string      `json:"channel_id"`
	Options   []string    `json:"options"`
	CreatedAt time.Time   `json:"created_at"`
	ClosedAt  time.Time   `json:"closed_at"`
	Results   map[int]int `json:"results"`
	IsActive  bool
}
