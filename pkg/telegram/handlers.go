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
	"regexp"
	"strconv"
	"strings"
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

// Кнопка отправки фото для пользователя
var stopUploadPhotoButton = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Готово👌", "upload_done"),
	),
)

// handleCallback обработка калбеков
func (b *Bot) handleCallback(callbackQuery *tgbotapi.CallbackQuery) {
	// обработка калбека от администратора
	if callbackQuery.Message.Chat.ID == b.AdminChatID {
		// разбиваем сообщение на котором висят кнопки (сама заявка админа) на массив
		s := strings.Fields(callbackQuery.Message.Text)
		// В нашем случае последний элемент массива будет chat_id (string)
		strUserID := s[len(s)-1]

		// todo обработать ошибку если не получилось найти chat_id
		// Преобразовываем строку в число и получаем числовой `chat_id` пользователя отправившего заявку
		userChatID, _ := strconv.ParseInt(strUserID, 10, 64)

		// Получаем пользователя Ид которого было в заявке
		user, err := getUser(b.db, userChatID)
		if err != nil {
			log.Panic("Ошибка получения пользователя: ", err)
		}
		// Создаём новое сообщение для админа с пустым текстом
		adminMsg := tgbotapi.NewMessage(b.AdminChatID, "")

		switch user.Status {
		case statusAccepted:
			adminMsg.Text = fmt.Sprintf(
				"Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>%s</b>.",
				userChatID, b.Statuses[statusAccepted])
			adminMsg.ParseMode = parseModeHTMl
			b.bot.Send(adminMsg)

			return
		case statusRejected:
			adminMsg.Text = fmt.Sprintf(
				"Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>%s</b>.",
				userChatID, b.Statuses[statusRejected])
			adminMsg.ParseMode = parseModeHTMl
			b.bot.Send(adminMsg)

			return
		case statusBanned:
			adminMsg.Text = fmt.Sprintf(
				"Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>%s</b>.",
				userChatID, b.Statuses[statusBanned])
			adminMsg.ParseMode = parseModeHTMl
			b.bot.Send(adminMsg)

			return
		case statusWaiting:
			userMsg := tgbotapi.NewMessage(userChatID, "")
			// todo переменная выше уже объявлена
			adminMsg = tgbotapi.NewMessage(b.AdminChatID, "")

			// Действия админа по отношению к заявке
			switch callbackQuery.Data {
			case callbackAccept:
				// Создаём конфиг для ссылки на вступление в группу
				inviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{
					ChatConfig: tgbotapi.ChatConfig{
						ChatID: b.OwnerGroupID,
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

				// отправляем приветственное сообщение пользователю
				userMsg.Text = userInviteMsg
				userMsg.ParseMode = parseModeHTMl
				b.bot.Send(userMsg)

				// отправляем ссылку на группу для пользователя
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
	} else {
		switch callbackQuery.Data {
		case "upload_done":
			// завершаем работу и отправляем админу заявку
			// todo ограничить кол-во фото которые можно загрузить
			// todo придумать как убрать кнопку готово после нажатия и успешной отправки заявки

			chatID := callbackQuery.Message.Chat.ID
			var user models.User
			// todo надо переделать
			if err := b.db.Where("telegram_id = ?", chatID).First(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Panic("Пользователь не найден")
				}
			}

			if user.State == stateCompleted {
				answerCallback := tgbotapi.NewCallback(callbackQuery.ID, "Заявку вже було відправлено!")
				if _, err := b.bot.Request(answerCallback); err != nil {
					panic(err)
				}

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
				user.TelegramID)

			// Сообщение администратору
			adminMsg := tgbotapi.NewMessage(b.AdminChatID, adminMsgText)
			adminMsg.ReplyMarkup = requestButtons
			rq, _ := b.bot.Send(adminMsg)

			mgc := createMediaGroup(&user, chatID, b.AdminChatID)
			//// Формируем галерею с комментарием
			//files := make([]interface{}, len(user.Photos))
			//caption := fmt.Sprintf("ChatID: <b>%d</b>", chatID)
			//for i, s := range user.Photos {
			//	if i == 0 {
			//		photo := tgbotapi.InputMediaPhoto{
			//			BaseInputMedia: tgbotapi.BaseInputMedia{
			//				Type:            "photo",
			//				Media:           tgbotapi.FileID(s),
			//				Caption:         caption,
			//				ParseMode:       parseModeHTMl,
			//				CaptionEntities: nil,
			//			}}
			//		files[i] = photo
			//	} else {
			//		files[i] = tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(s))
			//	}
			//}
			//cfg := tgbotapi.NewMediaGroup(
			//	b.AdminChatID,
			//	files,
			//)
			mgc.ReplyToMessageID = rq.MessageID
			if _, err := b.bot.SendMediaGroup(mgc); err != nil {
				log.Panic(err)
			}

			// Отправляем сообщение пользователю
			msg := tgbotapi.NewMessage(chatID, userDoneRequestMsg)
			b.bot.Send(msg)

			// Сбрасываем состояние пользователя
			user.State = stateCompleted
			user.Status = statusWaiting
			updateUser(b.db, &user)

			// Удаляем MessageID пользователя, который отправил заявку
			delete(lastBotMessageIDInChat, chatID)

			return
		default:
			// Неизвестная команда
			return
		}
	}

	// Уведомление о нажатии на кнопку калбека
	answerCallback := tgbotapi.NewCallback(callbackQuery.ID, "Зроблено :)")
	if _, err := b.bot.Request(answerCallback); err != nil {
		panic(err)
	}
}

// handleCommands обработка команд
func (b *Bot) handleCommands(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Hello, I'm your bot!")
		b.bot.Send(msg)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "I don't know that command")
		b.bot.Send(msg)
	}
}

// handleMessage обработка сообщений
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// ID текущего чата/пользователя
	chatID := message.Chat.ID

	// todo сделать обработку сообщений из группы
	// Если сообщение из чата группы, пропускаем его
	if chatID == b.OwnerGroupID {
		b.handleMessageFromGroup(message)
	}

	user, err := getUser(b.db, chatID)
	if err != nil {
		log.Panic("Ошибка получения пользователя: ", err)
	}

	userReplyMsg := tgbotapi.NewMessage(chatID, userAlreadyDoneMsg)
	userReplyMsg.ParseMode = parseModeHTMl

	// Проверяем статус пользователя
	switch user.Status {
	case statusAccepted:
		userReplyMsg.Text = userAlreadyDoneMsg
		b.bot.Send(userReplyMsg)

		return
	case statusRejected:
		userReplyMsg.Text = userRejectMsg
		b.bot.Send(userReplyMsg)

		return
	case statusBanned:
		userReplyMsg.Text = userBannedMsg
		b.bot.Send(userReplyMsg)

		return
	case statusWaiting:
		userReplyMsg.Text = userWaitingMsg
		b.bot.Send(userReplyMsg)

		return
	case statusNew:
		// Очищаем ввод пользователя от emoji
		message.Text = gomoji.RemoveEmojis(message.Text)

		// Если после очистки от emoji сообщение стало пустым, просим заново ввести ответ
		if message.Text == "" && user.State != statePhoto {
			userReplyMsg.Text = userReplyPlease

			return
		}

		// todo нужно проверять обновился ли пользователь и если что возвращать ошибку
		switch user.State {
		case stateInitial:
			// todo нужно переделать
			// Отправляем приветственное сообщение
			userReplyMsg.Text = userWelcomeMsg
			b.bot.Send(userReplyMsg)

			// Отправляем первый вопрос
			userReplyMsg.Text = askUserName
			b.bot.Send(userReplyMsg)
			// Изменяем состояние пользователя и сохраняем данные
			user.State = stateName
			updateUser(b.db, user)

			return
		case stateName:
			// Записываем введенный ответ на предыдущий вопрос от пользователя и обновляем состояние
			user.Name = message.Text
			user.State = stateCity
			// Сохраняем данные пользователя
			updateUser(b.db, user)

			// Отправляем следующий вопрос пользователю
			userReplyMsg.Text = askUserCity
			b.bot.Send(userReplyMsg)

			return
		case stateCity:
			// Записываем введенный ответ на предыдущий вопрос от пользователя и обновляем состояние
			user.City = message.Text
			user.State = stateCar
			// Сохраняем данные пользователя
			updateUser(b.db, user)

			// Отправляем пользователю следующий вопрос
			userReplyMsg.Text = askUserCar
			b.bot.Send(userReplyMsg)

			return
		case stateCar:
			// Записываем введенный ответ на предыдущий вопрос от пользователя и обновляем состояние
			user.Car = message.Text
			user.State = stateEngine
			// Сохраняем данные пользователя
			updateUser(b.db, user)

			// Отправляем пользователю следующий вопрос
			userReplyMsg.Text = askUserEngine
			b.bot.Send(userReplyMsg)

			return
		case stateEngine:
			// Записываем введенный ответ на предыдущий вопрос от пользователя и обновляем состояние
			user.Engine = message.Text
			user.State = statePhoto
			// Сохраняем данные пользователя
			updateUser(b.db, user)

			// Отправляем пользователю следующий вопрос
			userReplyMsg.Text = askUserPhoto
			b.bot.Send(userReplyMsg)

			return
		case statePhoto:
			if message.Photo != nil && len(message.Photo) > 0 {
				b.handlePhoto(message, user)
			} else {
				// Если пришло текстовое сообщение смотрим есть ли загруженные у пользователя фото
				// Если есть, просим нажать готово, или загрузить ещё
				if len(user.Photos) > 0 {
					// todo подумать над этим
					// Удаляем сообщение с кнопкой которое было при загрузке фото
					delM := tgbotapi.NewDeleteMessage(message.Chat.ID, lastBotMessageIDInChat[message.Chat.ID])
					b.bot.Send(delM)

					// Отправляем новое сообщение с кнопкой
					txt := fmt.Sprintf("Ви успішно завантажили %d фото. Натисніть \"Готово\".", len(user.Photos))
					m := tgbotapi.NewMessage(message.Chat.ID, txt)
					m.ReplyMarkup = &stopUploadPhotoButton
					newMsg, _ := b.bot.Send(m)

					// Запоминаем ИД сообщения с кнопкой "готово"
					lastBotMessageIDInChat[message.Chat.ID] = newMsg.MessageID
					return
				}
				// Просим пользователя загрузить фото если у него ещё нет загруженных фото
				msg := tgbotapi.NewMessage(message.Chat.ID, askUserPhoto)
				b.bot.Send(msg)

				return
			}
		}
	}
}

