package telegram

import (
	"bmwBot/pkg/telegram/models"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// getAdminID получаем ID администратора
func getAdminID() int64 {
	id, err := strconv.ParseInt(os.Getenv("OWNER_ACC"), 10, 64)
	if err != nil {
		log.Panic("Не удалось получить ID администратора")
	}

	return id
}

// getOwnerGroupID получаем ID группы в которую нужно принять пользователя
func getOwnerGroupID() int64 {
	id, err := strconv.ParseInt(os.Getenv("SUPERGROUP_F30_ID"), 10, 64)
	if err != nil {
		log.Panic("Не удалось получить ID закрытой группы")
	}

	return id
}

// getRandomDots получаем случайное количество точек для изменения сообщения
func getRandomDots() string {
	dots := []string{
		".",
		"..",
		"...",
	}
	rand.Seed(time.Now().Unix())

	return dots[rand.Intn(len(dots))]
}

func createMediaGroup(user *models.User, chatID int64, adminChatID int64) tgbotapi.MediaGroupConfig {
	// Формируем галерею с комментарием
	files := make([]interface{}, len(user.Photos))
	caption := fmt.Sprintf("ChatID: <b>%d</b>", chatID)
	for i, s := range user.Photos {
		if i == 0 {
			photo := tgbotapi.InputMediaPhoto{
				BaseInputMedia: tgbotapi.BaseInputMedia{
					Type:            "photo",
					Media:           tgbotapi.FileID(s),
					Caption:         caption,
					ParseMode:       parseModeHTMl,
					CaptionEntities: nil,
				}}
			files[i] = photo
		} else {
			files[i] = tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(s))
		}
	}
	cfg := tgbotapi.NewMediaGroup(
		adminChatID,
		files,
	)

	return cfg
}
