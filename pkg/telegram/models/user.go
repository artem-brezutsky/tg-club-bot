package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
)

const (
	stateInitial = iota
	stateName
	stateCity
	stateCar
	stateEngine
	statePhoto
	stateCompleted
)

const (
	statusNew = iota
	statusWaiting
	statusAccepted
	statusRejected
	statusBanned
)

type StringArray []string

// User Модель пользователя
type User struct {
	gorm.Model
	ChatID   int64 `gorm:"unique_index"`
	UserName string
	Name     string
	City     string
	Car      string
	Engine   string
	Photos   StringArray `gorm:"type:json"`
	State    int
	Status   int
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

type statuses struct {
	New      int // user.status=0
	Waiting  int // user.status=1
	Accepted int // user.status=2
	Rejected int // user.status=3
	Banned   int // user.status=4
}

// UserStatuses Статусы пользователя
var UserStatuses = statuses{
	New:      statusNew,
	Waiting:  statusWaiting,
	Accepted: statusAccepted,
	Rejected: statusRejected,
	Banned:   statusBanned,
}

type states struct {
	Initial   int // user.state=0
	Name      int // user.state=1
	City      int // user.state=2
	Car       int // user.state=3
	Engine    int // user.state=4
	Photo     int // user.state=5
	Completed int // user.state=6
}

// UserStates Состояния пользователя
var UserStates = states{
	Initial:   stateInitial,
	Name:      stateName,
	City:      stateCity,
	Car:       stateCar,
	Engine:    stateEngine,
	Photo:     statePhoto,
	Completed: stateCompleted,
}
