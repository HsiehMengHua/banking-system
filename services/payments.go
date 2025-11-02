package services

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/psp"
	"banking-system/repos"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

//go:generate mockgen -source=payments.go -destination=mock/payments.go

type PaymentService interface {
	Deposit(req *models.DepositRequest) (redirectUrl string)
	Confirm(req *psp.ConfirmRequest)
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
		UUID:          req.UUID,
		WalletID:      user.Wallet.ID,
		Amount:        req.Amount,
		Status:        entities.TransactionStatuses.Pending,
		Type:          entities.TransactionTypes.Deposit,
		PaymentMethod: req.PaymentMethod,
	}

	res, err := srv.paymentServiceProvider.PayIn()
	if err != nil {
		log.Panicf("Payment service provider '%s' error: %v", req.PaymentMethod, err)
	}

	if err := srv.transactionRepo.Create(tx); err != nil {
		log.Panicf("Failed to create transaction: %v", err)
	}

	return res.RedirectUrl
}

func (srv *paymentService) Confirm(req *psp.ConfirmRequest) {
	tx, err := srv.transactionRepo.GetByUUID(uuid.MustParse(req.TransactionID))
	if err != nil {
		log.Panicf("Failed to get transaction: %v", err)
	}

	if tx.Status != entities.TransactionStatuses.Pending {
		log.Debugf("Transaction %s is not in PENDING status, current status: %s", req.TransactionID, tx.Status)
		return
	}

	tx.Status = entities.TransactionStatuses.Completed

	switch tx.Type {
	case entities.TransactionTypes.Deposit:
		tx.Wallet.Balance += tx.Amount
	default:
		log.Panicf("Unknown transaction type: %s", tx.Type)
	}

	srv.transactionRepo.Update(tx)
}
