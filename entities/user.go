package entities

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(20);unique;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	Name         string `gorm:"type:varchar(100)"`

	Wallet       Wallet
	BankAccounts []BankAccount
}
