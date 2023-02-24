package main

import (
	"bmwBot/pkg/config"
	"bmwBot/pkg/telegram"
	"bmwBot/pkg/telegram/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Подключаемся к базе данным
	db, err := gorm.Open(mysql.Open(cfg.DNS), &gorm.Config{})
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
