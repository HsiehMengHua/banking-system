package models

import (
	"banking-system/entities"
	"time"

	"github.com/google/uuid"
)

type TransactionResponse struct {
	UUID          uuid.UUID                `json:"uuid"`
	Type          entities.TransactionType `json:"type"`
	Status        entities.TransactionStatus `json:"status"`
	Amount        float64                  `json:"amount"`
	PaymentMethod string                   `json:"payment_method,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
}
