package models

type DepositRequest struct {
	UserID        uint    `json:"user_id" binding:"required"`
	Currency      string  `json:"currency" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method"`
}
