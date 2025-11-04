package psp

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type fakePay struct {
}

func NewFakePay() PaymentServiceProvider {
	return &fakePay{}
}

const (
	fakePayURL = "https://fake-payment-service-provider-production.up.railway.app"
)

func (*fakePay) PayIn() (*PayInResponse, error) {
	log.Debug("Simulate third party deposit process...")

	return &PayInResponse{
		TransactionID: "<transaction_id>",
		RedirectUrl:   fmt.Sprintf("%s", fakePayURL),
	}, nil
}

func (*fakePay) PayOut() (*PayOutResponse, error) {
	log.Debug("Simulate third party withdrawal process...")

	return &PayOutResponse{
		TransactionID: "<transaction_id>",
	}, nil
}
