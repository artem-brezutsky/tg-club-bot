package storage

import "telegram_bot/pkg/telegram/models"

type UserRepository interface {
	Get(chatID int64) (*models.User, error)
	Create(chatID int64) *models.User
	Update(user *models.User) error
	Delete(charID int64) error
}
