package psp

type PayInResponse struct {
	TransactionID string
	RedirectUrl   string
}

type PayOutResponse struct {
	TransactionID string
}
