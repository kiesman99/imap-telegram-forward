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
		slog.Info("Could not load .env file")
	}

	userId := requireEnvVar(TELEGRAM_ID)
	parsedUserId, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		log.Fatalf("could not parse TELEGRAM_ID: %v", err)
		panic(1)
	}

	imapServerAdress := requireEnvVar(IMAP_SERVER_ADRESS)
	imapUser := requireEnvVar(IMAP_USER)
	imapPassword := requireEnvVar(IMAP_PASSWORD)
	telegramBotToken := requireEnvVar(TELEGRAM_BOT_TOKEN)

	return Environment{
		IMAP_SERVER_ADRESS: imapServerAdress,
		IMAP_USER:          imapUser,
		IMAP_PASSWORD:      imapPassword,
		TELEGRAM_BOT_TOKEN: telegramBotToken,
		TELEGRAM_ID:        parsedUserId,
	}
}

func requireEnvVar(key string) string {
	value, existing := os.LookupEnv(key)
	if !existing {
		log.Fatalf("could not find %s in environment", key)
		panic(1)
	}
	return value
}
