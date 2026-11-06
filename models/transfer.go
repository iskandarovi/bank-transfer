package models

import "time"

type Transfer struct {
	ID           int64     `json:"id"`
	SenderID     int64     `json:"sender_id"`
	SenderName   string    `json:"sender_name"`
	ReceiverID   int64     `json:"receiver_id"`
	ReceiverName string    `json:"receiver_name"`
	Amount       float64   `json:"amount"`
	CreatedAt    time.Time `json:"created_at"`
}
