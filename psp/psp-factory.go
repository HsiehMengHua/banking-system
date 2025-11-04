package psp

import "log"

//go:generate mockgen -source=psp-factory.go -destination=mock/psp-factory.go

type PSPFactory interface {
	NewPaymentServiceProvider(paymentMethod PaymentMethod) PaymentServiceProvider
}

type pspFactory struct {
}

func NewPSPFactory() PSPFactory {
	return &pspFactory{}
}

type PaymentMethod string

var PaymentMethods = struct {
	FakePay PaymentMethod

	// others for example...
	// CreditCard PaymentMethod
	// PayPal     PaymentMethod
}{
	FakePay: "fake_pay",

	// others for example...
	// CreditCard: "credit_card",
	// PayPal:     "paypal",
}

func (f *pspFactory) NewPaymentServiceProvider(paymentMethod PaymentMethod) PaymentServiceProvider {
	switch paymentMethod {
	case PaymentMethods.FakePay:
		return NewFakePay()

	default:
		log.Panicf("payment method '%s' is not supported", paymentMethod)
		return nil
	}
}
