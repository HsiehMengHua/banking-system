package entities

import "gorm.io/gorm"

type Wallet struct {
	gorm.Model
	UserID   uint    `gorm:"unique;not null"`
	Currency string  `gorm:"type:varchar(3);not null"`
	Balance  float64 `gorm:"type:numeric(18,4);not null"`
	Status   string  `gorm:"type:varchar(20);not null"`
}
