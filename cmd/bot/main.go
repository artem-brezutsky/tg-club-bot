package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"telegram_bot/pkg/config"
	"telegram_bot/pkg/storage/postgres"
	"telegram_bot/pkg/telegram"
)

func main() {
	// Получаем конфигурацию приложения
	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Подключаемся к базе данным и создаем новый репозиторий пользователя
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	userRepo := postgres.NewUserRepository(dsn)

	// Инициализируем Telegram Bot API
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal("Не удалось инициализировать Telegram Bot API: ", err)
	}

	// Отладка приложения
	bot.Debug = cfg.Debug

	// Создаём новый экземпляр бота
	telegramBot := telegram.NewBot(bot, userRepo, cfg)
	if err = telegramBot.Start(); err != nil {
		log.Fatal(err)
	}
}
