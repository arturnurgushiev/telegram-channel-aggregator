package service

import (
	"context"
	"errors"
	"fmt"
	"telegram-channel-aggregator/internal/repository"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type TelegramManager struct {
	repo    *repository.Sub
	userBot *telegram.Client
	tgBot   *tgbotapi.BotAPI
	ch      chan subCheckRequest
}

type subCheckRequest struct {
	channelName string
	replyToChan chan int
}

func NewTelegramManager(r *repository.Sub, userBot *telegram.Client, tgBot *tgbotapi.BotAPI) *TelegramManager {
	return &TelegramManager{
		repo:    r,
		userBot: userBot,
		tgBot:   tgBot,
		ch:      make(chan subCheckRequest, 100),
	}
}

func (telegram *TelegramManager) Start(ctx context.Context) error {
	telegram.monitorStart(ctx)
	return nil
}

func (telegram *TelegramManager) AddSubscription(userID int64, channelName string) error {
	replyChan := make(chan int, 1)
	select {
	case telegram.ch <- subCheckRequest{channelName: channelName, replyToChan: replyChan}:
	case <-time.After(5 * time.Second):
		return errors.New("service busy, try later")
	}

	select {
	case response := <-replyChan:
		if response != 0 {
			return errors.New("channel not found")
		}
	case <-time.After(5 * time.Second):
		return errors.New("timeout checking channel")
	}

	return telegram.repo.AddSubscription(userID, channelName)
}

func (telegram *TelegramManager) RemoveSubscription(userID int64, channelName string) error {
	err := telegram.repo.RemoveSubscription(userID, channelName)
	if err != nil {
		return err
	}

	subs, err := telegram.repo.GetSubscribers(channelName)
	if err != nil {
		return err
	}

	if len(subs) == 0 {

		err = telegram.repo.RemoveChannel(channelName)
		if err != nil {
			return err
		}

	}
	return nil
}

func (telegram *TelegramManager) CheckPosts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			channels, err := telegram.repo.GetAllChannels()
			if err != nil {
				continue
			}

			for i, _ := range channels {
				telegram.checkAndSend(channels[i])
			}
			time.Sleep(3 * time.Minute)
		}
	}
}

func (telegram *TelegramManager) CheckSub(ctx context.Context) {
	for {
		select {
		case req := <-telegram.ch:
			id, err := telegram.getChannelLastPostId(ctx, req.channelName)
			if err != nil {
				req.replyToChan <- 1
				continue
			}
			err = telegram.repo.AddChannel(req.channelName, id)
			if err != nil {
				req.replyToChan <- 1
				continue
			}
			req.replyToChan <- 0
		case <-ctx.Done():
			return
		}
	}
}

func (telegram *TelegramManager) getChannelLastPostId(ctx context.Context, channelName string) (int, error) {
	client := telegram.userBot
	api := client.API()
	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: channelName})
	if err != nil {
		return 0, err
	}
	users := resolved.GetChats()
	if len(users) != 1 {
		return 0, errors.New("wrong number of channel found")
	}
	channel := users[0].(*tg.Channel)

	resp, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{Peer: channel.AsInputPeer(), Limit: 100})
	if err != nil {
		return 0, errors.New("messages get History")
	}

	mod, ok := resp.AsModified()
	if !ok {
		return 0, errors.New("as modified")
	}

	msgs := mod.GetMessages()
	for _, msg := range msgs {
		return msg.GetID(), nil
	}

	return 0, errors.New("message not found")
}

func (telegram *TelegramManager) checkAndSend(channel string) {
	posts := telegram.getPosts(channel)
	if posts == nil || len(posts) == 0 {
		return
	}
	userIDs, err := telegram.repo.GetSubscribers(channel)
	if err != nil {
		return
	}

	bot := telegram.tgBot
	for _, userID := range userIDs {
		for i := range posts {
			if posts[i] == "" {
				continue
			}
			messageToUser := fmt.Sprintf("Новое сообщение из @%s\n\n%s", channel, posts[i])
			msg := tgbotapi.NewMessage(userID, messageToUser)
			bot.Send(msg)
		}
	}
}

func (telegram *TelegramManager) getPosts(channel string) []string {
	client := telegram.userBot
	api := client.API()
	resolved, err := api.ContactsResolveUsername(context.Background(), &tg.ContactsResolveUsernameRequest{Username: channel})
	if err != nil {
		return nil
	}
	users := resolved.GetChats()
	if len(users) != 1 {
		return nil
	}
	tgChannel := users[0].(*tg.Channel)

	lastPostId := telegram.repo.GetChannelLastPostId(channel)
	resp, err := api.MessagesGetHistory(context.Background(), &tg.MessagesGetHistoryRequest{
		Peer:  tgChannel.AsInputPeer(),
		Limit: 100,
		MinID: lastPostId,
	})

	if err != nil {
		return nil
	}

	mod, ok := resp.AsModified()
	if !ok {
		return nil
	}

	msgs := mod.GetMessages()
	result := make([]string, 0)

	for i, msg := range msgs {
		if i == 0 {
			telegram.repo.UpdateLastPostId(channel, msg.GetID())
		}
		m, ok := msg.(*tg.Message)
		if ok {
			result = append(result, m.Message)
		}
	}

	return result
}

func (telegram *TelegramManager) monitorStart(ctx context.Context) {
	client := telegram.userBot
	client.Run(ctx, telegram.monitor)
}

func (telegram *TelegramManager) monitor(ctx context.Context) error {
	go telegram.CheckSub(ctx)
	go telegram.CheckPosts(ctx)
	<-ctx.Done()
	return nil
}
