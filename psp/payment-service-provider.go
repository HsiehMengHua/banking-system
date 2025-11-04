package psp

//go:generate mockgen -source=psp.go -destination=mock/psp.go

type PaymentServiceProvider interface {
	PayIn() (*PayInResponse, error)
	PayOut() (*PayOutResponse, error)
}
