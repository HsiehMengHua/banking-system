package psp

import (
	log "github.com/sirupsen/logrus"
)

//go:generate mockgen -source=psp.go -destination=mock/psp.go

type PaymentServiceProvider interface {
	PayIn() (*PayInResponse, error)
	PayOut() (*PayOutResponse, error)
}

type paymentServiceProvider struct {
}

func NewPaymentServiceProvider() PaymentServiceProvider {
	return &paymentServiceProvider{}
}

func (*paymentServiceProvider) PayIn() (*PayInResponse, error) {
	log.Debug("Simulate third party deposit process...")

	return &PayInResponse{
		TransactionID: "<transaction_id>",
		RedirectUrl:   "<redirect_url>",
	}, nil
}

func (*paymentServiceProvider) PayOut() (*PayOutResponse, error) {
	log.Debug("Simulate third party withdrawal process...")

	return &PayOutResponse{
		TransactionID: "<transaction_id>",
	}, nil
}
