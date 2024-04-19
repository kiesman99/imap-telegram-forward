package main

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"gopkg.in/telebot.v3"
)

type Message struct {
	Subject string `json:"subject"`
	From    string `json:"from"`
}

type Mailclient struct {
	Client *imapclient.Client

	stopPoller chan struct{}
}

func NewMailclient(env *Environment) *Mailclient {
	client := connectImapClient(env.IMAP_SERVER_ADRESS, env.IMAP_USER, env.IMAP_PASSWORD)
	return &Mailclient{
		Client:     client,
		stopPoller: make(chan struct{}),
	}
}

func (m *Mailclient) Close() {
	m.Client.Close()
}

func connectImapClient(serverAdress string, imapUser string, imapPass string) *imapclient.Client {
	options := &imapclient.Options{}

	client, err := imapclient.DialTLS(serverAdress, options)
	if err != nil {
		panic("could not connect to imap server")
	}

	if err := client.Login(imapUser, imapPass).Wait(); err != nil {
		log.Fatalf("failed to login: %v", err)
	}
	return client
}

func (m *Mailclient) StartPoller(bot *telebot.Bot, userId int64) {
	alreadyNotifiedMap := make(map[uint32]struct{})
	slog.Info("Starting mail poller")
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				slog.Info("Checking for unread messages", "userId", userId)
				// bot.Send(&telebot.User{ID: userId}, "Checking for unread messages")

				uids, err := m.getUnreadMessageUids()
				if err != nil {
					slog.Error("could not read unread messages", err)
				}

				if uids == nil {
					continue
				}

				toNotify := filterAlreadyNotified(uids, alreadyNotifiedMap)

				if len(*toNotify) == 0 {
					continue
				}

				messages, err := m.fetchMessages(*toNotify)
				if err != nil {
					slog.Error("could not fetch messages", err)
					continue
				}

				m.SendNewUnreadNotification(userId, bot, messages)

				nums, ok := toNotify.Nums()
				if ok {
					for _, num := range nums {
						alreadyNotifiedMap[num] = struct{}{}
					}
				}
			case <-m.stopPoller:
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *Mailclient) SendNewUnreadNotification(userId int64, bot *telebot.Bot, messages []*imapclient.FetchMessageBuffer) {
	for _, msg := range messages {
		from := fmt.Sprintf("%s@%s", msg.Envelope.From[0].Mailbox, msg.Envelope.From[0].Host)
		slog.Info("New unread message", "from", from, "subject", msg.Envelope.Subject)
		bot.Send(&telebot.User{ID: userId}, fmt.Sprintf("New unread message from %s: %s", from, msg.Envelope.Subject))
	}
}

func filterAlreadyNotified(data *imap.SearchData, alreadyNotified map[uint32]struct{}) *imap.SeqSet {
	var toNotify = new(imap.SeqSet)
	all := data.AllSeqNums()
	for _, seq := range all {
		if _, ok := alreadyNotified[seq]; !ok {
			toNotify.AddNum(seq)
		}
	}
	return toNotify
}

func (m *Mailclient) FetchUnreadMessages() ([]Message, error) {
	slog.Info("Fetching unread messages")

	selectInboxCmd := m.Client.Select("INBOX", nil)

	_, err := selectInboxCmd.Wait()
	if err != nil {
		slog.Error("could not select inbox", err)
		return nil, err
	}

	data, err := m.getUnreadMessageUids()
	if err != nil {
		log.Fatalf("could not read unread messages: %v", err)
		return nil, err
	}

	if len(data.AllSeqNums()) != 0 {
		messages, err := m.fetchMessages(data.All)
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

func (m *Mailclient) getUnreadMessageUids() (*imap.SearchData, error) {
	selectCmd := m.Client.Select("INBOX", nil)
	if _, err := selectCmd.Wait(); err != nil {
		log.Fatalf("failed to select INBOX: %v", err)
	}

	data, err := m.Client.Search(&imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}, nil).Wait()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Mailclient) fetchMessages(numSet imap.NumSet) ([]*imapclient.FetchMessageBuffer, error) {
	fetchOptions := &imap.FetchOptions{Envelope: true}
	selectCmd := m.Client.Select("INBOX", nil)
	fetchCmd := m.Client.Fetch(numSet, fetchOptions)

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

func (m *Mailclient) GetMailboxes() ([]string, error) {
	mailboxes, err := m.Client.List("", "%", nil).Collect()
	if err != nil {
		log.Fatalf("failed to list mailboxes: %v", err)
		return nil, err
	}
	var boxes = make([]string, len(mailboxes))
	for _, mbox := range mailboxes {
		boxes = append(boxes, mbox.Mailbox)
	}

	return boxes, nil
}
