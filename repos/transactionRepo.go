package repos

import (
	"banking-system/database"
	"banking-system/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockgen -source=transactionRepo.go -destination=mock/transactionRepo.go

type TransactionRepo interface {
	Create(tx *entities.Transaction) error
	GetByUUID(uuid.UUID) (*entities.Transaction, error)
	Update(tx *entities.Transaction) error
}

type transactionRepo struct {
}

func NewTransactionRepo() TransactionRepo {
	return &transactionRepo{}
}

func (*transactionRepo) Create(transaction *entities.Transaction) error {
	if err := database.DB.Create(&transaction).Error; err != nil {
		return err
	}
	return nil
}

func (*transactionRepo) GetByUUID(uuid uuid.UUID) (*entities.Transaction, error) {
	var transaction entities.Transaction
	result := database.DB.Preload("Wallet").First(&transaction, uuid)
	return &transaction, result.Error
}

func (*transactionRepo) Update(transaction *entities.Transaction) error {
	return database.DB.Transaction(func(db *gorm.DB) error {
		if err := db.Save(transaction).Error; err != nil {
			return err
		}

		if err := db.Save(&transaction.Wallet).Error; err != nil {
			return err
		}

		return nil
	})
}
