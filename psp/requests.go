package psp

type ConfirmRequest struct {
	TransactionID string
	Amount        float64
}

type CancelRequest struct {
	TransactionID string
}
