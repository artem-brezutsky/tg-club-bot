package telegram

import (
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
	"telegram_bot/pkg/telegram/models"
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

// handleCallback Обработка калбеков
func (b *Bot) handleCallback(callbackQuery *tgbotapi.CallbackQuery) {
	// обработка калбека от администратора
	if callbackQuery.Message.Chat.ID == b.adminChatID {
		// разбиваем сообщение на котором висят кнопки (сама заявка админа) на массив
		s := strings.Fields(callbackQuery.Message.Text)
		// В нашем случае последний элемент массива будет chat_id (string)
		strUserID := s[len(s)-1]

		// todo обработать ошибку если не получилось найти chat_id
		// Преобразовываем строку в число и получаем числовой `chat_id` пользователя отправившего заявку
		userChatID, _ := strconv.ParseInt(strUserID, 10, 64)

		// Получаем пользователя Ид которого было в заявке
		user, err := b.userRepo.Get(userChatID)
		if err != nil {
			log.Panic("Ошибка получения пользователя: ", err)
		}
		// Создаём новое сообщение для админа с пустым текстом
		adminMsg := tgbotapi.NewMessage(b.adminChatID, "")

		switch user.Status {
		case models.UserStatuses.Accepted:
			adminMsg.Text = fmt.Sprintf(
				"Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>%s</b>.",
				userChatID, b.statuses[models.UserStatuses.Accepted])
			adminMsg.ParseMode = parseModeHTMl
			b.bot.Send(adminMsg)

			return
		case models.UserStatuses.Rejected:
			adminMsg.Text = fmt.Sprintf(
				"Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>%s</b>.",
				userChatID, b.statuses[models.UserStatuses.Rejected])
			adminMsg.ParseMode = parseModeHTMl
			b.bot.Send(adminMsg)

			return
		case models.UserStatuses.Banned:
			adminMsg.Text = fmt.Sprintf(
				"Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>%s</b>.",
				userChatID, b.statuses[models.UserStatuses.Banned])
			adminMsg.ParseMode = parseModeHTMl
			b.bot.Send(adminMsg)

			return
		case models.UserStatuses.Waiting:
			userMsg := tgbotapi.NewMessage(userChatID, "")
			// todo переменная выше уже объявлена
			adminMsg = tgbotapi.NewMessage(b.adminChatID, "")

			// Действия админа по отношению к заявке
			switch callbackQuery.Data {
			case callbackAccept:
				// Создаём конфиг для ссылки на вступление в группу
				inviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{
					ChatConfig: tgbotapi.ChatConfig{
						ChatID: b.closedGroupID,
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
				userMsg.Text = b.messages.UserResponses.InviteMsg
				userMsg.ParseMode = parseModeHTMl
				b.bot.Send(userMsg)

				// отправляем ссылку на группу для пользователя
				userMsg.Text = fmt.Sprintf("Ось ваше <a href=\"%s\">%s</a>", chatInviteLink.InviteLink, chatInviteLink.Name)
				userMsg.ParseMode = parseModeHTMl
				b.bot.Send(userMsg)

				// todo Обновляем статусы пользователя (принять в группу)
				user.Status = models.UserStatuses.Accepted
				b.userRepo.Update(user)

				// Ответное сообщение администратору
				adminMsg.Text = fmt.Sprintf("Користувача з <b>ChatID: %d</b> підтверджено, посилання на вступ до групи надіслано!", userChatID)
				adminMsg.ParseMode = parseModeHTMl
				b.bot.Send(adminMsg)
			case callbackReject:
				// Обновляем статус пользователя
				user.Status = models.UserStatuses.Rejected
				b.userRepo.Update(user)

				// Отравляем уведомление пользователю
				userMsg.Text = b.messages.UserResponses.RejectMsg
				userMsg.ParseMode = parseModeHTMl
				b.bot.Send(userMsg)

				// todo вынести в константу
				// Отправляем уведомление админу
				adminMsg.Text = "Користувач був успішно відхилений!"
				b.bot.Send(adminMsg)
			case callbackBanned:
				// Обновляем статус пользователя
				user.Status = models.UserStatuses.Banned
				b.userRepo.Update(user)

				// Отравляем уведомление пользователю
				userMsg.Text = b.messages.UserResponses.BannedMsg
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

			user, _ := b.userRepo.Get(chatID)

			if user.State == models.UserStates.Completed {
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
				user.ChatID)

			// Сообщение администратору
			adminMsg := tgbotapi.NewMessage(b.adminChatID, adminMsgText)
			adminMsg.ReplyMarkup = requestButtons
			rq, _ := b.bot.Send(adminMsg)

			mgc := createMediaGroup(user, chatID, b.adminChatID)
			mgc.ReplyToMessageID = rq.MessageID
			if _, err := b.bot.SendMediaGroup(mgc); err != nil {
				log.Panic(err)
			}

			// Отправляем сообщение пользователю
			msg := tgbotapi.NewMessage(chatID, b.messages.UserResponses.DoneRequestMsg)
			b.bot.Send(msg)

			// Сбрасываем состояние пользователя
			user.State = models.UserStates.Completed
			user.Status = models.UserStatuses.Waiting
			b.userRepo.Update(user)

			// Удаляем MessageID пользователя, который отправил заявку
			delete(b.lastMessageID, chatID)

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

// handleCommands Обработка команд
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

// handleMessage Обработка сообщений
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// ID текущего чата/пользователя
	chatID := message.Chat.ID

	// todo сделать обработку сообщений из группы
	// Если сообщение из чата группы, пропускаем его
	if chatID == b.closedGroupID {
		b.handleMessageFromGroup(message)
	}

	user, err := b.userRepo.Get(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = b.userRepo.Create(chatID)
		} else {
			log.Fatalln("Не корректная работа с базой данных.", err)
		}
	}

	userReplyMsg := tgbotapi.NewMessage(chatID, "")
	userReplyMsg.ParseMode = parseModeHTMl

	// Проверяем статус пользователя
	switch user.Status {
	case models.UserStatuses.Accepted:
		userReplyMsg.Text = b.messages.UserResponses.AlreadyDoneMsg
		b.bot.Send(userReplyMsg)

		return
	case models.UserStatuses.Rejected:
		userReplyMsg.Text = b.messages.UserResponses.RejectMsg
		b.bot.Send(userReplyMsg)

		return
	case models.UserStatuses.Banned:
		userReplyMsg.Text = b.messages.UserResponses.BannedMsg
		b.bot.Send(userReplyMsg)

		return
	case models.UserStatuses.Waiting:
		userReplyMsg.Text = b.messages.UserResponses.WaitingMsg
		b.bot.Send(userReplyMsg)

		return
	case models.UserStatuses.New:
		// Очищаем ввод пользователя от emoji
		message.Text = gomoji.RemoveEmojis(message.Text)

		// Если после очистки от emoji сообщение стало пустым, просим заново ввести ответ
		// Если это не фото для состояния с ожиданием фото
		// todo переделать
		if message.Text == "" && user.State != models.UserStates.Photo {
			userReplyMsg.Text = b.messages.UserResponses.ReplyPlease
			b.bot.Send(userReplyMsg)

			return
		}

		// todo нужно проверять обновился ли пользователь и если что возвращать ошибку
		switch user.State {
		case models.UserStates.Initial:
			// todo нужно переделать
			// Отправляем приветственное сообщение
			userReplyMsg.Text = b.messages.UserResponses.WelcomeMsg
			b.bot.Send(userReplyMsg)

			// Отправляем первый вопрос
			userReplyMsg.Text = b.messages.Questions.UserName
			b.bot.Send(userReplyMsg)
			// Изменяем состояние пользователя и сохраняем данные
			user.State = models.UserStates.Name
			b.userRepo.Update(user)

			return
		case models.UserStates.Name:
			// Записываем введенный ответ на предыдущий вопрос от пользователя и обновляем состояние
			user.Name = message.Text
			user.State = models.UserStates.City
			// Сохраняем данные пользователя
			b.userRepo.Update(user)

			// Отправляем следующий вопрос пользователю
			userReplyMsg.Text = b.messages.Questions.UserCity
			b.bot.Send(userReplyMsg)

			return
		case models.UserStates.City:
			// Записываем введенный ответ на предыдущий вопрос от пользователя и обновляем состояние
			user.City = message.Text
			user.State = models.UserStates.Car
			// Сохраняем данные пользователя
			b.userRepo.Update(user)

			// Отправляем пользователю следующий вопрос
			userReplyMsg.Text = b.messages.Questions.UserCar
			b.bot.Send(userReplyMsg)

			return
		case models.UserStates.Car:
			// Записываем введенный ответ на предыдущий вопрос от пользователя и обновляем состояние
			user.Car = message.Text
			user.State = models.UserStates.Engine
			// Сохраняем данные пользователя
			b.userRepo.Update(user)

			// Отправляем пользователю следующий вопрос
			userReplyMsg.Text = b.messages.Questions.UserEngine
			b.bot.Send(userReplyMsg)

			return
		case models.UserStates.Engine:
			// Записываем введенный ответ на предыдущий вопрос от пользователя и обновляем состояние
			user.Engine = message.Text
			user.State = models.UserStates.Photo
			// Сохраняем данные пользователя
			b.userRepo.Update(user)

			// Отправляем пользователю следующий вопрос
			userReplyMsg.Text = b.messages.Questions.UserPhoto
			b.bot.Send(userReplyMsg)

			return
		case models.UserStates.Photo:
			if message.Photo != nil && len(message.Photo) > 0 {
				// todo отправляет несколько кнопок готово если кол-во фото большое иногда
				b.handlePhoto(message, user)
			} else {
				// Если пришло текстовое сообщение смотрим есть ли загруженные у пользователя фото
				// Если есть, просим нажать готово, или загрузить ещё
				if len(user.Photos) > 0 {
					// todo подумать над этим
					// Удаляем сообщение с кнопкой которое было при загрузке фото
					delM := tgbotapi.NewDeleteMessage(message.Chat.ID, b.lastMessageID[message.Chat.ID])
					b.bot.Send(delM)

					// Отправляем новое сообщение с кнопкой
					txt := fmt.Sprintf("Ви успішно завантажили %d фото. Натисніть \"Готово\".", len(user.Photos))
					m := tgbotapi.NewMessage(message.Chat.ID, txt)
					m.ReplyMarkup = &stopUploadPhotoButton
					newMsg, _ := b.bot.Send(m)

					// Запоминаем ИД сообщения с кнопкой "готово"
					b.lastMessageID[message.Chat.ID] = newMsg.MessageID
					return
				}
				// Просим пользователя загрузить фото если у него ещё нет загруженных фото
				msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.Questions.UserPhoto)
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
		b.userRepo.Update(user)
	} else {
		rdDots := getRandomDots()
		txt := fmt.Sprintf("Ви успішно завантажили %d фото.\nНатисніть \"Готово\"%s", len(user.Photos), rdDots)
		m := tgbotapi.NewEditMessageText(chatID, b.lastMessageID[chatID], txt)
		m.ReplyMarkup = &stopUploadPhotoButton

		newMessage, err := b.bot.Send(m)
		if err == nil {
			b.lastMessageID[chatID] = newMessage.MessageID
			return
		}
	}

	// сообщение пользователю об успешной загрузке фото
	txt := fmt.Sprintf("Ви успішно завантажили %d фото.\nНатисніть \"Готово\".", len(user.Photos))

	if b.lastMessageID[chatID] != 0 && messageID < b.lastMessageID[chatID] {
		m := tgbotapi.NewEditMessageText(chatID, b.lastMessageID[chatID], txt)
		m.ReplyMarkup = &stopUploadPhotoButton

		newMessage, err := b.bot.Send(m)
		if err == nil {
			b.lastMessageID[chatID] = newMessage.MessageID

			return
		}
	} else if b.lastMessageID[chatID] != 0 && messageID > b.lastMessageID[chatID] {
		m := tgbotapi.NewDeleteMessage(chatID, b.lastMessageID[chatID])
		b.bot.Send(m)
	}

	m := tgbotapi.NewMessage(chatID, txt)
	m.ReplyMarkup = &stopUploadPhotoButton

	newMessage, err := b.bot.Send(m)
	if err != nil {
		return
	}

	b.lastMessageID[chatID] = newMessage.MessageID
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

				adminMsg := tgbotapi.NewMessage(b.adminChatID, "")
				adminMsg.ParseMode = parseModeHTMl

				// todo сделать обработку ошибок, что бы вероятно отправлялись админу или мне?
				// Находим пользователя
				user, err := b.userRepo.Get(chatID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						adminMsg.Text = fmt.Sprintf("Користувача з <b>ID: %d</b> не існує в базі!", chatID)
					} else {
						adminMsg.Text = "Сталася помилка!"
					}

					b.bot.Send(adminMsg)
					return
				}

				user.State = models.UserStates.Initial
				user.Status = models.UserStatuses.New
				user.Photos = nil
				b.userRepo.Update(user)

				adminMsg.Text = fmt.Sprintf("Користувача з <b>ID: %d</b> було оновлено", chatID)
				b.bot.Send(adminMsg)

				return
			} else {
				adminMsg := tgbotapi.NewMessage(b.adminChatID, "Введи ID користувача якого ти хочешь видалити з бази.")
				b.bot.Send(adminMsg)

				return
			}
		}
	} else {
		adminMsg := tgbotapi.NewMessage(b.adminChatID, "Привіт Адмін.\nЯкщо ти хочеш оновити дані користувача, то введи команду:\n/refresh + ID користувача")
		b.bot.Send(adminMsg)

		return
	}
}

// handleMessageFromGroup Обработка сообщений из группы
func (b *Bot) handleMessageFromGroup(message *tgbotapi.Message) {
	if message.NewChatMembers != nil {
		for _, newMember := range message.NewChatMembers {
			var replyName string
			switch {
			case newMember.UserName != "":
				replyName = newMember.UserName
				break
			case newMember.FirstName != "" && newMember.FirstName != "ㅤ":
				// Имя может быть с символом пустоты ;)
				replyName = newMember.FirstName
				break
			case newMember.LastName != "":
				replyName = newMember.LastName
				break
			default:
				replyName = "Водій BMW:\\)"
			}

			// todo возможно стоит переделать под ParseMode=HTML, что бы лучше контролировать содержимое
			mention := fmt.Sprintf("[%v](tg://user?id=%v)", escapeString(replyName), strconv.FormatInt(newMember.ID, 10))
			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(b.messages.GroupWelcomeMsg, mention))
			msg.ParseMode = "MarkdownV2"
			b.bot.Send(msg)

			return
		}
	} else if message.LeftChatMember != nil {
		// todo реализация отправки сообщения когда пользователь покинул группу
		return
	}

	// todo возможная реализация обработки всех сообщений в группе
	return
}
