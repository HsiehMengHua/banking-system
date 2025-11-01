package services

import (
	"banking-system/database"
	"banking-system/entities"
	"banking-system/models"

	"github.com/google/uuid"
)

//go:generate mockgen -source=promotions.go -destination=mock/prmootions.go

type PaymentService interface {
	Deposit(req *models.DepositRequest)
}

type paymentService struct {
}

func NewPaymentService() PaymentService {
	return &paymentService{}
}

func (*paymentService) Deposit(req *models.DepositRequest) {
	var user entities.User
	database.DB.Preload("Wallet").First(&user, req.UserID)

	tx := &entities.Transaction{
		UUID:          uuid.New(),
		WalletID:      user.Wallet.ID,
		Amount:        req.Amount,
		Status:        entities.TransactionStatuses.Pending,
		Type:          entities.TransactionTypes.Deposit,
		PaymentMethod: req.PaymentMethod,
	}
	database.DB.Create(tx)

}
