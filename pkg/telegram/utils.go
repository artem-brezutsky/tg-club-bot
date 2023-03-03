package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
	"telegram_bot/pkg/telegram/models"
)

// getRandomDots получаем случайное количество точек для изменения сообщения
func getRandomDots(str string) string {

	count := strings.Count(str, ".")
	switch count {
	case 1:
		return str + "."
	case 2:
		return str + "."
	case 3:
		return strings.TrimRight(str, ".")
	default:
		return str + "."
	}
}

// createMediaGroup Формируем медиа группу из фото пользователя
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

// buildUserDataMessage Формируем сообщение с данными пользователя
func (b *Bot) buildUserDataMessage(user *models.User) string {
	return fmt.Sprintf(
		"Нова заявка на вступ:\n\n"+
			"Ім'я: %s\n"+
			"Місто: %s\n"+
			"Автомобіль: %s\n"+
			"Двигун: %s\n"+
			"Дізнались з: %s\n"+
			"ChatID: %d\n",
		user.Name,
		user.City,
		user.Car,
		user.Engine,
		user.HearAbout,
		user.ChatID)
}

// escapeString Удаляем из никнейма пользователя запрещенные символи
func escapeString(s string) string {
	var escaped strings.Builder
	for _, r := range s {
		switch r {
		case '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!':
			escaped.WriteRune('\\')
		}
		escaped.WriteRune(r)
	}
	return escaped.String()
}

func addLinkToUser(userID int64) string {
	userLink := fmt.Sprintf("<a href=\"tg://user?id=%s\">%s</a>", strconv.FormatInt(userID, 10), "User Link Text")
	return userLink
}
