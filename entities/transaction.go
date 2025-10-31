package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	Deposit     TransactionType = "DEPOSIT"
	Withdrawal  TransactionType = "WITHDRAWAL"
	TransferIn  TransactionType = "TRANSFER_IN"
	TransferOut TransactionType = "TRANSFER_OUT"
)

type TransactionStatus string

const (
	Pending   TransactionStatus = "PENDING"
	Completed TransactionStatus = "COMPLETED"
	Failed    TransactionStatus = "FAILED"
	Canceled  TransactionStatus = "CANCELED"
)

type Transaction struct {
	gorm.Model

	TransactionUUID      uuid.UUID `gorm:"type:uuid;unique;not null"`
	WalletID             uint      `gorm:"not null"`
	RelatedTransactionID *uint     `gorm:"index"`

	Type   TransactionType   `gorm:"type:varchar(20);not null"`
	Status TransactionStatus `gorm:"type:varchar(20);not null"`

	Amount        float64 `gorm:"type:numeric(18,4);not null"`
	PaymentMethod string  `gorm:"type:varchar(50)"`
}
