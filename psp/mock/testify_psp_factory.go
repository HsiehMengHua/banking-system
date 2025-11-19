package mock_psp

import (
	"banking-system/psp"

	"github.com/stretchr/testify/mock"
)

type MockPSPFactoryTestify struct {
	mock.Mock
}

func (m *MockPSPFactoryTestify) NewPaymentServiceProvider(paymentMethod psp.PaymentMethod) psp.PaymentServiceProvider {
	args := m.Called(paymentMethod)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(psp.PaymentServiceProvider)
}
