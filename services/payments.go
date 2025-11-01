package services

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/psp"
	"banking-system/repos"
	"log"

	"github.com/google/uuid"
)

//go:generate mockgen -source=payments.go -destination=mock/payments.go

type PaymentService interface {
	Deposit(req *models.DepositRequest) (redirectUrl string)
}

type paymentService struct {
	userRepo               repos.UserRepo
	transactionRepo        repos.TransactionRepo
	paymentServiceProvider psp.PaymentServiceProvider
}

func NewPaymentService(userRepo repos.UserRepo, transactionRepo repos.TransactionRepo, paymentServiceProvider psp.PaymentServiceProvider) PaymentService {
	return &paymentService{
		userRepo:               userRepo,
		transactionRepo:        transactionRepo,
		paymentServiceProvider: paymentServiceProvider,
	}
}

func (srv *paymentService) Deposit(req *models.DepositRequest) (redirectUrl string) {
	user, _ := srv.userRepo.Get(req.UserID)

	tx := &entities.Transaction{
		UUID:          uuid.New(),
		WalletID:      user.Wallet.ID,
		Amount:        req.Amount,
		Status:        entities.TransactionStatuses.Pending,
		Type:          entities.TransactionTypes.Deposit,
		PaymentMethod: req.PaymentMethod,
	}

	srv.transactionRepo.Create(tx)

	res, err := srv.paymentServiceProvider.PayIn()
	if err != nil {
		log.Panicf("Payment service provider '%s' error: %v", req.PaymentMethod, err)
	}

	return res.RedirectUrl
}
