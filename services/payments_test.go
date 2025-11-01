package services_test

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/services"
	"testing"

	repoMock "banking-system/repos/mock"

	"github.com/golang/mock/gomock"
)

var (
	userRepoMock        *repoMock.MockUserRepo
	transactionRepoMock *repoMock.MockTransactionRepo
)

func TestValidDeposit(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepoMock = repoMock.NewMockUserRepo(ctrl)
	transactionRepoMock = repoMock.NewMockTransactionRepo(ctrl)

	req := &models.DepositRequest{
		UserID:        1,
		Currency:      "TWD",
		Amount:        100.00,
		PaymentMethod: "AnyPay",
	}

	givenUser(req.UserID)
	expectTransactionCreated()

	sut := services.NewPaymentService(userRepoMock, transactionRepoMock)
	sut.Deposit(req)
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

func expectTransactionCreated() {
	transactionRepoMock.EXPECT().
		Create(gomock.Any()).
		Times(1)
}
