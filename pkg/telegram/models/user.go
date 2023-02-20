package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
)

type StringArray []string

// User Модель пользователя
type User struct {
	gorm.Model
	TelegramID int64 `gorm:"unique_index"`
	Name       string
	City       string
	Car        string
	Engine     string
	Photos     StringArray `gorm:"type:json"`
	State      int
	Status     int
}

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal StringArray value: %v", value)
	}

	return json.Unmarshal(b, &a)
}
