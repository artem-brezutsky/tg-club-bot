package telegram

import (
	"bmwBot/pkg/telegram/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/forPelevin/gomoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"strings"
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

const (
	callbackAccept = "accept_request"
	callbackReject = "reject_request"
	callbackBanned = "fuck_off_dog"
)

const parseModeHTMl = "HTML"

/** Вопросы */
const (
	askUserName   = "Як тебе звати?"
	askUserCity   = "З якого ти міста?"
	askUserCar    = "Яке в тебе авто?"
	askUserEngine = "Який двигун?"
	askUserPhoto  = "Надійшли фото автомобіля, щоб було видно державний номер авто - після натисни «ГОТОВО»\nЯкщо вважаєш за необхідне приховати номерний знак - це твоє право, але ми повинні розуміти, що ти з України та тобі можна довіряти."
)

// todo что то сделать с этими ссылками в статичных текстах
const (
	userReplyPlease    = "Будь ласка, дай відповідь на питання вище!"
	userWelcomeMsg     = "Привіт, зараз я поставлю тобі кілька запитань!"
	userAlreadyDoneMsg = "Ваша заявку вже було розглянуто, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyiv"
	userWaitingMsg     = "Наразі ваша заявка на розгляді, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyiv"
	userRejectMsg      = "Вашу заявку було відхилено, для інформації зв'яжіться з адміністрацією. @fclubkyiv"
	userDoneRequestMsg = "Дякуємо, найближчим часом ви отримаєте посилання на чат. Якщо протягом тривалого часу ви не отримали посилання - зв'яжіться з адміністрацією - @fclubkyiv."
	userBannedMsg      = "Ваша заявка була заблокована, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyiv"
)

// Bot Основная структура приложения
type Bot struct {
	bot *tgbotapi.BotAPI
	db  *gorm.DB
}

func NewBot(bot *tgbotapi.BotAPI, db *gorm.DB) *Bot {
	return &Bot{
		bot: bot,
		db:  db,
	}
}

// Start запуск бота
func (b *Bot) Start() error {
	log.Printf("Авторизация в аккаунте: %s", b.bot.Self.UserName)

	updates := b.initUpdatesChannel()
	// test git branches
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
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}

	return nil
}

