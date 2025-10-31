package main

import (
	"banking-system/database"
	"banking-system/router"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	loadEnv()
	database.Connect()
}

func loadEnv() {
	env := os.Getenv("APP_ENV")
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
	r.Run() // listens on 0.0.0.0:8080 by default
}
