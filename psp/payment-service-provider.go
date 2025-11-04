package psp

//go:generate mockgen -source=payment-service-provider.go -destination=mock/payment-service-provider.go

type PaymentServiceProvider interface {
	PayIn(req *PayInRequest) (*PayInResponse, error)
	PayOut() (*PayOutResponse, error)
}
