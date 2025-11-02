package models

import "github.com/google/uuid"

type WithdrawRequest struct {
	UUID     uuid.UUID `json:"uuid"`
	UserID   uint      `json:"user_id" binding:"required"`
	Currency string    `json:"currency" binding:"required"`
	Amount   float64   `json:"amount" binding:"required,gt=0"`
}
