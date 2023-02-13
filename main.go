package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
)

const StatusNew = 2
const StatusAccepted = 3
const StatusDeclined = 4

// Request Сущность пользователя
type Request struct {
	ChatId    int64
	UserName  string
	FirstName string
	LastName  string
	MessageID int
	Status    int
	Step      int
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	Token := os.Getenv("TOKEN")
	OwnerAcc, _ := strconv.ParseInt(os.Getenv("OWNER_ACC"), 10, 64)

	bot, err := tgbotapi.NewBotAPI(Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(5)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// todo Вынести в конфиг
	dsn := "admin:root@tcp(127.0.0.1:3306)/bmw_club_bot"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	for update := range updates {
		if update.Message != nil { // If we got a message
			//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			// Ид текущего чата
			chatID := update.Message.Chat.ID
			messageID := update.Message.MessageID
			msg := tgbotapi.NewMessage(chatID, "")

			var userRequest Request

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
				}

				log.Printf("Пользователь найден: [%d]", userRequest.ChatId)
			} else {
				// Если запись не найдена, создаем нового пользователя
				userRequest = Request{ChatId: chatID, MessageID: messageID}
				db.Create(&userRequest)
				log.Printf("Пользователь создан: [%d]", userRequest.ChatId)
				// todo Возможно проверить на ошибку создания пользователя?
			}

			if userRequest.Step == 0 {
				log.Println("Новый пользователь, начинаем диалог...")
				msg = tgbotapi.NewMessage(chatID, "Ваше имя?")
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}

			// Проверка команда ли это?
			if update.Message.IsCommand() {

				msg := tgbotapi.NewMessage(chatID, "")

				switch update.Message.Command() {
				case "start":

				default:
					msg.Text = "Неизвестная команда"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}

				// На этом этапе мы уже обработали пользователя, получили его данные или создали новую запись
				// todo возможно нужно вывести в лог уведомление что пользователь обработан
				log.Println("Обработана команда /start!")
				log.Println("Пользователь обработан!")
			} else {
				// Сообщение отправленное пользователем, обрабатываем и определяем на каком шаге пользователь
				msg := tgbotapi.NewMessage(chatID, "")
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
						continue
					case StatusDeclined:
						// Пользователь отклонён
						msg.Text = "Ваша заявка была отклонена!"
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						continue
					}

				} else {
					// Если запись не найдена, создаем нового пользователя
					userRequest = Request{ChatId: chatID, MessageID: messageID}
					db.Create(&userRequest)
					// todo Возможно проверить на ошибку создания пользователя?
				}
			}

			// Определяем на каком шаге пользователь, что бы начать диалог
			switch userRequest.Step {
			case 0:
				// Начальный шаг, пользователь новый и еще не получал сообщения
			case 1:
				// Пользователь получил первое сообщение
				// todo возможно нужно проверять отправлено ли сообщение пользователю

			}
			if userRequest.Step == 0 {
				// Первый шаг
			}

			continue

			//var userRequest Request
			//userRequest.ChatId = update.Message.From.ID
			//userRequest.MessageID = update.Message.MessageID
			//userRequest.LastName = update.Message.From.LastName
			//userRequest.FirstName = update.Message.From.FirstName
			//userRequest.UserName = update.Message.From.UserName

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

			// Если это просто сообщение

			if update.Message.From.ID == OwnerAcc {
				//handleOwnerMessage(update)
				if update.Message.ReplyToMessage != nil {
					var replyUserRequest Request
					replyUserRequest, err = getUserRequestForMessageId(*db, update.Message.ReplyToMessage.MessageID)
					if err != nil {
						log.Fatal(err.Error())
					}
					//replyUser := db.Where("message_id = ?", update.Message.ReplyToMessage.MessageID).First(&userRequest)
					msg := tgbotapi.NewMessage(replyUserRequest.ChatId, update.Message.Text)
					bot.Send(msg)
					continue
				}
				ownerGreeting := "Hello My Kid!"
				msg := tgbotapi.NewMessage(OwnerAcc, ownerGreeting)
				//msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
				continue
			}

			//msg = tgbotapi.NewForward(OwnerAcc, update.Message.From.ID, update.Message.MessageID)
			//msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}

//func handleOwnerMessage(update tgbotapi.Update) {
//}
//
//func handleCommand(update tgbotapi.Update) {
//}

//func openDb() *badger.DB {
//	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	return db
//}

func getUserRequestForMessageId(db gorm.DB, messageId int) (Request, error) {
	var userRequest Request

	res := db.Where("message_id = ?", messageId-1).First(&userRequest)
	if res.Error != nil {
		log.Fatal(res.Error.Error())
		return userRequest, res.Error
	}

	return userRequest, nil
}

//func findUser(chatID int64) bool {
//	find, err := db.Where("chat_id = ?", "chatID").First(&user)
//	if err != nil {
//		msg := tgbotapi.NewMessage(chatID, "Произошла ошибка! Бот может работать неправильно!")
//		telegramBot.API.Send(msg)
//	}
//	return find
//}