// handlePhoto Обработка фотографий
func (b *Bot) handlePhoto(message *tgbotapi.Message, user *models.User) {
	// ID чата/пользователя
	chatID := message.Chat.ID
	// ID текущего сообщения
	messageID := message.MessageID
	// получаем fileID фото с лучшим качеством
	photoID := (message.Photo)[len(message.Photo)-1].FileID

	if len(user.Photos) < maxUploadPhoto {
		// Добавляем fileID в фото пользователя
		user.Photos = append(user.Photos, photoID)
		// сохраняем фото
		updateUser(b.db, user)
	} else {
		rdDots := getRandomDots()
		txt := fmt.Sprintf("Ви успішно завантажили %d фото.\nНатисніть \"Готово\"%s", len(user.Photos), rdDots)
		m := tgbotapi.NewEditMessageText(chatID, lastBotMessageIDInChat[chatID], txt)
		m.ReplyMarkup = &stopUploadPhotoButton

		newMessage, err := b.bot.Send(m)
		if err == nil {
			lastBotMessageIDInChat[chatID] = newMessage.MessageID
			return
		}
	}

	// сообщение пользователю об успешной загрузке фото
	txt := fmt.Sprintf("Ви успішно завантажили %d фото. Натисніть \"Готово\".", len(user.Photos))

	if lastBotMessageIDInChat[chatID] != 0 && messageID < lastBotMessageIDInChat[chatID] {
		m := tgbotapi.NewEditMessageText(chatID, lastBotMessageIDInChat[chatID], txt)
		m.ReplyMarkup = &stopUploadPhotoButton

		newMessage, err := b.bot.Send(m)
		if err == nil {
			lastBotMessageIDInChat[chatID] = newMessage.MessageID

			return
		}
	} else if lastBotMessageIDInChat[chatID] != 0 && messageID > lastBotMessageIDInChat[chatID] {
		m := tgbotapi.NewDeleteMessage(chatID, lastBotMessageIDInChat[chatID])
		b.bot.Send(m)
	}

	m := tgbotapi.NewMessage(chatID, txt)
	m.ReplyMarkup = &stopUploadPhotoButton

	newMessage, err := b.bot.Send(m)
	if err != nil {
		return
	}

	lastBotMessageIDInChat[chatID] = newMessage.MessageID
	return
}

