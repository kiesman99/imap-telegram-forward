package main

import (
	"log"
	"log/slog"
	"time"

	"gopkg.in/telebot.v3"
)

type TelegramBot struct {
	Bot *telebot.Bot
}

func NewTelegramBot(env *Environment) (*TelegramBot, error) {
	pref := telebot.Settings{
		Token:  env.TELEGRAM_BOT_TOKEN,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &TelegramBot{
		Bot: bot,
	}, nil
}

func (t *TelegramBot) StartBot() {
	// t.Bot.Handle("/unread", t.HandleUnread)
	// t.Bot.Handle("/whoami", t.HandleWhoAmI)
	t.Bot.Start()
	slog.Info("Telegram bot started")
}

func (t *TelegramBot) Close() {
	slog.Info("Telegram bot stopped")
	t.Bot.Stop()
}
