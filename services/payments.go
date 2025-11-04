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
	MIN_DEPOSIT_AMOUNT  = 1.00
	MAX_DEPOSIT_AMOUNT  = 100000.00
	MIN_TRANSFER_AMOUNT = 1.00
	MAX_TRANSFER_AMOUNT = 100000.00
	confirmCallbackPath = "/api/v1/payments/confirm"
	cancelCallbackPath  = "/api/v1/payments/cancel"
)

type PaymentService interface {
	Deposit(req *models.DepositRequest, baseURL string) (redirectUrl string, err error)
	Withdraw(req *models.WithdrawRequest) error
	Transfer(req *models.TransferRequest) error
	Confirm(req *psp.ConfirmRequest) error
	Cancel(req *psp.CancelRequest) error
}

type paymentService struct {
	userRepo        repos.UserRepo
	transactionRepo repos.TransactionRepo
	pspFactory      psp.PSPFactory
}

func NewPaymentService(userRepo repos.UserRepo, transactionRepo repos.TransactionRepo, pspFactory psp.PSPFactory) PaymentService {
	return &paymentService{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		pspFactory:      pspFactory,
	}
}

func (srv *paymentService) Deposit(req *models.DepositRequest, baseURL string) (redirectUrl string, err error) {
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

	provider := srv.pspFactory.NewPaymentServiceProvider(req.PaymentMethod)
	res, err := provider.PayIn(&psp.PayInRequest{
		TransactionID:      tx.UUID.String(),
		Amount:             tx.Amount,
		ConfirmCallbackURL: fmt.Sprintf("%s%s", baseURL, confirmCallbackPath),
		CancelCallbackURL:  fmt.Sprintf("%s%s", baseURL, cancelCallbackPath),
	})
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
		UUID:          req.UUID,
		WalletID:      user.Wallet.ID,
		Amount:        req.Amount,
		Status:        entities.TransactionStatuses.Pending,
		Type:          entities.TransactionTypes.Withdrawal,
		PaymentMethod: req.PaymentMethod,
		Wallet:        &user.Wallet,
	}

	if err := srv.transactionRepo.Create(tx); err != nil {
		log.Panicf("Failed to create transaction: %v", err)
	}

	user.Wallet.Balance -= req.Amount
	if err := srv.userRepo.UpdateWallet(user); err != nil {
		log.Panicf("Failed to update wallet for user ID %d: %v", req.UserID, err)
	}

	provider := srv.pspFactory.NewPaymentServiceProvider(req.PaymentMethod)
	_, err = provider.PayOut()
	if err != nil {
		log.Panicf("Payment service provider error: %v", err)
	}

	return nil
}

func (srv *paymentService) Transfer(req *models.TransferRequest) error {
	if req.Amount < MIN_TRANSFER_AMOUNT {
		return fmt.Errorf("transfer amount %.2f is below minimum allowed amount %.2f", req.Amount, MIN_TRANSFER_AMOUNT)
	}

	if req.Amount > MAX_TRANSFER_AMOUNT {
		return fmt.Errorf("transfer amount %.2f exceeds maximum allowed amount %.2f", req.Amount, MAX_TRANSFER_AMOUNT)
	}

	if req.SenderUserID == req.RecipientUserID {
		return fmt.Errorf("cannot transfer to the same user")
	}

	sender, err := srv.userRepo.Get(req.SenderUserID)
	if err != nil {
		log.Panicf("Failed to get sender user: %v", err)
	}

	recipient, err := srv.userRepo.Get(req.RecipientUserID)
	if err != nil {
		log.Panicf("Failed to get recipient user: %v", err)
	}

	if sender.Wallet.Balance < req.Amount {
		return fmt.Errorf("insufficient balance: current balance %.2f, requested amount %.2f", sender.Wallet.Balance, req.Amount)
	}

	transferOutTx := &entities.Transaction{
		UUID:     req.UUID,
		WalletID: sender.Wallet.ID,
		Amount:   req.Amount,
		Status:   entities.TransactionStatuses.Completed,
		Type:     entities.TransactionTypes.TransferOut,
		Wallet:   &sender.Wallet,
	}

	transferInTx := &entities.Transaction{
		UUID:     uuid.New(),
		WalletID: recipient.Wallet.ID,
		Amount:   req.Amount,
		Status:   entities.TransactionStatuses.Completed,
		Type:     entities.TransactionTypes.TransferIn,
		Wallet:   &recipient.Wallet,
	}

	sender.Wallet.Balance -= req.Amount
	recipient.Wallet.Balance += req.Amount

	if err := srv.transactionRepo.CreateTransferTransactions(transferOutTx, transferInTx); err != nil {
		log.Panicf("Failed to create transfer: %v", err)
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
	case entities.TransactionTypes.Withdrawal:
		// No action needed, amount already deducted during withdrawal initiation
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
