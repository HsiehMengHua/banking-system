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
	userRepoMock        *repoMock.MockUserRepo
	transactionRepoMock *repoMock.MockTransactionRepo
	pspFactoryMock      *pspMock.MockPSPFactory
	paymentProviderMock *pspMock.MockPaymentServiceProvider
)

func TestValidDeposit(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)
	paymentProviderMock = pspMock.NewMockPaymentServiceProvider(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	}

	givenUserHasBalance(req.UserID, 0)
	givenPayInResponse("https://doesnt.matter", nil)

	// assert transaction is created
	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, pspFactoryMock)
	_, err := sut.Deposit(req)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestDeposit_PspRespondsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Amount:        100.00,
		PaymentMethod: "FragilePay",
	}

	givenUserHasBalance(req.UserID, 0)
	givenPayInResponse("", errors.New("Something went wrong QQ"))

	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, pspFactoryMock)

	expectPanic(t, func() { sut.Deposit(req) })
}

func TestDeposit_MinimumAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Amount:        services.MIN_DEPOSIT_AMOUNT,
		PaymentMethod: "AnyPay",
	}

	givenUserHasBalance(req.UserID, 0)
	givenPayInResponse("https://doesnt.matter", nil)
	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, pspFactoryMock)
	_, err := sut.Deposit(req)

	assert.Nil(t, err, "Expected no error, got: %v", err)
}

func TestDeposit_MaximumAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Amount:        services.MAX_DEPOSIT_AMOUNT,
		PaymentMethod: "AnyPay",
	}

	givenUserHasBalance(req.UserID, 0)
	givenPayInResponse("https://doesnt.matter", nil)
	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, pspFactoryMock)
	_, err := sut.Deposit(req)

	assert.Nil(t, err, "Expected no error, got: %v", err)
}

func TestDeposit_BelowMinimum(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Amount:        services.MIN_DEPOSIT_AMOUNT - 1.00,
		PaymentMethod: "AnyPay",
	}

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, pspFactoryMock)

	_, err := sut.Deposit(req)

	assert.NotNil(t, err, "Expected error for amount below minimum, got nil")
}

func TestDeposit_AboveMaximum(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)

	req := &models.DepositRequest{
		UUID:          uuid.New(),
		UserID:        1,
		Amount:        services.MAX_DEPOSIT_AMOUNT + 1.00,
		PaymentMethod: "AnyPay",
	}

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, pspFactoryMock)

	_, err := sut.Deposit(req)
	assert.NotNil(t, err, "Expected error for amount above maximum, got nil")
}

func TestWithdraw_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)

	req := &models.WithdrawRequest{
		UUID:   uuid.New(),
		UserID: 1,
		Amount: 50.00,
	}

	// given user has sufficient balance
	givenUserHasBalance(req.UserID, 100.00)

	// assertions
	expectTransactionCreated()
	expectWalletUpdated()
	expectPayOutCalled()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, pspFactoryMock)
	err := sut.Withdraw(req)

	assert.Nil(t, err)
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)
	pspFactoryMock = pspMock.NewMockPSPFactory(ctrl)

	req := &models.WithdrawRequest{
		UUID:   uuid.New(),
		UserID: 1,
		Amount: 150.00, // More than balance
	}

	givenUserHasBalance(req.UserID, 100)

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock, pspFactoryMock)
	err := sut.Withdraw(req)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
}

func givenUserHasBalance(userID uint, amount float64) {
	userRepoMock.EXPECT().
		Get(gomock.Any()).
		Return(&entities.User{
			Wallet: entities.Wallet{
				UserID:  userID,
				Balance: amount,
			},
		}, nil).
		AnyTimes()
}

func givenPayInResponse(redirectUrl string, err error) {
	pspFactoryMock.EXPECT().
		NewPaymentServiceProvider(gomock.Any()).
		Return(paymentProviderMock).
		Times(1)

	paymentProviderMock.EXPECT().
		PayIn(gomock.Any()).
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

func expectWalletUpdated() {
	userRepoMock.EXPECT().
		UpdateWallet(gomock.Any()).
		Times(1)
}

func expectPayOutCalled() {
	pspFactoryMock.EXPECT().
		NewPaymentServiceProvider(gomock.Any()).
		Return(paymentProviderMock).
		Times(1)

	paymentProviderMock.EXPECT().
		PayOut().
		Return(&psp.PayOutResponse{}, nil).
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
