package config

import (
	"github.com/spf13/viper"
	"os"
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
	// Для Railway не нужен файл .env
	//if err := godotenv.Load(); err != nil {
	//	fmt.Println("No .env file found or error loading .env file")
	//}

	if err := setUpViper(); err != nil {
		return nil, err
	}

	var cfg Config

	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.TelegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.AdminID, _ = strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	cfg.ClosedGroupID, _ = strconv.ParseInt(os.Getenv("CLOSED_GROUP_ID"), 10, 64)
	cfg.PostgresHost = os.Getenv("POSTGRES_HOST")
	cfg.PostgresUser = os.Getenv("POSTGRES_USER")
	cfg.PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	cfg.PostgresDb = os.Getenv("POSTGRES_DB")
	cfg.Debug, _ = strconv.ParseBool(os.Getenv("TG_DEBUG"))
	cfg.InvitesGroupID, _ = strconv.ParseInt(os.Getenv("INVITES_GROUP_ID"), 10, 64)
	cfg.NotificationGroupID, _ = strconv.ParseInt(os.Getenv("NOTIFICATION_GROUP_ID"), 10, 64)
	cfg.AdminUserName = os.Getenv("ADMIN_USERNAME")

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
