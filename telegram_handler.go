package main

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

func (t *TelegramBot) HandleUnread(c telebot.Context) error {
	return c.Send("unread")
}

func (t *TelegramBot) HandleWhoAmI(c telebot.Context) error {
	return c.Send(fmt.Sprintf("You are %s", c.Sender().Username))
}
