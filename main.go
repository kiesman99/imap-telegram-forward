package main

import (
	"log"
	"log/slog"
)

func main() {
	slog.Info("Starting application")
	env := loadEnvironment()

	mailClient := NewMailclient(&env)
	defer mailClient.Close()

	telegramBot, err := NewTelegramBot(&env)
	if err != nil {
		log.Fatalf("could not create telegram bot: %v", err)
		panic(1)
	}
	defer telegramBot.Close()

	mailClient.StartPoller(telegramBot.Bot, env.TELEGRAM_ID)
	telegramBot.StartBot()
}
