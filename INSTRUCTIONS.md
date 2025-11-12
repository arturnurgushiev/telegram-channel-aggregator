# Инструкция по запуску Telegram Channel Aggregator

## Предварительные требования

### 1. Получение API credentials для Telegram
1. На сайте [my.telegram.org](https://my.telegram.org) получите:
   - API ID
   - API Hash
2. Напишите [@BotFather](https://t.me/BotFather) в Telegram, получите
   - api_token для бота

### 2. Клонируйте репозиторий

### 3. Настройка базы данных PostgreSQL

1. Создание базы данных
```bash
sudo -u postgres psql -c "CREATE DATABASE telegram_aggregator;"
sudo -u postgres psql -c "CREATE USER aggregator_user WITH PASSWORD 'your_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE telegram_aggregator TO aggregator_user;"
```
2. Выполнение миграций
```bash
psql -U aggregator_user -d telegram_aggregator -f migrations/create_tables.sql
```
### 4. Заполните config.go
```go
package config

var (
    BotApiToken        = "YOUR_BOT_API_TOKEN_HERE"
    UserApiId          = YOUR_API_ID_HERE
    UserApiHash        = "YOUR_API_HASH_HERE"
    PhoneNumber        = "YOUR_PHONE_NUMBER"
    Password           = ""
    DBConnectionString = "postgres://aggregator_user:your_password@localhost/telegram_aggregator?sslmode=disable"
)
```
### 5. Пройдите авторизацию
Нужно будет ввести код из СМС и не забыть выше в config.go ввести пароль от двухфакторной аутентификации
```bash
go mod tidy
go run auth_tool.go
```

### 6. Сборка и запуск основного приложения
```
go build -o aggregator main.go
./aggregator
```