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
	DSN := os.Getenv("DSN")

	bot, err := tgbotapi.NewBotAPI(Token)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	db, err := gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Ответы на вопросы
	answer1 := ""
	answer2 := ""
	answer3 := ""
	answer4 := ""

	// Вопросы
	question1 := "Як тебе звати?"
	question2 := "З якого ти міста?"
	question3 := "Яке в тебе авто?"
	question4 := "Який двигун?"
	question5 := "Надійшли фото автомобіля, щоб було видно державний номер авто - після натисни «ГОТОВО»\nЯкщо вважаєш за необхідне приховати номерний знак - це твоє право, але ми повинні розуміти, що ти з України та тобі можна довіряти."

	// todo что то сделать с этими ссылками в статичных текстах
	sendReplyPlease := "Будь ласка, дай відповідь на питання вище!"
	welcomeMsg := "Привіт, зараз я поставлю тобі кілька запитань!"
	userAlreadyDoneMsg := "Ваша заявку вже було розглянуто, якщо виникли питання - зв'яжіться з <a href=\"tg://user?id=6225178130\">адміністратором</a>."
	userWaitingMsg := "Наразі ваша заявка на розгляді, якщо виникли питання - зв'яжіться з <a href=\"tg://user?id=6225178130\">адміністратором</a>."
	rejectRequestMsg := "Вашу заявку було відхилено, для інформації зв'яжіться з <a href=\"tg://user?id=6225178130\">адміністратором</a>."
	wellDoneMessage := "Дякуємо, найближчим часом ви отримаєте посилання на чат. Якщо протягом тривалого часу ви не отримали посилання - зв'яжіться з адміністрацією - @fclubkyiv."

	// Кнопка готов
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

	// Массив ИД файлов для отправки
	var answerFileIds []string = nil
	var isDocumentFiles = false
	var isPhotoFiles = false

	for update := range updates {

		if update.Message != nil { // If we got a message
			// Пропускаем сообщения если они в из суперчата
			if update.Message.Chat.ID == SupergroupF30Id {
				continue
			}

			// Ид текущего чата с пользователем (UserID)
			chatID := update.Message.Chat.ID
			// Создаём пустое сообщение адресованное пользователю
			msg := tgbotapi.NewMessage(chatID, "")
			// Создаём пустой запрос(пользователя)
			var userRequest Request

			// @todo тестируем сообщения от конкретного человека
			//if update.Message.From.ID == 123 {
			//}

			// Проверка команда ли это?
			//if update.Message.IsCommand() {
			//	msg := tgbotapi.NewMessage(chatID, "")
			//
			//	switch update.Message.Command() {
			//	case "restart":
			//		msg.Text = "Удаляю пользователя с ID: " + strconv.FormatInt(chatID, 10) + "..."
			//		if _, err := bot.Send(msg); err != nil {
			//			log.Panic(err)
			//		}
			//		db.Where("chat_id = ?", chatID).Delete(&Request{})
			//
			//		msg.Text = "Пользователь удален. Можете попробовать заново создать заявку отправив любое сообщение в бот."
			//		if _, err := bot.Send(msg); err != nil {
			//			log.Panic(err)
			//		}
			//		continue
			//	case "start":
			//		// Стартовая команда которая начинает диалог
			//		msg.Text = welcomeMsg
			//		if _, err := bot.Send(msg); err != nil {
			//			log.Panic(err)
			//		}
			//	default:
			//		msg.Text = "Неизвестная команда"
			//		if _, err := bot.Send(msg); err != nil {
			//			log.Panic(err)
			//		}
			//		continue
			//	}
			//}

			// Проверка пользователя на существование
			result := db.Where("chat_id = ?", chatID).First(&userRequest)
			if result.RowsAffected > 0 { // Есть ли пользователь в БД?
				// Если есть пользователь, проверяем его статус
				switch userRequest.Status {
				case StatusAccepted:
					// Пользователь уже зарегистрирован и добавлен в группу
					msg.Text = userAlreadyDoneMsg
					msg.ParseMode = "HTML"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					log.Printf("Пользователь уже принят: [%d]", userRequest.ChatId)
					continue
				case StatusDeclined:
					// Пользователь отклонён
					msg.Text = userAlreadyDoneMsg
					msg.ParseMode = "HTML"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					log.Printf("Пользователь уже отклонён: [%d]", userRequest.ChatId)
					continue
				case StatusWaiting:
					msg.Text = userWaitingMsg
					msg.ParseMode = "HTML"
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
				msg.Text = welcomeMsg
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
				// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
				update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
				if update.Message.Text == "" {
					// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
					msg.Text = sendReplyPlease
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					// Задаём предыдущий вопрос
					msg.Text = question1
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}

				// Записываем ответ на вопрос
				answer1 = update.Message.Text
				// Задаём следующий вопрос
				msg.Text = question2
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				// Переводим пользователя на следующий шаг
				userRequest.Step = userRequest.Step + 1
				db.Save(&userRequest)
				continue
			case 2:
				// todo Вынести в отдельный метод.
				// Проверяем ответ пользователя на emoji
				// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
				update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
				if update.Message.Text == "" {
					msg.Text = sendReplyPlease
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					// Задаём предыдущий вопрос
					msg.Text = question2
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}
				// Записываем ответ на вопрос
				answer2 = update.Message.Text
				// Задаём следующий вопрос
				msg.Text = question3
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				// Переводим пользователя на следующий шаг
				userRequest.Step = userRequest.Step + 1
				db.Save(&userRequest)
				continue
			case 3:
				// todo Вынести в отдельный метод.
				// Проверяем ответ пользователя на emoji
				update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
				if update.Message.Text == "" {
					// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
					msg.Text = sendReplyPlease
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					// Задаём предыдущий вопрос
					msg.Text = question3
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}
				// Записываем ответ на вопрос
				answer3 = update.Message.Text
				// Задаём следующий вопрос
				msg.Text = question4
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				// Переводим пользователя на следующий шаг
				userRequest.Step = userRequest.Step + 1
				db.Save(&userRequest)
				continue
			case 4:
				// todo Вынести в отдельный метод.
				// Проверяем ответ пользователя на emoji
				update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
				if update.Message.Text == "" {
					// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
					msg.Text = sendReplyPlease
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					// Задаём предыдущий вопрос
					msg.Text = question4
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}
				// Записываем ответ на вопрос
				answer4 = update.Message.Text
				// Задаём следующий вопрос
				msg.Text = question5
				// Отправляем кнопку "Готово"
				msg.ReplyMarkup = doneButton
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				// Переводим пользователя на следующий шаг
				userRequest.Step = userRequest.Step + 1
				db.Save(&userRequest)
				continue
			case 5:
				// Проверяем что бы ответ пользователя были фото
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
						msg.Text = question5
						msg.ReplyMarkup = doneButton
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
						continue
					}
					msg.Text = wellDoneMessage
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}

					// Обновляем статусы
					userRequest.Step = 6
					userRequest.Status = StatusWaiting
					db.Save(&userRequest)
					// Готово👌

					// Отправка заявки админу
					totalAnswer := "Ім'я: " + answer1 + " \n"
					totalAnswer += "Місто: " + answer2 + " \n"
					totalAnswer += "Авто: " + answer3 + " \n"
					totalAnswer += "Двигун: " + answer4 + " \n"
					totalAnswer += "ChatID: " + strconv.FormatInt(chatID, 10) + " \n"

					// Отправляем текст заявки
					// Добавляем кнопки для отправки принятия
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
					// Если пришел какой-то текст кроме "готово", отправляем ещё раз вопрос о фото
					msg.Text = question5
					msg.ReplyMarkup = doneButton
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					continue
				}
			}

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
				replText := "Користувач був розглянутий! \n"
				replText += "Поточний статус користувача з ChatID: " +
					strconv.FormatInt(requestUserChatID, 10) + " - Принятий!"

				msg := tgbotapi.NewMessage(OwnerAcc, replText)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				continue
			} else if user.Status == StatusDeclined {
				replText := "Користувач був розглянутий! \n"
				replText += "Поточний статус користувача з ChatID: " +
					strconv.FormatInt(requestUserChatID, 10) + " - Відхилений!"

				msg := tgbotapi.NewMessage(OwnerAcc, replText)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				continue
			} else if user.Status == StatusBanned {
				replText := "Користувач був розглянутий! \n"
				replText += "Поточний статус користувача з ChatID: " +
					strconv.FormatInt(requestUserChatID, 10) + " - Заблокований!"

				msg := tgbotapi.NewMessage(OwnerAcc, replText)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				continue
			}

			// Если команда fuck_off_dog, обрабатываем и шлём сообщение клиенту оставившему заявку
			if callback.Text == "fuck_off_dog" {
				// Блокируем пользователя
				replText := "Вибачте, Ваша заявка була заблокована!\n"
				replText += "У разі виникнення питань – зв'яжіться з <a href=\"tg://user?id=6225178130\">адміністратором</a>."
				msg := tgbotapi.NewMessage(requestUserChatID, replText)
				msg.ParseMode = "HTML"
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				// todo обновлять статус, вероятно сделать новый, что-то типа "бана"
				user.Status = StatusBanned
				db.Save(&user)

				// Ответное сообщение администратору
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Користувач був успішно заблокованний!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			} else if callback.Text == "accept_request" {
				// Создаём конфиг для ссылки на вступление в группу
				inviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{
					ChatConfig: tgbotapi.ChatConfig{
						ChatID: SupergroupF30Id,
					},
					Name:               "посилання на групу",
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
				replyText := "Привіт!\nТвої відповіді стосовно вступу в <b>F-club Kyiv</b> були оброблені нашою командою. Ознайомся з простими умовами спілкування в нашому клубі та приєднуйся до нас! \n\n1. Поважай інших учасників. Нецензурна лайка, цькування, використання непристойних стікерів - заборонено(але якщо це в тему, то всі розуміють😂)\n2. Не влаштовуємо «Барахолку»! Ти можешь запропонувати, якщо в тебе є щось корисне для автомобіля, чи будь що, але не треба про це писати кожного дня і робити рекламні оголошення. \n3. Якщо ти хочеш запропонувати свої послугу(сто, детейлінг, автомийки, итд) - повідом про це адміністрації і зробіть гарне оголошення разом - це все безкоштовно !! \n 4. Ми розуміємо, що зараз без цього ніяк, але маємо про це попросити - якомога менше суперечок стосовно політики. Ми всі підтримуємо Україну і не шукаємо зради!\n 5. Стосовно використання GIF , ми не проти цього, але не треба постити дуже багато, один за одним! \n 6. Май повагу до інших власників автомобілів, не у кожного така гарна машина, як в тебе!  \n\nМаєш бажання отримати клубний стікер на авто чи номерну рамку - відпиши на це повідомлення\U0001FAE1\n\nТримай посилання, для вступу в чат!\n     P.s.Не забудь привітатися з нових товаришами, та розповісти який в тебе автомобіль!\n\n\n\nДонати для розвитку!(За бажанням) \n\nF-Club Kyiv \n\n🎯Ціль: 100 000.00 ₴\n\n🔗Посилання на банку\nhttps://send.monobank.ua/jar/S87zLF6xL\n\n💳Номер картки банки\n5375 4112 0304 9692"
				msg := tgbotapi.NewMessage(requestUserChatID, replyText)
				msg.ParseMode = "HTML"
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}

				respText := "Ось ваше <a href=\"" + chatInviteLink.InviteLink + "\">" + chatInviteLink.Name + "</a>\n"
				msg = tgbotapi.NewMessage(requestUserChatID, respText)
				msg.ParseMode = "HTML"
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
				// todo Обновляем статусы пользователя (принять в группу)
				user.Status = StatusAccepted
				db.Save(&user)

				// Ответное сообщение администратору
				// Вероятно нужно сюда выводить chat_id, что бы понять кого приняли в группу
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Користувач підтверджено, посилання на вступ до групи надіслано!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			} else if callback.Text == "reject_request" {
				// todo что то придумать тут
				respText := rejectRequestMsg
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
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Заявку користувача відхилено, інформацію надіслано!")
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
			}

			// Отправка колбека обратно
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}
		}
	}
}