// todo переделать
// handleAdminMessage Обработка сообщений от администратора
func (b *Bot) handleAdminMessage(message *tgbotapi.Message) {
	if message.IsCommand() {
		switch message.Command() {
		case "refresh":
			if message.CommandArguments() != "" {
				match := regexp.MustCompile(`^\d+$`).FindStringSubmatch(message.CommandArguments())
				if len(match) == 0 {
					// Если параметры не содержат только числа, отправляем пользователю сообщение об ошибке
					msg := tgbotapi.NewMessage(message.Chat.ID, "Параметр команди має бути цілим числом.")
					b.bot.Send(msg)

					return
				}
				// Получаем ChatID из переданного параметра
				chatID, _ := strconv.ParseInt(match[0], 10, 64)
				// Находим пользователя
				var user models.User
				if err := b.db.Where("telegram_id = ?", chatID).First(&user).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						adminMsg := tgbotapi.NewMessage(b.AdminChatID, fmt.Sprintf("Користувача з <b>ID: %d</b> не існує в базі!", chatID))
						adminMsg.ParseMode = parseModeHTMl
						b.bot.Send(adminMsg)

						return
					}
				}
				user.State = stateInitial
				user.Status = statusNew
				user.Photos = nil
				updateUser(b.db, &user)

				adminMsg := tgbotapi.NewMessage(b.AdminChatID, fmt.Sprintf("Користувача з <b>ID: %d</b> було оновлено", chatID))
				adminMsg.ParseMode = parseModeHTMl
				b.bot.Send(adminMsg)

				return
			} else {
				adminMsg := tgbotapi.NewMessage(b.AdminChatID, "Введи ID користувача якого ти хочешь видалити з бази.")
				b.bot.Send(adminMsg)

				return
			}
		}
	} else {
		adminMsg := tgbotapi.NewMessage(b.AdminChatID, "Привіт Адмін.\nЯкщо ти хочеш оновити дані користувача, то введи команду:\n/refresh + ID користувача")
		b.bot.Send(adminMsg)

		return
	}
}

// handleMessageFromGroup Обработка сообщений из группы
func (b *Bot) handleMessageFromGroup(message *tgbotapi.Message) {
	return
}
