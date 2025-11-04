package services

import (
	"banking-system/models"
	"banking-system/repos"
	"time"
)

//go:generate mockgen -source=transaction.go -destination=mock/transaction.go

type TransactionService interface {
	GetByUserID(userID uint, cutoffDate time.Time) ([]models.TransactionResponse, error)
}

type transactionService struct {
	transactionRepo repos.TransactionRepo
}

func NewTransactionService(transactionRepo repos.TransactionRepo) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
	}
}

func (srv *transactionService) GetByUserID(userID uint, cutoffDate time.Time) ([]models.TransactionResponse, error) {
	transactions, err := srv.transactionRepo.GetByUserID(userID, cutoffDate)
	if err != nil {
		return nil, err
	}

	var transactionResponses []models.TransactionResponse
	for _, tx := range transactions {
		transactionResponses = append(transactionResponses, models.TransactionResponse{
			UUID:          tx.UUID,
			Type:          tx.Type,
			Status:        tx.Status,
			Amount:        tx.Amount,
			PaymentMethod: string(tx.PaymentMethod),
			CreatedAt:     tx.CreatedAt,
		})
	}

	return transactionResponses, nil
}
