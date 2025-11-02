package services

import (
	"banking-system/entities"
	"banking-system/models"
	"banking-system/psp"
	"banking-system/repos"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

//go:generate mockgen -source=payments.go -destination=mock/payments.go

const (
	MIN_DEPOSIT_AMOUNT = 1.00
	MAX_DEPOSIT_AMOUNT = 100000.00
)

type PaymentService interface {
	Deposit(req *models.DepositRequest) (redirectUrl string, err error)
	Withdraw(req *models.WithdrawRequest) error
	Confirm(req *psp.ConfirmRequest) error
	Cancel(req *psp.CancelRequest) error
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

func (srv *paymentService) Deposit(req *models.DepositRequest) (redirectUrl string, err error) {
	if req.Amount < MIN_DEPOSIT_AMOUNT {
		return "", fmt.Errorf("deposit amount %.2f is below minimum allowed amount %.2f", req.Amount, MIN_DEPOSIT_AMOUNT)
	}

	if req.Amount > MAX_DEPOSIT_AMOUNT {
		return "", fmt.Errorf("deposit amount %.2f exceeds maximum allowed amount %.2f", req.Amount, MAX_DEPOSIT_AMOUNT)
	}

	user, _ := srv.userRepo.Get(req.UserID)

	tx := &entities.Transaction{
		UUID:          req.UUID,
		WalletID:      user.Wallet.ID,
		Amount:        req.Amount,
		Status:        entities.TransactionStatuses.Pending,
		Type:          entities.TransactionTypes.Deposit,
		PaymentMethod: req.PaymentMethod,
	}

	if err := srv.transactionRepo.Create(tx); err != nil {
		log.Panicf("Failed to create transaction: %v", err)
	}

	res, err := srv.paymentServiceProvider.PayIn()
	if err != nil {
		log.Panicf("Payment service provider '%s' error: %v", req.PaymentMethod, err)
	}

	return res.RedirectUrl, nil
}

func (srv *paymentService) Withdraw(req *models.WithdrawRequest) error {
	user, err := srv.userRepo.Get(req.UserID)
	if err != nil {
		log.Panicf("Failed to get user with ID %d: %v", req.UserID, err)
	}

	if user.Wallet.Balance < req.Amount {
		return fmt.Errorf("insufficient balance: current balance %.2f, requested amount %.2f", user.Wallet.Balance, req.Amount)
	}

	tx := &entities.Transaction{
		UUID:     req.UUID,
		WalletID: user.Wallet.ID,
		Amount:   req.Amount,
		Status:   entities.TransactionStatuses.Pending,
		Type:     entities.TransactionTypes.Withdrawal,
		Wallet:   user.Wallet,
	}

	if err := srv.transactionRepo.Create(tx); err != nil {
		log.Panicf("Failed to create transaction: %v", err)
	}

	user.Wallet.Balance -= req.Amount
	if err := srv.userRepo.UpdateWallet(user); err != nil {
		log.Panicf("Failed to update wallet for user ID %d: %v", req.UserID, err)
	}

	_, err = srv.paymentServiceProvider.PayOut()
	if err != nil {
		log.Panicf("Payment service provider error: %v", err)
	}

	return nil
}

func (srv *paymentService) Confirm(req *psp.ConfirmRequest) error {
	tx, err := srv.transactionRepo.GetByUUID(uuid.MustParse(req.TransactionID))
	if err != nil {
		log.Panicf("Failed to get transaction: %v", err)
	}

	if tx.Status != entities.TransactionStatuses.Pending {
		log.Infof("Transaction '%s' already processed with status: %s", req.TransactionID, tx.Status)
		return nil
	}

	tx.Status = entities.TransactionStatuses.Completed

	switch tx.Type {
	case entities.TransactionTypes.Deposit:
		tx.Wallet.Balance += tx.Amount
	default:
		log.Panicf("Unknown transaction type: %s", tx.Type)
	}

	updated, err := srv.transactionRepo.UpdateConditional(tx, entities.TransactionStatuses.Pending)
	if err != nil {
		log.Panicf("Failed to update transaction: %v", err)
	}

	if !updated {
		log.Infof("Transaction '%s' was already processed by another request", req.TransactionID)
	}

	return nil
}

func (srv *paymentService) Cancel(req *psp.CancelRequest) error {
	tx, err := srv.transactionRepo.GetByUUID(uuid.MustParse(req.TransactionID))
	if err != nil {
		log.Panicf("Failed to get transaction: %v", err)
	}

	if tx.Status != entities.TransactionStatuses.Pending {
		log.Infof("Transaction '%s' already processed with status: %s", req.TransactionID, tx.Status)
		return nil
	}

	tx.Status = entities.TransactionStatuses.Canceled

	switch tx.Type {
	case entities.TransactionTypes.Withdrawal:
		tx.Wallet.Balance += tx.Amount
	case entities.TransactionTypes.Deposit:
	default:
		log.Panicf("Unknown transaction type: %s", tx.Type)
	}

	updated, err := srv.transactionRepo.UpdateConditional(tx, entities.TransactionStatuses.Pending)
	if err != nil {
		log.Panicf("Failed to update transaction: %v", err)
	}

	if !updated {
		log.Infof("Transaction '%s' was already processed by another request", req.TransactionID)
	}

	return nil
}
