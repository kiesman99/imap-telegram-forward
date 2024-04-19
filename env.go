package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	IMAP_SERVER_ADRESS = "IMAP_SERVER_ADRESS"
	IMAP_USER          = "IMAP_USER"
	IMAP_PASSWORD      = "IMAP_PASSWORD"

	TELEGRAM_BOT_TOKEN = "TELEGRAM_BOT_TOKEN"
	TELEGRAM_ID        = "TELEGRAM_ID"
)

type Environment struct {
	IMAP_SERVER_ADRESS string
	IMAP_USER          string
	IMAP_PASSWORD      string
	TELEGRAM_BOT_TOKEN string
	TELEGRAM_ID        int64
}

func loadEnvironment() Environment {
	slog.Info("Loading environment variables")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	userId := os.Getenv(TELEGRAM_ID)
	parsedUserId, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		log.Fatalf("could not parse TELEGRAM_ID: %v", err)
		panic(1)
	}

	return Environment{
		IMAP_SERVER_ADRESS: os.Getenv(IMAP_SERVER_ADRESS),
		IMAP_USER:          os.Getenv(IMAP_USER),
		IMAP_PASSWORD:      os.Getenv(IMAP_PASSWORD),
		TELEGRAM_BOT_TOKEN: os.Getenv(TELEGRAM_BOT_TOKEN),
		TELEGRAM_ID:        parsedUserId,
	}
}
