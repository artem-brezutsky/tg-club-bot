package main

import (
	"bmwBot/pkg/telegram"
	"bmwBot/pkg/telegram/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

func main() {
	// Подключаем файл .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	// Подключаемся к базе данным
	db, err := gorm.Open(mysql.Open(os.Getenv("DSN")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Выполняем миграции
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Не удалось выполнить миграцию: ", err)
	}

	// Инициализируем Telegram Bot API
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("Не удалось инициализировать Telegram Bot API: ", err)
	}

	//bot.Debug = true

	// Создаём новый экземпляр бота
	telegramBot := telegram.NewBot(bot, db)
	if err := telegramBot.Start(); err != nil {
		log.Fatal(err)
	}
}
