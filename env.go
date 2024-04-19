package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	IMAP_SERVER_ADRESS = "IMAP_SERVER_ADRESS"
	IMAP_USER          = "IMAP_USER"
	IMAP_PASSWORD      = "IMAP_PASSWORD"

	TELEGRAM_BOT_TOKEN = "TELEGRAM_BOT_TOKEN"
)

type Environment struct {
	IMAP_SERVER_ADRESS string
	IMAP_USER          string
	IMAP_PASSWORD      string
	TELEGRAM_BOT_TOKEN string
}

func loadEnvironment() Environment {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return Environment{
		IMAP_SERVER_ADRESS: os.Getenv(IMAP_SERVER_ADRESS),
		IMAP_USER:          os.Getenv(IMAP_USER),
		IMAP_PASSWORD:      os.Getenv(IMAP_PASSWORD),
		TELEGRAM_BOT_TOKEN: os.Getenv(TELEGRAM_BOT_TOKEN),
	}
}