// handleMessage обработка сообщений
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	ownerGroupID, err := b.getOwnerGroupID()
	if err != nil {
		log.Panic(err)
	}

	// todo сделать обработку сообщений из группы
	// Если сообщение из чата группы, пропускаем его
	if message.Chat.ID == ownerGroupID {
		return
	}

	adminID, err := b.getAdminID()
	if err != nil {
		log.Panic(err)
	}

	// Кнопка готово
	var doneButton = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Готово👌"),
		),
	)

	// Кнопки для ответа администратора
	var requestButtons = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Прийняти", "accept_request"),
			tgbotapi.NewInlineKeyboardButtonData("Відхилити", "reject_request"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Заблокувати орка", "fuck_off_dog"),
		),
	)

	user, err := getUser(b.db, message.Chat.ID)
	if err != nil {
		log.Panic("Ошибка получения пользователя: ", err)
	}

	// Проверяем статус пользователя
	switch user.Status {
	case statusAccepted:
		msg := tgbotapi.NewMessage(message.Chat.ID, userAlreadyDoneMsg)
		msg.ParseMode = parseModeHTMl
		b.bot.Send(msg)
	case statusRejected:
		msg := tgbotapi.NewMessage(message.Chat.ID, userRejectMsg)
		msg.ParseMode = parseModeHTMl
		b.bot.Send(msg)
	case statusBanned:
		msg := tgbotapi.NewMessage(message.Chat.ID, userBannedMsg)
		msg.ParseMode = parseModeHTMl
		b.bot.Send(msg)
	case statusWaiting:
		msg := tgbotapi.NewMessage(message.Chat.ID, userWaitingMsg)
		msg.ParseMode = parseModeHTMl
		b.bot.Send(msg)
	case statusNew:
		switch user.State {
		case stateInitial:
			msg := tgbotapi.NewMessage(message.Chat.ID, "")
			msg.Text = userWelcomeMsg
			b.bot.Send(msg)
			msg.Text = askUserName
			b.bot.Send(msg)
			user.State = stateName
			updateUser(b.db, user)
		case stateName:
			message.Text = gomoji.RemoveEmojis(message.Text)
			userMsg := tgbotapi.NewMessage(message.Chat.ID, "")
			if message.Text == "" {
				// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
				userMsg.Text = userReplyPlease
				b.bot.Send(userMsg)
				return
			}

			// Отправляем пользователю следующий вопрос
			user.Name = message.Text
			userMsg.Text = askUserCity
			b.bot.Send(userMsg)

			// Обновляем состояние пользователя
			user.State = stateCity
			updateUser(b.db, user)
		case stateCity:
			message.Text = gomoji.RemoveEmojis(message.Text)
			userMsg := tgbotapi.NewMessage(message.Chat.ID, "")
			if message.Text == "" {
				// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
				userMsg.Text = userReplyPlease
				b.bot.Send(userMsg)
				return
			}

			// Отправляем пользователю следующий вопрос
			user.City = message.Text
			userMsg.Text = askUserCar
			b.bot.Send(userMsg)

			// Обновляем состояние пользователя
			user.State = stateCar
			updateUser(b.db, user)
		case stateCar:
			message.Text = gomoji.RemoveEmojis(message.Text)
			userMsg := tgbotapi.NewMessage(message.Chat.ID, "")
			if message.Text == "" {
				// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
				userMsg.Text = userReplyPlease
				b.bot.Send(userMsg)
				return
			}

			// Отправляем пользователю следующий вопрос
			user.Car = message.Text
			userMsg.Text = askUserEngine
			b.bot.Send(userMsg)

			// Обновляем состояние пользователя
			user.State = stateEngine
			updateUser(b.db, user)
		case stateEngine:
			message.Text = gomoji.RemoveEmojis(message.Text)
			userMsg := tgbotapi.NewMessage(message.Chat.ID, "")
			if message.Text == "" {
				// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
				userMsg.Text = userReplyPlease
				b.bot.Send(userMsg)
				return
			}

			// todo в каждом сообщении нужно убирать смайлы и проверять не пустая ли строка
			// Отправляем пользователю следующий вопрос
			user.Engine = message.Text
			userMsg.Text = askUserPhoto
			b.bot.Send(userMsg)

			// Обновляем состояние пользователя
			user.State = statePhoto
			updateUser(b.db, user)
		case statePhoto:
			if message.Photo != nil && len(message.Photo) > 0 {
				// Получаем первое фото из слайса
				photoID := (message.Photo)[1].FileID

				// Добавляем фото в фото пользователя
				user.Photos = append(user.Photos, photoID)

				// Отправляем сообщение пользователю
				msg := tgbotapi.NewMessage(message.Chat.ID, "Фото було успішно додано, завантаж ще, або натисни <b>Готово</b>.")
				msg.ParseMode = parseModeHTMl
				msg.ReplyMarkup = doneButton
				b.bot.Send(msg)

				// Обновляем пользователя в базе данных
				updateUser(b.db, user)
			} else if message.Text == "Готово👌" {
				// Проверяем что бы у пользователя было загружено хоть одно фото
				if len(user.Photos) == 0 {
					// Если фото нет - отправляем уведомление
					msg := tgbotapi.NewMessage(message.Chat.ID, "Ви не завантажили жодного фото!")
					msg.ReplyMarkup = doneButton
					b.bot.Send(msg)
					return
				}

				// Отправляем сообщение администратору
				adminMsgText := fmt.Sprintf(
					"Новая заявка от пользователя. Данные:\n\n"+
						"Имя: %s\n"+
						"Город: %s\n"+
						"Автомобиль: %s\n"+
						"Двигатель: %s\n"+
						"ChatID: %d",
					user.Name,
					user.City,
					user.Car,
					user.Engine,
					message.From.ID)

				// Сообщение администратору
				adminMsg := tgbotapi.NewMessage(adminID, adminMsgText)
				adminMsg.ReplyMarkup = requestButtons
				rq, _ := b.bot.Send(adminMsg)

				// Формируем галерею с комментарием
				files := make([]interface{}, len(user.Photos))
				caption := fmt.Sprintf("ChatID: %d", message.Chat.ID)
				for i, s := range user.Photos {
					if i == 0 {
						photo := tgbotapi.InputMediaPhoto{
							BaseInputMedia: tgbotapi.BaseInputMedia{
								Type:            "photo",
								Media:           tgbotapi.FileID(s),
								Caption:         caption,
								ParseMode:       "",
								CaptionEntities: nil,
							}}
						files[i] = photo
					} else {
						files[i] = tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(s))
					}
				}
				cfg := tgbotapi.NewMediaGroup(
					adminID,
					files,
				)
				cfg.ReplyToMessageID = rq.MessageID
				if _, err := b.bot.SendMediaGroup(cfg); err != nil {
					log.Panic(err)
				}

				// Отправляем сообщение пользователю
				msg := tgbotapi.NewMessage(message.Chat.ID, userDoneRequestMsg)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				b.bot.Send(msg)

				// Сбрасываем состояние пользователя
				user.State = stateCompleted
				user.Status = statusWaiting
				updateUser(b.db, user)
			} else {
				// Просим пользователя загрузить фото
				msg := tgbotapi.NewMessage(message.Chat.ID, askUserPhoto)
				b.bot.Send(msg)
			}
		}
	}
}

