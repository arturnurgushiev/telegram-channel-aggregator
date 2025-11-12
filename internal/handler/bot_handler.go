package handler

import (
	"errors"
	"fmt"
	"strings"

	"telegram-channel-aggregator/internal/service"
)

type Bot struct {
	service *service.TelegramManager
}

func NewBot(manager *service.TelegramManager) *Bot {
	return &Bot{service: manager}
}

func (bot *Bot) Handle(userID int64, msg string) (string, error) {
	commands := strings.Fields(msg)
	if len(commands) > 2 || len(commands) == 0 {
		return "", errors.New("неверная команда")
	}
	switch commands[0] {
	case "/start":
		return "Hi there!", nil
	case "/sub":
		return bot.HandleSub(userID, commands)
	case "/unsub":
		return bot.HandleUnsub(userID, commands)
	}
	return "", errors.New("неверная команда")
}

func (bot *Bot) HandleSub(userID int64, commands []string) (string, error) {
	if len(commands) != 2 {
		return "", errors.New("неверная команда")
	}
	channelName := commands[1]
	if channelName == "" || channelName[0] != '@' {
		return "", errors.New("канал должен начинаться с @")
	}
	channelName = channelName[1:]
	err := bot.service.AddSubscription(userID, channelName)
	if err != nil {
		return "", err
	}
	return "Успешная подписка", nil
}

func (bot *Bot) HandleUnsub(userID int64, commands []string) (string, error) {
	if len(commands) != 2 {
		return "", errors.New("неверная команда")
	}
	channelName := commands[1]
	if channelName == "" || channelName[0] != '@' {
		return "", errors.New("канал должен начинаться с @")
	}
	channelName = channelName[1:]
	err := bot.service.RemoveSubscription(userID, channelName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Вы были отписаны от канала %v", commands[1]), nil
}
