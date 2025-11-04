package models

import "github.com/google/uuid"

type WithdrawRequest struct {
	UUID   uuid.UUID `json:"uuid"`
	UserID uint
	Amount float64 `json:"amount" binding:"required,gt=0"`
}
