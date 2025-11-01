package psp

import (
	log "github.com/sirupsen/logrus"
)

//go:generate mockgen -source=psp.go -destination=mock/psp.go

type PaymentServiceProvider interface {
	PayIn() (*DepositResponse, error)
}

type paymentServiceProvider struct {
}

func NewPaymentServiceProvider() PaymentServiceProvider {
	return &paymentServiceProvider{}
}

func (*paymentServiceProvider) PayIn() (*DepositResponse, error) {
	log.Debug("Simulate third party deposit process...")

	return &DepositResponse{
		TransactionID: "<transaction_id>",
		RedirectUrl:   "<redirect_url>",
	}, nil
}
