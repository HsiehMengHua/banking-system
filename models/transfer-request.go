package models

import "github.com/google/uuid"

type TransferRequest struct {
	UUID              uuid.UUID `json:"uuid"`
	SenderUserID      uint
	RecipientUsername string  `json:"recipient_username" binding:"required"`
	Amount            float64 `json:"amount" binding:"required,gt=0"`
}
