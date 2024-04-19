package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	tele "gopkg.in/telebot.v3"
)

type Message struct {
	Subject string `json:"subject"`
	From    string `json:"from"`
}

func main() {
	env := loadEnvironment()
	client := connectImapClient(&env)
	defer client.Close()

	pref := tele.Settings{
		Token:  env.TELEGRAM_BOT_TOKEN,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer bot.Stop()
	defer bot.Close()

	bot.Handle("/mails", func(c tele.Context) error {
		first, err := firstMessage(client)
		if err != nil {
			return c.Send("Oh oh. Could not fetch mails")
		}
		return c.Send(first)
	})

	bot.Handle("/boxes", func(c tele.Context) error {
		boxes := getMailboxes(client)
		return c.Send(boxes)
	})

	bot.Handle("/unread", func(c tele.Context) error {
		messages, err := fetchUnreadMessages(client)

		if err != nil {
			return c.Send("Could not fetch unread messages")
		}

		if len(messages) == 0 {
			return c.Send("No unread messages")
		}

		ms, err := json.Marshal(messages)
		if err != nil {
			return c.Send("Could not marshal messages")
		}

		return c.Send(string(ms))
	})

	bot.Handle("/start", func(c tele.Context) error {
		err = registerDefaultCommands(bot)
		if err != nil {
			log.Fatalf("could not register default commands: %v", err)
		}
		bot.SetMenuButton(c.Sender(), &tele.MenuButton{
			Type: tele.MenuButtonCommands,
		})
		slog.Info(fmt.Sprintf("User %s", c.Sender().FirstName))

		return c.Send(fmt.Sprintf("Hello %s", c.Sender().FirstName))
	})

	bot.Start()

	if err := client.Logout().Wait(); err != nil {
		log.Fatalf("failed to logout: %v", err)
	}
}

func fetchUnreadMessages(client *imapclient.Client) ([]Message, error) {
	if _, err := client.Select("INBOX", nil).Wait(); err != nil {
		log.Fatalf("could not select inbox: %v", err)
	}

	data, err := getUnreadMessageUids(client)
	if err != nil {
		log.Fatalf("could not read unread messages: %v", err)
		return nil, err
	}

	if len(data.AllSeqNums()) != 0 {
		messages, err := fetchMessages(client, data.All)
		if err != nil {
			log.Fatalf("could not fetch messages: %v", err)
			return nil, err
		}

		var msgs = make([]Message, len(messages))

		for _, msg := range messages {
			from := fmt.Sprintf("%s@%s", msg.Envelope.From[0].Mailbox, msg.Envelope.From[0].Host)

			msgs = append(msgs, Message{
				Subject: msg.Envelope.Subject,
				From:    from,
			})
		}

		return msgs, nil
	}

	return []Message{}, nil
}

func connectImapClient(env *Environment) *imapclient.Client {
	options := &imapclient.Options{}

	client, err := imapclient.DialTLS(env.IMAP_SERVER_ADRESS, options)
	if err != nil {
		panic("could not connect to imap server")
	}

	if err := client.Login(env.IMAP_USER, env.IMAP_PASSWORD).Wait(); err != nil {
		log.Fatalf("failed to login: %v", err)
	}
	return client
}

func firstMessage(client *imapclient.Client) (string, error) {
	selectedMbox, err := client.Select("INBOX", nil).Wait()
	if err != nil {
		log.Fatalf("failed to select INBOX: %v", err)
	}
	log.Printf("INBOX contains %v messages", selectedMbox.NumMessages)

	if selectedMbox.NumMessages > 0 {
		seqSet := imap.SeqSetNum(1)
		fetchOptions := &imap.FetchOptions{Envelope: true}
		messages, err := client.Fetch(seqSet, fetchOptions).Collect()
		if err != nil {
			log.Fatalf("failed to fetch first message in INBOX: %v", err)
			return "", errors.New("could not fetch messagess")
		}
		log.Printf("subject of first message in INBOX: %v", messages[0].Envelope.Subject)

		return messages[0].Envelope.Subject, nil
	}

	return "", errors.New("no Messages")
}

func getMailboxes(client *imapclient.Client) string {
	mailboxes, err := client.List("", "%", nil).Collect()
	if err != nil {
		log.Fatalf("failed to list mailboxes: %v", err)
	}
	log.Printf("Found %v mailboxes", len(mailboxes))
	var boxes = make([]string, len(mailboxes))
	for _, mbox := range mailboxes {
		boxes = append(boxes, mbox.Mailbox)
		// log.Printf(" - %v", mbox.Mailbox)
	}

	return strings.Join(boxes, "\n")
}

func getUnreadMessageUids(client *imapclient.Client) (*imap.SearchData, error) {
	selectCmd := client.Select("INBOX", nil)
	if _, err := selectCmd.Wait(); err != nil {
		log.Fatalf("failed to select INBOX: %v", err)
	}

	data, err := client.Search(&imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}, nil).Wait()

	if err != nil {
		return nil, err
	}

	// fmt.Printf("Found %v unread messages\n", data.Count)
	// fmt.Printf("%v\n", len(data.All()))

	return data, nil
}

func fetchMessages(c *imapclient.Client, numSet imap.NumSet) ([]*imapclient.FetchMessageBuffer, error) {
	fetchOptions := &imap.FetchOptions{Envelope: true}
	selectCmd := c.Select("INBOX", nil)
	fetchCmd := c.Fetch(numSet, fetchOptions)

	if _, err := selectCmd.Wait(); err != nil {
		log.Fatalf("failed to select INBOX: %v", err)
	}

	messages, err := fetchCmd.Collect()
	if err != nil {
		log.Fatalf("failed to fetch message: %v", err)
		return nil, err
	}

	return messages, nil
}
