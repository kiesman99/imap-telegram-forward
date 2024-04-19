package main

import "gopkg.in/telebot.v3"

func registerDefaultCommands(b *telebot.Bot) error {

	unreadCmd := telebot.Command{
		Text:        "unread",
		Description: "Fetch unread messages",
	}

	boxesCmd := telebot.Command{
		Text:        "boxes",
		Description: "List mailboxes",
	}

	commandParams := &telebot.CommandParams{
		Commands: []telebot.Command{unreadCmd, boxesCmd},
	}

	if err := b.SetCommands(commandParams); err != nil {
		return err
	}

	return nil
}
