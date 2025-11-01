package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

var TransactionTypes = &struct {
	Deposit     TransactionType
	Withdrawal  TransactionType
	TransferIn  TransactionType
	TransferOut TransactionType
}{
	Deposit:     "DEPOSIT",
	Withdrawal:  "WITHDRAWAL",
	TransferIn:  "TRANSFER_IN",
	TransferOut: "TRANSFER_OUT",
}

type TransactionStatus string

var TransactionStatuses = &struct {
	Pending   TransactionStatus
	Completed TransactionStatus
	Failed    TransactionStatus
	Canceled  TransactionStatus
}{
	Pending:   "PENDING",
	Completed: "COMPLETED",
	Failed:    "FAILED",
	Canceled:  "CANCELED",
}

type Transaction struct {
	gorm.Model

	UUID                 uuid.UUID         `gorm:"type:uuid;unique;not null"`
	WalletID             uint              `gorm:"not null"`
	RelatedTransactionID *uint             `gorm:"index"`
	Type                 TransactionType   `gorm:"type:varchar(20);not null"`
	Status               TransactionStatus `gorm:"type:varchar(20);not null"`
	Amount               float64           `gorm:"type:numeric(18,4);not null"`
	PaymentMethod        string            `gorm:"type:varchar(50)"`
}
