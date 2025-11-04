package models

import (
	"banking-system/psp"

	"github.com/google/uuid"
)

type DepositRequest struct {
	UUID          uuid.UUID `json:"uuid" binding:"required"`
	UserID        uint
	Amount        float64           `json:"amount" binding:"required,gt=0"`
	PaymentMethod psp.PaymentMethod `json:"payment_method"`
}
