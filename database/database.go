package database

import (
	"banking-system/entities"
	"os"

	log "github.com/sirupsen/logrus"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_DSN")

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Warnf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&entities.User{}, &entities.Wallet{}, &entities.Transaction{}, &entities.BankAccount{})
	if err != nil {
		log.Warnf("Failed to run database migration: %v", err)
	}

	log.Info("Database connection established and migration complete!")
}
