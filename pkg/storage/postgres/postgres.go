package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"telegram_bot/pkg/telegram/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(dsn string) *UserRepository {
	// Пытаемся установить соединение с базой данных
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных: ", err)
	}

	// Выполняем миграции
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Не удалось выполнить миграцию: ", err)
	}

	return &UserRepository{db: db}
}

func (userRepo *UserRepository) Get(chatID int64) (*models.User, error) {
	var user *models.User
	result := userRepo.db.Where(&models.User{ChatID: chatID}).First(&user)

	return user, result.Error
}

func (userRepo *UserRepository) Create(chatID int64) *models.User {
	user := models.User{
		ChatID: chatID,
		Status: models.UserStatuses.New,
		State:  models.UserStates.Initial,
	}

	if err := userRepo.db.Create(&user).Error; err != nil {
		log.Fatalf("Не удалось создать пользователя с ChatID: %d", chatID)
	}

	return &user
}

func (userRepo *UserRepository) Update(user *models.User) error {
	if err := userRepo.db.Save(user).Error; err != nil {
		log.Printf("Ошибка обнвления пользователя с ChatID: %d", user.ChatID)
	}

	return nil
}

func (userRepo *UserRepository) Delete(chatID int64) error {
	return nil
}
