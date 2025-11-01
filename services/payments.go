package services

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/repos"

	"github.com/google/uuid"
)

//go:generate mockgen -source=payments.go -destination=mock/payments.go

type PaymentService interface {
	Deposit(req *models.DepositRequest)
}

type paymentService struct {
	userRepo        repos.UserRepo
	transactionRepo repos.TransactionRepo
}

func NewPaymentService(userRepo repos.UserRepo, transactionRepo repos.TransactionRepo) PaymentService {
	return &paymentService{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

func (srv *paymentService) Deposit(req *models.DepositRequest) {
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
}
