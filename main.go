package main

import (
	"encoding/json"
	"github.com/forPelevin/gomoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"strings"
)

const StatusAccepted = 3
const StatusDeclined = 4
const StatusWaiting = 5
const StatusBanned = 7

// todo Сделать проверку того что отправляют в ответе, что бы текст был текстом, не стикер или эмодзи!!! Иначе возвращать на шаг назад
// todo Сделать кнопки для админа, которые будут принимать либо отклонять заявки

// Request Сущность пользователя
type Request struct {
	Id     int
	ChatId int64
	Status int
	Step   int
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	Token := os.Getenv("TOKEN")
	OwnerAcc, _ := strconv.ParseInt(os.Getenv("OWNER_ACC"), 10, 64)
	//SupergroupId, _ := strconv.ParseInt(os.Getenv("SUPERGROUP_ID"), 10, 64)
	SupergroupF30Id, _ := strconv.ParseInt(os.Getenv("SUPERGROUP_F30_ID"), 10, 64)

	bot, err := tgbotapi.NewBotAPI(Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// todo Вынести в конфиг
	dsn := "admin:root@tcp(127.0.0.1:3306)/bmw_club_bot"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Ответы на вопросы
	var (
		answer1 = ""
		answer2 = ""
		answer3 = ""
	)

	var (
		question1 = "Как тебя зовут?"
		question2 = "Какое у тебя авто?"
		question3 = "Какой двигатель?"
	)

	var doneButton = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Готово👌"),
		),
	)

	var requestButtons = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Принять", "accept_request"),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", "reject_request"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Послать нахуй пса", "fuck_off_dog"),
		),
	)

	// Массив ИД файлов для отправки
	var answerFileIds []string = nil
	var isDocumentFiles = false
	var isPhotoFiles = false

	for update := range updates {

		fromChat := update.FromChat()
		if fromChat.ID == SupergroupF30Id {
			continue
		}

		if update.Message != nil { // If we got a message

			// Ид текущего чата
			chatID := update.Message.Chat.ID
			msg := tgbotapi.NewMessage(chatID, "")

			var userRequest Request

			// @todo тестируем
			//if update.Message.From.ID == 123 {
			//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			//
			//switch update.Message.Text {
			//case "open":
			//	msg.ReplyMarkup = but
			//case "close":
			//	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			//}
			//
			//if _, err := bot.Send(msg); err != nil {
			//	log.Panic(err)
			//}
			//
			//continue
			// Если это фото
			//if update.Message.Photo != nil {
			//	//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			//	answerFileIds = append(answerFileIds, update.Message.Photo[1].FileID)
			//	msg.ReplyMarkup = doneButton
			//	if _, err := bot.Send(msg); err != nil {
			//		log.Panic(err)
			//	}
			//} else if update.Message.Document != nil &&
			//	strings.Contains(update.Message.Document.MimeType, "image") {
			//	//msg := tgbotapi.NewDocument(OwnerAcc, tgbotapi.FileID(update.Message.Document.FileID))
			//	answerFileIds = append(answerFileIds, update.Message.Document.FileID)
			//	msg.ReplyMarkup = doneButton
			//	if _, err := bot.Send(msg); err != nil {
			//		log.Panic(err)
			//	}
			//}

			//if update.Message.Text == "Готово👌" {
			//	msg := tgbotapi.NewMessage(OwnerAcc, "Заявка принята")
			//	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			//	if _, err := bot.Send(msg); err != nil {
			//		log.Panic(err)
			//	}
			//	log.Println("Готово👌")
			//	// Готово👌
			//	continue
			//}

			//msg := tgbotapi.NewMessage(OwnerAcc, update.Message.Text)
			//bot.Send(msg)
			//continue
			//handleOwnerMessage(update)
			//if update.Message.ReplyToMessage != nil {
			//	var replyUserRequest Request
			//	replyUserRequest, err = getUserRequestForMessageId(*db, update.Message.ReplyToMessage.MessageID)
			//	if err != nil {
			//		log.Fatal(err.Error())
			//	}
			//	//replyUser := db.Where("message_id = ?", update.Message.ReplyToMessage.MessageID).First(&userRequest)
			//	msg := tgbotapi.NewMessage(replyUserRequest.ChatId, update.Message.Text)
			//	bot.Send(msg)
			//	continue
			//}
			//ownerGreeting := "Hello My Kid!"
			//msg := tgbotapi.NewMessage(OwnerAcc, ownerGreeting)
			//msg.ReplyToMessageID = update.Message.MessageID
			//
			//bot.Send(msg)
			//continue
			//}

			if update.Message.From.ID == OwnerAcc {

				// Создаём конфиг для ссылки на вступление в группу
				inviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{
					ChatConfig: tgbotapi.ChatConfig{
						ChatID: SupergroupF30Id,
					},
					Name:               "",
					ExpireDate:         0,
					MemberLimit:        1,
					CreatesJoinRequest: false,
				}

				// Отправляем запрос на получение ссылки по конфигу
				resp, _ := bot.Request(inviteLinkConfig)
				// Собираем массив сырых байт с ответа
				data := []byte(resp.Result)
				// Создает экземпляр типа ChatInviteLink для заполнения его ответом
				var inviteLink2 tgbotapi.ChatInviteLink
				// Распарсиваем ответ в созданный выше экземпляр типа ChatInviteLink
				_ = json.Unmarshal(data, &inviteLink2)

				log.Println(inviteLink2.InviteLink)
			}

			// Проверка пользователя на существование
			result := db.Where("chat_id = ?", chatID).First(&userRequest)
			if result.RowsAffected > 0 { // Есть ли пользователь в БД?
				// Если есть пользователь, проверяем его статус
				switch userRequest.Status {
				case StatusAccepted:
					// Пользователь уже зарегистрирован и добавлен в группу
					msg.Text = "Вы уже приняты!"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					log.Printf("Пользователь уже принят: [%d]", userRequest.ChatId)
					continue
				case StatusDeclined:
					// Пользователь отклонён
					msg.Text = "Ваша заявка уже была отклонена!"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					log.Printf("Пользователь уже отклонён: [%d]", userRequest.ChatId)
					continue
				case StatusWaiting:
					msg.Text = "Ваша заявка на рассмотрении!"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}

				log.Printf("Пользователь найден: [%d]", userRequest.ChatId)
			} else {
				// Если запись не найдена, создаем нового пользователя
				userRequest = Request{ChatId: chatID}
				db.Create(&userRequest)
				log.Printf("Пользователь создан: [%d]", userRequest.ChatId)
				// todo Возможно проверить на ошибку создания пользователя?
			}

			switch userRequest.Step {
			case 0:
				log.Println("Новый пользователь, начинаем диалог...")
				msg.Text = "Привет, сейчас я задам тебе несколько вопросов."
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}

				msg.Text = question1
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}

				userRequest.Step = userRequest.Step + 1
				db.Save(&userRequest)
				continue
			case 1:
				// todo Вынести в отдельный метод.
				// Проверяем ответ пользователя на emoji
				update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
				if update.Message.Text == "" {
					msg.Text = "Пожалуйста, ответьте на вопрос выше!"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					msg.Text = question1
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}
				answer1 = update.Message.Text
				msg.Text = question2
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				userRequest.Step = userRequest.Step + 1
				db.Save(&userRequest)
				continue
			case 2:
				// todo Вынести в отдельный метод.
				// Проверяем ответ пользователя на emoji
				update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
				if update.Message.Text == "" {
					msg.Text = "Пожалуйста, ответьте на вопрос выше!"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					msg.Text = question2
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}
				answer2 = update.Message.Text
				msg.Text = question3
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				userRequest.Step = userRequest.Step + 1
				db.Save(&userRequest)
				continue
			case 3:
				// todo Вынести в отдельный метод.
				// Проверяем ответ пользователя на emoji
				update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
				if update.Message.Text == "" {
					msg.Text = "Пожалуйста, ответьте на вопрос выше!"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					msg.Text = question3
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}
				answer3 = update.Message.Text
				msg.Text = "Пришли фото автомобиля. После этого нажми кнопку \"Готово\""
				msg.ReplyMarkup = doneButton
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				userRequest.Step = userRequest.Step + 1
				db.Save(&userRequest)
				continue
			case 4:
				if update.Message.Photo != nil {
					answerFileIds = append(answerFileIds, update.Message.Photo[1].FileID)
					isPhotoFiles = true
					isDocumentFiles = false
					continue
				} else if update.Message.Document != nil &&
					strings.Contains(update.Message.Document.MimeType, "image") {
					answerFileIds = append(answerFileIds, update.Message.Document.FileID)
					isDocumentFiles = true
					isPhotoFiles = false
					continue
				}

				if update.Message.Text == "Готово👌" {
					if answerFileIds == nil {
						msg.Text = "Пришли фото автомобиля. После этого нажми кнопку \"Готово\""
						msg.ReplyMarkup = doneButton
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						continue
					}
					msg.Text = "Спасибо, твоя заявка отправлена администратору!"
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}

					// Обновляем статусы
					userRequest.Step = 5
					userRequest.Status = StatusWaiting
					db.Save(&userRequest)
					// Готово👌

					// Отправка заявки админу
					totalAnswer := "Имя: " + answer1 + " \n"
					totalAnswer += "Автомобиль: " + answer2 + " \n"
					totalAnswer += "Двигатель: " + answer3 + " \n"
					totalAnswer += "ChatID: " + strconv.FormatInt(chatID, 10) + " \n"

					// Отправляем текст заявки
					// todo сделать кнопки принять/отклонить
					msg := tgbotapi.NewMessage(OwnerAcc, totalAnswer)
					msg.ReplyMarkup = requestButtons
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}

					files := make([]interface{}, len(answerFileIds))
					for i, s := range answerFileIds {
						if isPhotoFiles {
							files[i] = tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(s))
							continue
						}

						if isDocumentFiles {
							files[i] = tgbotapi.NewInputMediaDocument(tgbotapi.FileID(s))
							continue
						}
					}
					cfg := tgbotapi.NewMediaGroup(
						OwnerAcc,
						files,
					)

					if _, err := bot.SendMediaGroup(cfg); err != nil {
						log.Panic(err)
					}

					// todo придумать как чистить массив с файлами? Если это нужно будет? Массив не очищается после заполнения файлами
					answerFileIds = nil
					continue
				} else {
					msg.Text = "Пришли фото автомобиля. После этого нажми кнопку \"Готово\""
					msg.ReplyMarkup = doneButton
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}
			}

			//if userRequest.Step == 0 {
			//
			//} else if userRequest.Step == 1 {
			//
			//} else if userRequest.Step == 2 {
			//
			//} else if userRequest.Step == 3 {
			//
			//} else if userRequest.Step == 4 {
			//
			//}

			//continue
			// Проверка команда ли это?
			//if update.Message.IsCommand() {
			//
			//	msg := tgbotapi.NewMessage(chatID, "")
			//
			//	switch update.Message.Command() {
			//	case "start":
			//
			//	default:
			//		msg.Text = "Неизвестная команда"
			//		if _, err := bot.Send(msg); err != nil {
			//			log.Panic(err)
			//		}
			//		continue
			//	}
			//
			//	// На этом этапе мы уже обработали пользователя, получили его данные или создали новую запись
			//	// todo возможно нужно вывести в лог уведомление что пользователь обработан
			//	log.Println("Обработана команда /start!")
			//	log.Println("Пользователь обработан!")
			//} else {
			//	// Сообщение отправленное пользователем, обрабатываем и определяем на каком шаге пользователь
			//	msg := tgbotapi.NewMessage(chatID, "")
			//	result := db.Where("chat_id = ?", chatID).First(&userRequest)
			//	if result.RowsAffected > 0 { // Есть ли пользователь в БД?
			//		// Если есть пользователь, проверяем его статус
			//		switch userRequest.Status {
			//		case StatusAccepted:
			//			// Пользователь уже зарегистрирован и добавлен в группу
			//			msg.Text = "Вы уже приняты!"
			//			if _, err := bot.Send(msg); err != nil {
			//				log.Panic(err)
			//			}
			//			continue
			//		case StatusDeclined:
			//			// Пользователь отклонён
			//			msg.Text = "Ваша заявка была отклонена!"
			//			if _, err := bot.Send(msg); err != nil {
			//				log.Panic(err)
			//			}
			//			continue
			//		}
			//
			//	} else {
			//		// Если запись не найдена, создаем нового пользователя
			//		userRequest = Request{ChatId: chatID}
			//		db.Create(&userRequest)
			//		// todo Возможно проверить на ошибку создания пользователя?
			//	}
			//}

			// найти пользователя, либо создать его
			//db.Clauses(clause.OnConflict{
			//	Columns:   []clause.Column{{Name: "chat_id"}},
			//	DoUpdates: clause.AssignmentColumns([]string{"message_id"}),
			//}).Create(&userRequest)

			// Начинаем проверять входящее сообщение
			// Если это команда
			//if update.Message.IsCommand() {
			//	// Обработка команды /start
			//
			//	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			//
			//	switch update.Message.Command() {
			//	case "start":
			//
			//		// задать вопрос
			//		msg.Text = "Привіт! З якого ти міста? за бажанням - вкажи власне ім'я?"
			//	default:
			//		msg.Text = "I don't know that command"
			//	}
			//
			//	if _, err := bot.Send(msg); err != nil {
			//		log.Panic(err)
			//	}
			//	//handleCommand(update)
			//
			//	continue
			//}

			//if update.Message.From.ID == OwnerAcc {
			//	//handleOwnerMessage(update)
			//	if update.Message.ReplyToMessage != nil {
			//		var replyUserRequest Request
			//		replyUserRequest, err = getUserRequestForMessageId(*db, update.Message.ReplyToMessage.MessageID)
			//		if err != nil {
			//			log.Fatal(err.Error())
			//		}
			//		//replyUser := db.Where("message_id = ?", update.Message.ReplyToMessage.MessageID).First(&userRequest)
			//		msg := tgbotapi.NewMessage(replyUserRequest.ChatId, update.Message.Text)
			//		bot.Send(msg)
			//		continue
			//	}
			//	ownerGreeting := "Hello My Kid!"
			//	msg := tgbotapi.NewMessage(OwnerAcc, ownerGreeting)
			//	//msg.ReplyToMessageID = update.Message.MessageID
			//
			//	bot.Send(msg)
			//	continue
			//}

			//msg = tgbotapi.NewForward(OwnerAcc, update.Message.From.ID, update.Message.MessageID)
			//msg.ReplyToMessageID = update.Message.MessageID
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)

			// разбиваем сообщение на котором висят кнопки (сама заявка админа) на массив
			s := strings.Fields(update.CallbackQuery.Message.Text)
			// В нашем случае последний элемент массива будет chat_id (string)
			strChatID := s[len(s)-1]
			// Преобразовываем string to int64
			requestUserChatID, err := strconv.ParseInt(strChatID, 10, 64)
			if err != nil {
				panic(err)
			}

			// Получаем пользователя, заявку которого рассматриваем
			var user Request
			result := db.Where("chat_id = ?", requestUserChatID).First(&user)
			if result.Error != nil {
				log.Panic(result.Error.Error())
			}

			// Проверка был ли рассмотрен уже текущий пользователь при нажатии на ответные кнопки
			// todo переделать все статусы что бы брать их тайтлы
			// todo вынести всё в отдельные функции
			if user.Status == StatusAccepted {
				replText := "Пользователь был рассмотрен! \n"
				replText += "Текущий статус пользователя с ChatID: " +
					strconv.FormatInt(OwnerAcc, 10) + " - Принят"

				msg := tgbotapi.NewMessage(OwnerAcc, replText)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				continue
			} else if user.Status == StatusDeclined {
				replText := "Пользователь был рассмотрен! \n"
				replText += "Текущий статус пользователя с ChatID: " +
					strconv.FormatInt(OwnerAcc, 10) + " - Отклонён"

				msg := tgbotapi.NewMessage(OwnerAcc, replText)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				continue
			} else if user.Status == StatusBanned {
				replText := "Пользователь был рассмотрен! \n"
				replText += "Текущий статус пользователя с ChatID: " +
					strconv.FormatInt(OwnerAcc, 10) + " - В бане"

				msg := tgbotapi.NewMessage(OwnerAcc, replText)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				continue
			}

			// Если команда fuck_off_dog, обрабатываем и шлём сообщение клиенту оставившему заявку
			if callback.Text == "fuck_off_dog" {
				// Шлём пса на хуй
				msg := tgbotapi.NewMessage(requestUserChatID, "Иди на хуй, пёс!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				// todo обновлять статус, вероятно сделать новый, что-то типа "бана"
				user.Status = StatusBanned
				db.Save(&user)

				// Ответное сообщение администратору
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пользователь успешно послан на хуй!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			} else if callback.Text == "accept_request" {

				// Создаём конфиг для ссылки на вступление в группу
				inviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{
					ChatConfig: tgbotapi.ChatConfig{
						ChatID: SupergroupF30Id,
					},
					Name:               "ссылка на группу!",
					ExpireDate:         0,
					MemberLimit:        1,
					CreatesJoinRequest: false,
				}

				// todo обработать возможную ошибку из запроса
				// Отправляем запрос на получение ссылки по конфигу
				resp, _ := bot.Request(inviteLinkConfig)
				// Собираем массив сырых байт с ответа
				data := []byte(resp.Result)
				// Создает экземпляр типа ChatInviteLink для заполнения его ответом
				var chatInviteLink tgbotapi.ChatInviteLink
				// Распарсиваем ответ в созданный выше экземпляр типа ChatInviteLink
				_ = json.Unmarshal(data, &chatInviteLink)

				// todo бот должен формировать ссылку на вступление в группу, для 1 человека и отправлять её пользователю
				respText := "Поздравляем, ваша заявка принята! \n"
				respText += "Вот ваша <a href=\"" + chatInviteLink.InviteLink + "\">" + chatInviteLink.Name + "</a>\n"
				msg := tgbotapi.NewMessage(requestUserChatID, respText)
				msg.ParseMode = "HTML"
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				// todo Обновляем статусы пользователя (принять в группу)
				user.Status = StatusAccepted
				db.Save(&user)

				// Ответное сообщение администратору
				// Вероятно нужно сюда выводить chat_id, что бы понять кого приняли в группу
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пользователь подтвержден, ссылка на вступление в группу отправлена!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			} else if callback.Text == "reject_request" {

				// todo что то придумать тут
				respText := "Ваша заявка была отклонена, для информации свяжитесь с <a href=\"tg://user?id=6225178130\">администратором</a>."
				msg := tgbotapi.NewMessage(requestUserChatID, respText)
				msg.ParseMode = "HTML"
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				// todo Обновляем статусы пользователя (заявка отклонена)
				user.Status = StatusDeclined
				db.Save(&user)

				// Ответное сообщение администратору
				// Вероятно нужно сюда выводить chat_id, что бы понять кого приняли в группу
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Заявка пользователя отклонена, информация отправлена!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			}

			// Отправка колбека обратно
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}

			// And finally, send a message containing the data received.

		}
	}
}
