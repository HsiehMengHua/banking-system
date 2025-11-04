package psp

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

type fakePay struct {
}

func NewFakePay() PaymentServiceProvider {
	return &fakePay{}
}

func (*fakePay) PayIn(req *PayInRequest) (*PayInResponse, error) {
	log.Debug("Simulate third party deposit process...")

	return &PayInResponse{
		TransactionID: req.TransactionID,
		RedirectUrl:   fmt.Sprintf("%s/payin/%s?merchant=MH&amount=%.2f&confirm_callback=%s&cancel_callback=%s", os.Getenv("FAKE_PAYMENT_PROVIDER_URL"), req.TransactionID, req.Amount, req.ConfirmCallbackURL, req.CancelCallbackURL),
	}, nil
}

func (*fakePay) PayOut() (*PayOutResponse, error) {
	log.Debug("Simulate third party withdrawal process...")

	return &PayOutResponse{
		TransactionID: "<transaction_id>",
	}, nil
}
