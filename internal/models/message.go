package models

import "time"

type Message struct {
	ID        string    `json:"id" db:"id"`
	From      string    `json:"from" db:"from_client"`
	To        string    `json:"to" db:"to_client"`
	Channel   string    `json:"channel" db:"channel"`
	Content   string    `json:"content" db:"content"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}