package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_DSN")

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Panicf("Failed to connect to database: %v", err)
	}

	// AutoMigrate runs the database migrations, creating or updating the table based on the Item struct.
	// err = DB.AutoMigrate(&Item{})
	// if err != nil {
	// 	log.Panicf("Failed to run database migration: %v", err)
	// }

	log.Print("Database connection established and migration complete!")
}
