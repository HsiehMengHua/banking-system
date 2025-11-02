package models

import "github.com/google/uuid"

type TransferRequest struct {
	UUID            uuid.UUID `json:"uuid"`
	SenderUserID    uint      `json:"sender_user_id" binding:"required"`
	RecipientUserID uint      `json:"recipient_user_id" binding:"required"`
	Currency        string    `json:"currency" binding:"required"`
	Amount          float64   `json:"amount" binding:"required,gt=0"`
}
