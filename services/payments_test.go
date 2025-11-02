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
	"github.com/stretchr/testify/assert"
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
	givenPayInResponse("https://doesnt.matter", nil)

	// assert transaction is created
	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, paymentServiceProviderMock)
	_, err := sut.Deposit(req)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
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
	givenPayInResponse("", errors.New("Something went wrong QQ"))

	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, paymentServiceProviderMock)

	expectPanic(t, func() { sut.Deposit(req) })
}

func TestDeposit_MinimumAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Currency:      "TWD",
		Amount:        services.MIN_DEPOSIT_AMOUNT,
		PaymentMethod: "AnyPay",
	}

	givenUser(req.UserID)
	givenPayInResponse("https://doesnt.matter", nil)
	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, paymentServiceProviderMock)
	_, err := sut.Deposit(req)

	assert.Nil(t, err, "Expected no error, got: %v", err)
}

func TestDeposit_MaximumAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Currency:      "TWD",
		Amount:        services.MAX_DEPOSIT_AMOUNT,
		PaymentMethod: "AnyPay",
	}

	givenUser(req.UserID)
	givenPayInResponse("https://doesnt.matter", nil)
	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, paymentServiceProviderMock)
	_, err := sut.Deposit(req)

	assert.Nil(t, err, "Expected no error, got: %v", err)
}

func TestDeposit_BelowMinimum(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Currency:      "TWD",
		Amount:        services.MIN_DEPOSIT_AMOUNT - 1.00,
		PaymentMethod: "AnyPay",
	}

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, paymentServiceProviderMock)

	_, err := sut.Deposit(req)

	assert.NotNil(t, err, "Expected error for amount below minimum, got nil")
}

func TestDeposit_AboveMaximum(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	paymentServiceProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Currency:      "TWD",
		Amount:        services.MAX_DEPOSIT_AMOUNT + 1.00,
		PaymentMethod: "AnyPay",
	}

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, paymentServiceProviderMock)

	_, err := sut.Deposit(req)
	assert.NotNil(t, err, "Expected error for amount above maximum, got nil")
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

func givenPayInResponse(redirectUrl string, err error) {
	paymentServiceProviderMock.EXPECT().
		PayIn().
		Return(&psp.PayInResponse{
			RedirectUrl: redirectUrl,
		}, err).
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
