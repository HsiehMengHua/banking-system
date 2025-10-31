package entities

import "gorm.io/gorm"

type BankAccount struct {
	gorm.Model
	UserID        uint   `gorm:"not null"`
	BankCode      string `gorm:"type:varchar(50);not null"`
	AccountNumber string `gorm:"type:varchar(100);not null"`
}
