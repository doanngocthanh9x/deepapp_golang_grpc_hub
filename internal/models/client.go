package models

import "time"

type Client struct {
	ID       string    `json:"id" db:"id"`
	Status   string    `json:"status" db:"status"`
	LastSeen time.Time `json:"last_seen" db:"last_seen"`
}