package services

//go:generate mockgen -source=promotions.go -destination=mock/prmootions.go

type PaymentService interface {
	Deposit()
}

type paymentService struct {
}

func NewPaymentService() PaymentService {
	return &paymentService{}
}

func (*paymentService) Deposit() {
}
