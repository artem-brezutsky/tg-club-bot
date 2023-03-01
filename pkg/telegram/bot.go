package telegram

import (
	"bmwBot/pkg/config"
	"bmwBot/pkg/telegram/models"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"log"
)

const (
	stateInitial   = 0
	stateName      = 1
	stateCity      = 2
	stateCar       = 3
	stateEngine    = 4
	statePhoto     = 5
	stateCompleted = 6
)

const (
	statusNew      = 0
	statusWaiting  = 1
	statusAccepted = 2
	statusRejected = 3
	statusBanned   = 4
)

const maxUploadPhoto = 3

const (
	callbackAccept = "accept_request"
	callbackReject = "reject_request"
	callbackBanned = "fuck_off_dog"
)

const parseModeHTMl = "HTML"

/** Вопросы */
//const (
//	askUserName   = "Як тебе звати?"
//	askUserCity   = "З якого ти міста?"
//	askUserCar    = "Яке в тебе авто?"
//	askUserEngine = "Який двигун?"
//	askUserPhoto  = "Надійшли фото автомобіля, щоб було видно державний номер авто.\nЯкщо вважаєш за необхідне приховати номерний знак - це твоє право, але ми повинні розуміти, що ти з України та тобі можна довіряти."
//)

// todo что то сделать с этими ссылками в статичных текстах
//const (
//	userReplyPlease    = "Будь ласка, дай відповідь на питання вище!"
//	userWelcomeMsg     = "Привіт, зараз я поставлю тобі кілька запитань!"
//	userAlreadyDoneMsg = "Ваша заявку вже було розглянуто, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyiv"
//	userWaitingMsg     = "Наразі ваша заявка на розгляді, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyiv"
//	userRejectMsg      = "Вашу заявку було відхилено, для інформації зв'яжіться з адміністрацією. @fclubkyivadmin"
//	userDoneRequestMsg = "Дякуємо, найближчим часом ви отримаєте посилання на чат. Якщо протягом тривалого часу ви не отримали посилання - зв'яжіться з адміністрацією - @fclubkyivadmin."
//	userBannedMsg      = "Ваша заявка була заблокована, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyivadmin"
//	userInviteMsg      = "Привіт!\nТвої відповіді стосовно вступу в <b>F-club Kyiv</b> були оброблені нашою командою. Ознайомся з простими умовами спілкування в нашому клубі та приєднуйся до нас! \n\n1. Поважай інших учасників. Нецензурна лайка, цькування, використання непристойних стікерів - заборонено(але якщо це в тему, то всі розуміють😂)\n2. Не влаштовуємо «Барахолку»! Ти можешь запропонувати, якщо в тебе є щось корисне для автомобіля, чи будь що, але не треба про це писати кожного дня і робити рекламні оголошення. \n3. Якщо ти хочеш запропонувати свої послугу(сто, детейлінг, автомийки, итд) - повідом про це адміністрації і зробіть гарне оголошення разом - це все безкоштовно !! \n 4. Ми розуміємо, що зараз без цього ніяк, але маємо про це попросити - якомога менше суперечок стосовно політики. Ми всі підтримуємо Україну і не шукаємо зради!\n 5. Стосовно використання GIF , ми не проти цього, але не треба постити дуже багато, один за одним! \n 6. Май повагу до інших власників автомобілів, не у кожного така гарна машина, як в тебе!  \n\nМаєш бажання отримати клубний стікер на авто чи номерну рамку - відпиши на це повідомлення\U0001FAE1\n\nТримай посилання, для вступу в чат!\n     P.s.Не забудь привітатися з нових товаришами, та розповісти який в тебе автомобіль!\n\n\n\nДонати для розвитку!(За бажанням) \n\nF-Club Kyiv \n\n🎯Ціль: 100 000.00 ₴\n\n🔗Посилання на банку\nhttps://send.monobank.ua/jar/S87zLF6xL\n\n💳Номер картки банки\n5375 4112 0304 9692"
//)

// Bot Основная структура приложения
type Bot struct {
	bot           *tgbotapi.BotAPI
	db            *gorm.DB
	adminChatID   int64
	closedGroupID int64
	statuses      map[int]string
	messages      config.Messages
}

var lastBotMessageIDInChat map[int64]int

func NewBot(bot *tgbotapi.BotAPI, db *gorm.DB, cfg *config.Config) *Bot {
	return &Bot{
		bot:           bot,
		db:            db,
		adminChatID:   cfg.AdminID,
		closedGroupID: cfg.ClosedGroupID,
		statuses: map[int]string{
			statusNew:      "Новий",
			statusWaiting:  "В очікуванні",
			statusAccepted: "Прийнято",
			statusRejected: "Відхилено",
			statusBanned:   "Заблоковано",
		},
		messages: cfg.Messages,
	}
}

// Start запуск бота
func (b *Bot) Start() error {
	log.Printf("Авторизация в аккаунте: %s", b.bot.Self.UserName)
	lastBotMessageIDInChat = make(map[int64]int)
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

// initUpdatesChannel инициализация канала обновлений
func (b *Bot) initUpdatesChannel() tgbotapi.UpdatesChannel {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	return b.bot.GetUpdatesChan(updateConfig)
}

// handleUpdates инкапсулирет логику для работы с обновлениями
func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) error {
	for update := range updates {
		if update.Message != nil {
			if update.Message.Chat.ID == b.closedGroupID {
				continue
			}

			if update.Message.Chat.ID == b.adminChatID {
				b.handleAdminMessage(update.Message)
			} else {
				b.handleMessage(update.Message)
			}
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}

	return nil
}

// todo разделить вероятно метод создать и получить пользователя
// getUser Получение пользователя из базы данных по его ChatID, если пользователя нет - создаёт его
func getUser(db *gorm.DB, telegramID int64) (*models.User, error) {
	var user models.User
	if err := db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				TelegramID: telegramID,
				State:      stateInitial,
				Status:     statusNew,
			}
			if err := db.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &user, nil
}

// updateUser обновление данных пользователя в базе данных
func updateUser(db *gorm.DB, user *models.User) {
	if err := db.Save(user).Error; err != nil {
		log.Printf("Error updating user: %s", err)
	}
}
