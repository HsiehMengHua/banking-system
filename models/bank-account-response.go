package models

import "time"

type BankAccountResponse struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	BankCode      string    `json:"bank_code"`
	AccountNumber string    `json:"account_number"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
