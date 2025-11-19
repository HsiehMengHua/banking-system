package mock_psp

import (
	"banking-system/psp"

	"github.com/stretchr/testify/mock"
)

type MockPaymentServiceProviderTestify struct {
	mock.Mock
}

func (m *MockPaymentServiceProviderTestify) PayIn(req *psp.PayInRequest) (*psp.PayInResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*psp.PayInResponse), args.Error(1)
}

func (m *MockPaymentServiceProviderTestify) PayOut() (*psp.PayOutResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*psp.PayOutResponse), args.Error(1)
}
