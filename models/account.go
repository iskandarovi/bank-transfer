package models

import "time"

type Account struct {
	ID        int64     `json:"id"`
	OwnerName string    `json:"owner_name"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}
