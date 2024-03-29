package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"strconv"
)

type Config struct {
	TelegramToken       string
	AdminID             int64
	AdminUserName       string
	ClosedGroupID       int64
	PostgresHost        string
	PostgresUser        string
	PostgresPassword    string
	PostgresDb          string
	Messages            Messages
	Debug               bool
	InvitesGroupID      int64
	NotificationGroupID int64
}

type Messages struct {
	Questions
	UserResponses
}

type Questions struct {
	UserName   string `mapstructure:"askUserName"`
	HearAbout  string `mapstructure:"askHearAbout"`
	UserCity   string `mapstructure:"askUserCity"`
	UserCar    string `mapstructure:"askUserCar"`
	UserEngine string `mapstructure:"askUserEngine"`
	UserPhoto  string `mapstructure:"askUserPhoto"`
}

type UserResponses struct {
	ReplyPlease     string `mapstructure:"userReplyPlease"`
	WelcomeMsg      string `mapstructure:"userWelcomeMsg"`
	AlreadyDoneMsg  string `mapstructure:"userAlreadyDoneMsg"`
	WaitingMsg      string `mapstructure:"userWaitingMsg"`
	RejectMsg       string `mapstructure:"userRejectMsg"`
	DoneRequestMsg  string `mapstructure:"userDoneRequestMsg"`
	BannedMsg       string `mapstructure:"userBannedMsg"`
	InviteMsg       string `mapstructure:"userInviteMsg"`
	GroupWelcomeMsg string `mapstructure:"userGroupWelcomeMsg"`
}

func Init() (*Config, error) {
	// Подключаем файл .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	if err := setUpViper(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := fromEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setUpViper() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("main")

	return viper.ReadInConfig()
}

func unmarshal(cfg *Config) error {
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("messages.questions", &cfg.Messages.Questions); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("messages.user_responses", &cfg.Messages.UserResponses); err != nil {
		return err
	}

	return nil
}

func fromEnv(cfg *Config) error {
	if err := viper.BindEnv("TOKEN"); err != nil {
		return err
	}
	cfg.TelegramToken = viper.GetString("token")

	if err := viper.BindEnv("ADMIN_ID"); err != nil {
		return err
	}
	cfg.AdminID, _ = strconv.ParseInt(viper.GetString("admin_id"), 10, 64)

	if err := viper.BindEnv("CLOSED_GROUP_ID"); err != nil {
		return err
	}
	cfg.ClosedGroupID, _ = strconv.ParseInt(viper.GetString("closed_group_id"), 10, 64)

	if err := viper.BindEnv("POSTGRES_HOST"); err != nil {
		return err
	}
	cfg.PostgresHost = viper.GetString("POSTGRES_HOST")

	if err := viper.BindEnv("POSTGRES_USER"); err != nil {
		return err
	}
	cfg.PostgresUser = viper.GetString("POSTGRES_USER")

	if err := viper.BindEnv("POSTGRES_PASSWORD"); err != nil {
		return err
	}
	cfg.PostgresPassword = viper.GetString("POSTGRES_PASSWORD")

	if err := viper.BindEnv("POSTGRES_DB"); err != nil {
		return err
	}
	cfg.PostgresDb = viper.GetString("POSTGRES_DB")

	if err := viper.BindEnv("TG_DEBUG"); err != nil {
		cfg.Debug = false
		return err
	}
	cfg.Debug = viper.GetBool("TG_DEBUG")

	if err := viper.BindEnv("INVITES_GROUP_ID"); err != nil {
		return err
	}
	cfg.InvitesGroupID, _ = strconv.ParseInt(viper.GetString("invites_group_id"), 10, 64)

	if err := viper.BindEnv("NOTIFICATION_GROUP_ID"); err != nil {
		return err
	}
	cfg.NotificationGroupID, _ = strconv.ParseInt(viper.GetString("notification_group_id"), 10, 64)

	if err := viper.BindEnv("ADMIN_USERNAME"); err != nil {
		return err
	}
	cfg.AdminUserName = viper.GetString("ADMIN_USERNAME")

	return nil
}

// CreatePostgresDns todo вынести куда-то в интерфейс работы с базой данных
func CreatePostgresDns(cfg *Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s",
		cfg.PostgresHost,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDb,
	)
}
