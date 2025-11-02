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
	UpdateConditional(tx *entities.Transaction, expectedStatus entities.TransactionStatus) (bool, error)
	CreateTransferTransactions(transferOutTx *entities.Transaction, transferInTx *entities.Transaction) error
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

func (*transactionRepo) UpdateConditional(transaction *entities.Transaction, expectedStatus entities.TransactionStatus) (bool, error) {
	var updated bool
	err := database.DB.Transaction(func(db *gorm.DB) error {
		result := db.Model(&entities.Transaction{}).
			Where("uuid = ? AND status = ?", transaction.UUID, expectedStatus).
			Updates(map[string]interface{}{
				"status":     transaction.Status,
				"updated_at": db.NowFunc(),
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			updated = false
			return nil
		}

		if err := db.Save(&transaction.Wallet).Error; err != nil {
			return err
		}

		updated = true
		return nil
	})

	return updated, err
}

func (*transactionRepo) CreateTransferTransactions(transferOutTx *entities.Transaction, transferInTx *entities.Transaction) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(transferOutTx).Error; err != nil {
			return err
		}

		transferInTx.RelatedTransactionID = &transferOutTx.UUID
		if err := tx.Create(transferInTx).Error; err != nil {
			return err
		}

		if err := tx.Model(transferOutTx).Update("RelatedTransactionID", transferInTx.UUID).Error; err != nil {
			return err
		}

		if err := tx.Save(&transferOutTx.Wallet).Error; err != nil {
			return err
		}

		if err := tx.Save(&transferInTx.Wallet).Error; err != nil {
			return err
		}

		return nil
	})
}
