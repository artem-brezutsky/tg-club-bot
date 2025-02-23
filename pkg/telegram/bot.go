package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"telegram_bot/pkg/config"
	"telegram_bot/pkg/storage"
	"telegram_bot/pkg/telegram/models"
)

const (
	callbackAccept = "accept_request"
	callbackReject = "reject_request"
	callbackBanned = "fuck_off_dog"
	parseModeHTMl  = "HTML"
	maxUploadPhoto = 3
)

// Bot Основная структура приложения
type Bot struct {
	bot                 *tgbotapi.BotAPI
	userRepo            storage.UserRepository
	adminChatID         int64
	adminUserName       string
	closedGroupID       int64
	invitedGroupID      int64
	notificationGroupID int64
	messages            config.Messages
	lastMessage         map[int64]LastMessage
	statuses            map[int]string
}

type LastMessage struct {
	MessageID int
	Text      string
}

func NewBot(bot *tgbotapi.BotAPI, userRepo storage.UserRepository, cfg *config.Config) *Bot {
	return &Bot{
		bot:                 bot,
		userRepo:            userRepo,
		adminChatID:         cfg.AdminID,
		adminUserName:       cfg.AdminUserName,
		closedGroupID:       cfg.ClosedGroupID,
		invitedGroupID:      cfg.InvitesGroupID,
		notificationGroupID: cfg.NotificationGroupID,
		messages:            cfg.Messages,
		lastMessage:         make(map[int64]LastMessage),
		// todo что-то с этим придумать
		statuses: map[int]string{
			models.UserStatuses.New:      "Новий",
			models.UserStatuses.Waiting:  "В очікуванні",
			models.UserStatuses.Accepted: "Прийнято",
			models.UserStatuses.Rejected: "Відхилено",
			models.UserStatuses.Banned:   "Заблоковано",
		},
	}
}

// Start Запуск бота
func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)
	// Инициализируем канал обновлений
	updates := b.initUpdatesChannel()
	// Получаем обновления из Telegram API
	err := b.handleUpdates(updates)
	if err != nil {
		log.Panic(err)
		return err
	}

	return nil
}

// initUpdatesChannel Инициализация канала обновлений
func (b *Bot) initUpdatesChannel() tgbotapi.UpdatesChannel {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	return b.bot.GetUpdatesChan(updateConfig)
}

// handleUpdates Инкапсулирует логику для работы с обновлениями
func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) error {
	for update := range updates {
		if update.Message != nil {
			switch update.Message.Chat.ID {
			case b.closedGroupID:
				b.handleMessageFromGroup(update.Message)
				break
			case b.invitedGroupID:
				// todo Игнорируем пока что сообщения из группы с приглашениями
				b.handleMessageFromInvitedGroup(update.Message)
				break
			case b.notificationGroupID:
				// todo Игнорируем пока что сообщения из группы с уведомлениями
				b.handleMessageFromNotificationGroup(update.Message)
				break
			case b.adminChatID:
				b.handleAdminMessage(update.Message)
				break
			default:
				b.handleMessage(update.Message)
			}
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}

	return nil
}
