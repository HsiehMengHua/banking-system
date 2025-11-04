package psp

type PayInRequest struct {
	TransactionID      string
	Amount             float64
	ConfirmCallbackURL string
	CancelCallbackURL  string
}

type PayOutRequest struct {
	TransactionID string
	Amount        float64
}

type ConfirmRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

type CancelRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}
