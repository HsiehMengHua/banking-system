package psp

type ConfirmRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

type CancelRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}