// handleCallback обработка калбеков
func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	adminID, err := b.getAdminID()
	if err != nil {
		log.Panic(err)
	}

	ownerGroupID, err := b.getOwnerGroupID()
	if err != nil {
		log.Panic(err)
	}
	// разбиваем сообщение на котором висят кнопки (сама заявка админа) на массив
	s := strings.Fields(callback.Message.Text)

	// В нашем случае последний элемент массива будет chat_id (string)
	strUserID := s[len(s)-1]

	// Преобразовываем строку в число и получаем числовой `chat_id` пользователя отправившего заявку
	userChatID, _ := strconv.ParseInt(strUserID, 10, 64)

	user, err := getUser(b.db, userChatID)
	if err != nil {
		log.Panic("Ошибка получения пользователя:", err)
	}

	adminMsg := tgbotapi.NewMessage(adminID, "")
	switch user.Status {
	case statusAccepted:
		adminMsg.Text = fmt.Sprintf("Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>Прийнято</b>.", userChatID)
		adminMsg.ParseMode = parseModeHTMl
		b.bot.Send(adminMsg)
		// todo не уверен что нужно `continue`
		return
	case statusRejected:
		adminMsg.Text = fmt.Sprintf("Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>Відхилено</b>.", userChatID)
		adminMsg.ParseMode = parseModeHTMl
		b.bot.Send(adminMsg)
		// todo не уверен что нужно `continue`
		return
	case statusBanned:
		adminMsg.Text = fmt.Sprintf("Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>Заблоковано</b>.", userChatID)
		adminMsg.ParseMode = parseModeHTMl
		b.bot.Send(adminMsg)
		// todo не уверен что нужно `continue`
		return
	case statusWaiting:
		// Получаем данные из колбека
		callback := tgbotapi.NewCallback(callback.ID, callback.Data)
		userMsg := tgbotapi.NewMessage(userChatID, "")
		// todo переменная выше уже объявлена
		adminMsg := tgbotapi.NewMessage(adminID, "")

		// Действия админа по отношению к заявке
		switch callback.Text {
		case callbackAccept:
			// Создаём конфиг для ссылки на вступление в группу
			inviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: ownerGroupID,
				},
				Name:               "посилання на групу",
				ExpireDate:         0,
				MemberLimit:        1,
				CreatesJoinRequest: false,
			}

			// todo обработать возможную ошибку из запроса
			// Отправляем запрос на получение ссылки по конфигу
			resp, _ := b.bot.Request(inviteLinkConfig)
			// Собираем массив сырых байт с ответа
			data := []byte(resp.Result)
			// Создает экземпляр типа ChatInviteLink для заполнения его ответом
			var chatInviteLink tgbotapi.ChatInviteLink
			// Распарсиваем ответ в созданный выше экземпляр типа ChatInviteLink
			_ = json.Unmarshal(data, &chatInviteLink)

			// todo бот должен формировать ссылку на вступление в группу, для 1 человека и отправлять её пользователю
			userMsg.Text = "Привіт!\nТвої відповіді стосовно вступу в <b>F-club Kyiv</b> були оброблені нашою командою. Ознайомся з простими умовами спілкування в нашому клубі та приєднуйся до нас! \n\n1. Поважай інших учасників. Нецензурна лайка, цькування, використання непристойних стікерів - заборонено(але якщо це в тему, то всі розуміють😂)\n2. Не влаштовуємо «Барахолку»! Ти можешь запропонувати, якщо в тебе є щось корисне для автомобіля, чи будь що, але не треба про це писати кожного дня і робити рекламні оголошення. \n3. Якщо ти хочеш запропонувати свої послугу(сто, детейлінг, автомийки, итд) - повідом про це адміністрації і зробіть гарне оголошення разом - це все безкоштовно !! \n 4. Ми розуміємо, що зараз без цього ніяк, але маємо про це попросити - якомога менше суперечок стосовно політики. Ми всі підтримуємо Україну і не шукаємо зради!\n 5. Стосовно використання GIF , ми не проти цього, але не треба постити дуже багато, один за одним! \n 6. Май повагу до інших власників автомобілів, не у кожного така гарна машина, як в тебе!  \n\nМаєш бажання отримати клубний стікер на авто чи номерну рамку - відпиши на це повідомлення\U0001FAE1\n\nТримай посилання, для вступу в чат!\n     P.s.Не забудь привітатися з нових товаришами, та розповісти який в тебе автомобіль!\n\n\n\nДонати для розвитку!(За бажанням) \n\nF-Club Kyiv \n\n🎯Ціль: 100 000.00 ₴\n\n🔗Посилання на банку\nhttps://send.monobank.ua/jar/S87zLF6xL\n\n💳Номер картки банки\n5375 4112 0304 9692"
			userMsg.ParseMode = parseModeHTMl
			b.bot.Send(userMsg)

			userMsg.Text = fmt.Sprintf("Ось ваше <a href=\"%s\">%s</a>", chatInviteLink.InviteLink, chatInviteLink.Name)
			userMsg.ParseMode = parseModeHTMl
			b.bot.Send(userMsg)

			// todo Обновляем статусы пользователя (принять в группу)
			user.Status = statusAccepted
			updateUser(b.db, user)

			// Ответное сообщение администратору
			adminMsg.Text = fmt.Sprintf("Користувача з <b>ChatID: %d</b> підтверджено, посилання на вступ до групи надіслано!", userChatID)
			adminMsg.ParseMode = parseModeHTMl
			b.bot.Send(adminMsg)
		case callbackReject:
			// Обновляем статус пользователя
			user.Status = statusRejected
			updateUser(b.db, user)

			// Отравляем уведомление пользователю
			userMsg.Text = userRejectMsg
			userMsg.ParseMode = parseModeHTMl
			b.bot.Send(userMsg)

			// todo вынести в константу
			// Отправляем уведомление админу
			adminMsg.Text = "Користувач був успішно відхилений!"
			b.bot.Send(adminMsg)
		case callbackBanned:
			// Обновляем статус пользователя
			user.Status = statusBanned
			updateUser(b.db, user)

			// Отравляем уведомление пользователю
			userMsg.Text = userBannedMsg
			userMsg.ParseMode = parseModeHTMl
			b.bot.Send(userMsg)

			// todo вынести в константу
			// todo переменная выше уже объявлена
			// Отправляем уведомление админу
			adminMsg.Text = "Користувач був успішно заблокованний!"
			b.bot.Send(adminMsg)
		}
	}
}

// getAdminID получаем ID администратора
func (b *Bot) getAdminID() (int64, error) {
	return strconv.ParseInt(os.Getenv("OWNER_ACC"), 10, 64)
}

// getOwnerGroupID получаем ID группы в которую нужно принять пользователя
func (b *Bot) getOwnerGroupID() (int64, error) {
	return strconv.ParseInt(os.Getenv("SUPERGROUP_F30_ID"), 10, 64)
}

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
