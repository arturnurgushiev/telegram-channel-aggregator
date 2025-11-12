//go:build auth
// +build auth

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	conf "telegram-channel-aggregator/config"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

var (
	apiID    = conf.UserApiId
	apiHash  = conf.UserApiHash
	phone    = conf.PhoneNumber
	password = conf.Password
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Ошибка получения пути: %v", err)
	}
	path := filepath.Join(dir, "config", "session.json")

	ctx := context.Background()

	client := telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{Path: path},
	})

	var getCode = func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		fmt.Printf("Введите код ")
		reader := bufio.NewReader(os.Stdin)
		code, _ := reader.ReadString('\n')
		return strings.TrimSpace(code), nil
	}

	var authhelp auth.CodeAuthenticatorFunc = getCode

	err = client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("Начинаю авторизацию...")

		flow := auth.NewFlow(auth.Constant(phone, password, authhelp), auth.SendCodeOptions{})

		if err := flow.Run(ctx, client.Auth()); err != nil {
			return fmt.Errorf("ошибка авторизации: %w", err)
		}

		fmt.Println("Успешно вошли в Telegram!")

		return nil
	})

	if err != nil {
		log.Fatalf("Ошибка: %v", err)
	}
}
