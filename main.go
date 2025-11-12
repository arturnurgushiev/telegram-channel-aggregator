//go:build !auth
// +build !auth

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	conf "telegram-channel-aggregator/config"
	"telegram-channel-aggregator/internal/database"
	"telegram-channel-aggregator/internal/handler"
	"telegram-channel-aggregator/internal/repository"
	"telegram-channel-aggregator/internal/service"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gotd/td/telegram"
)

func main() {
	db := database.NewPostgres(conf.DBConnectionString)
	defer db.Close()

	repo := repository.NewSubRepository(db)

	userBot := telegram.NewClient(conf.UserApiId, conf.UserApiHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{Path: "config/session.json"},
	})

	tgBot, err := tgbotapi.NewBotAPI(conf.BotApiToken)
	if err != nil {
		panic(err)
	}

	manager := service.NewTelegramManager(repo, userBot, tgBot)

	botHandler := handler.NewBot(manager)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting UserBot monitor...")
		if err := manager.Start(ctx); err != nil {
			log.Printf("Monitor error: %v", err)
		}
	}()

	go startBotHandler(tgBot, botHandler, ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Bot stopped")
}

func startBotHandler(bot *tgbotapi.BotAPI, handler *handler.Bot, ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			log.Printf("Received command from %d: %s", update.Message.From.ID, update.Message.Text)

			response, err := handler.Handle(update.Message.From.ID, update.Message.Text)
			if err != nil {
				response = "Error: " + err.Error()
				log.Printf("Error handling command: %v", err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}

		case <-ctx.Done():
			return
		}
	}
}
