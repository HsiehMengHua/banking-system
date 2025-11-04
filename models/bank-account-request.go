package models

type CreateBankAccountRequest struct {
	UserID        uint   `json:"-"` // Read from header, not JSON
	BankCode      string `json:"bank_code" binding:"required,max=50"`
	AccountNumber string `json:"account_number" binding:"required,max=100"`
}

type UpdateBankAccountRequest struct {
	BankCode      string `json:"bank_code" binding:"required,max=50"`
	AccountNumber string `json:"account_number" binding:"required,max=100"`
}
