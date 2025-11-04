package models

import (
	"banking-system/psp"

	"github.com/google/uuid"
)

type WithdrawRequest struct {
	UUID          uuid.UUID `json:"uuid"`
	UserID        uint
	Amount        float64           `json:"amount" binding:"required,gt=0"`
	PaymentMethod psp.PaymentMethod `json:"payment_method" binding:"required"`
	BankAccountID uint              `json:"bank_account_id"`
}
