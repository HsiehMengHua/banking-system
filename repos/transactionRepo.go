package repos

import (
	"banking-system/database"
	"banking-system/entities"
)

//go:generate mockgen -source=transactionRepo.go -destination=mock/transactionRepo.go

type TransactionRepo interface {
	Create(tx *entities.Transaction) error
}

type transactionRepo struct {
}

func NewTransactionRepo() TransactionRepo {
	return &transactionRepo{}
}

func (*transactionRepo) Create(transaction *entities.Transaction) error {
	database.DB.Create(&transaction)
	return nil
}
