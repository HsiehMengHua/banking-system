package main

import (
	"banking-system/database"
	"banking-system/router"
	"os"

	"github.com/joho/godotenv"
)

var env string

func init() {
	loadEnv()
	database.Connect()
}

func loadEnv() {
	env = os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load() // The Original .env
}

func main() {
	r := router.Setup()

	if env == "development" && os.Getenv("RUN_HTTPS") == "y" {
		r.RunTLS(":8444", os.Getenv("CERT_FILE"), os.Getenv("CERT_KEY"))
	} else {
		r.Run(os.Getenv("APP_PORT"))
	}
}
