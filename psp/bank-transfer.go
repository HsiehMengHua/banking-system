package psp

import (
	log "github.com/sirupsen/logrus"
)

type bankTransfer struct {
}

func NewBankTransferPSP() PaymentServiceProvider {
	return &bankTransfer{}
}

func (*bankTransfer) PayIn(req *PayInRequest) (*PayInResponse, error) {
	log.Debug("Simulate bank transfer deposit process...")

	return &PayInResponse{
		TransactionID: req.TransactionID,
	}, nil
}

func (*bankTransfer) PayOut() (*PayOutResponse, error) {
	log.Debug("Simulate bank transfer withdrawal process...")

	return &PayOutResponse{}, nil
}
