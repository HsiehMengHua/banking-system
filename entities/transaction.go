package entities

import (
	"banking-system/psp"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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
	PaymentMethod psp.PaymentMethod `gorm:"type:varchar(50)"`

	WalletID uint `gorm:"not null"`
	Wallet   *Wallet

	RelatedTransactionID *uuid.UUID   `gorm:"index"`
	RelatedTransaction   *Transaction `gorm:"foreignKey:RelatedTransactionID;references:UUID"`
}

func (tx *Transaction) Complete() error {
	if tx.Status != TransactionStatuses.Pending {
		log.Infof("Transaction '%s' already processed with status: %s", tx.UUID, tx.Status)
		return nil
	}

	tx.Status = TransactionStatuses.Completed

	switch tx.Type {
	case TransactionTypes.Deposit:
		tx.Wallet.Balance += tx.Amount
	case TransactionTypes.Withdrawal:
		// No action needed, amount already deducted during withdrawal initiation
	default:
		log.Panicf("Unknown transaction type: %s", tx.Type)
	}

	return nil
}

func (tx *Transaction) Cancel() error {
	if tx.Status != TransactionStatuses.Pending {
		log.Infof("Transaction '%s' already processed with status: %s", tx.UUID, tx.Status)
		return nil
	}

	tx.Status = TransactionStatuses.Canceled

	switch tx.Type {
	case TransactionTypes.Withdrawal:
		tx.Wallet.Balance += tx.Amount
	case TransactionTypes.Deposit:
	default:
		log.Panicf("Unknown transaction type: %s", tx.Type)
	}

	return nil
}
