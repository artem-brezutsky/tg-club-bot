package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"telegram_bot/pkg/config"
	"telegram_bot/pkg/telegram"
	"telegram_bot/pkg/telegram/models"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Подключаемся к базе данным
	dsn := config.CreateDns(cfg)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Выполняем миграции
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Не удалось выполнить миграцию: ", err)
	}

	// Инициализируем Telegram Bot API
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal("Не удалось инициализировать Telegram Bot API: ", err)
	}

	//bot.Debug = true

	// Создаём новый экземпляр бота
	telegramBot := telegram.NewBot(bot, db, cfg)
	if err := telegramBot.Start(); err != nil {
		log.Fatal(err)
	}
}
