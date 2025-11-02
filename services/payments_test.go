package services_test

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/psp"
	"banking-system/services"
	"errors"
	"testing"

	pspMock "banking-system/psp/mock"
	repoMock "banking-system/repos/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

var (
	userRepoMock               *repoMock.MockUserRepo
	transactionRepoMock        *repoMock.MockTransactionRepo
	paymentServiceProviderMock *pspMock.MockPaymentServiceProvider
)

func TestValidDeposit(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Currency:      "TWD",
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	}

	givenUser(req.UserID)
	givenPspResponse()
	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, paymentServiceProviderMock)
	sut.Deposit(req)
}

func TestDeposit_PspRespondsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Currency:      "TWD",
		Amount:        100.00,
		PaymentMethod: "FragilePay",
	}

	givenUser(req.UserID)
	givenPspRespondsError()

	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, paymentServiceProviderMock)

	expectPanic(t, func() { sut.Deposit(req) })
}

func givenUser(userId uint) {
	userRepoMock.EXPECT().
		Get(gomock.Any()).
		Return(&entities.User{
			Wallet: entities.Wallet{
				UserID: userId,
			},
		}, nil).
		AnyTimes()
}

func givenPspResponse() {
	paymentServiceProviderMock.EXPECT().
		PayIn().
		Return(&psp.DepositResponse{
			RedirectUrl: "https://external.payment.page/payin",
		}, nil).
		Times(1)
}

func givenPspRespondsError() {
	paymentServiceProviderMock.EXPECT().
		PayIn().
		Return(&psp.DepositResponse{}, errors.New("Something went wrong QQ")).
		Times(1)
}

func expectTransactionCreated() {
	transactionRepoMock.EXPECT().
		Create(gomock.Any()).
		Times(1)
}

func expectPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}
