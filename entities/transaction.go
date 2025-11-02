package entities

import (
	"time"

	"github.com/google/uuid"
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
	CreatedAt time.Time
	UpdatedAt time.Time

	UUID          uuid.UUID         `gorm:"type:uuid;primaryKey;not null"`
	Type          TransactionType   `gorm:"type:varchar(20);not null"`
	Status        TransactionStatus `gorm:"type:varchar(20);not null"`
	Amount        float64           `gorm:"type:numeric(18,4);not null"`
	PaymentMethod string            `gorm:"type:varchar(50)"`

	WalletID uint `gorm:"not null"`
	Wallet   *Wallet

	RelatedTransactionID *uuid.UUID   `gorm:"index"`
	RelatedTransaction   *Transaction `gorm:"foreignKey:RelatedTransactionID;references:UUID"`
}
