package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
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

type StringArray []string

type User struct {
	gorm.Model
	TelegramID int64 `gorm:"unique_index"`
	Name       string
	City       string
	Car        string
	Engine     string
	Photos     StringArray `gorm:"type:json"`
	State      int
	Status     int
}

// todo Сделать проверку того что отправляют в ответе, что бы текст был текстом, не стикер или эмодзи!!! Иначе возвращать на шаг назад
// todo Сделать кнопки для админа, которые будут принимать либо отклонять заявки

//// Request Сущность пользователя
//type Request struct {
//	Id     int
//	ChatId int64
//	Status int
//	Step   int
//}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	//Token := os.Getenv("TOKEN")
	adminChatId, _ := strconv.ParseInt(os.Getenv("OWNER_ACC"), 10, 64)
	SupergroupF30Id, _ := strconv.ParseInt(os.Getenv("SUPERGROUP_F30_ID"), 10, 64)
	//DSN := os.Getenv("DSN")

	db, err := gorm.Open(mysql.Open(os.Getenv("DSN")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	//// Create the 'users' table if it does not exist
	//if !db.Ta(&User{}) {
	//	db.CreateTable(&User{})
	//}

	db.AutoMigrate(&User{})

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// Ответы на вопросы
	//answer1 := ""
	//answer2 := ""
	//answer3 := ""
	//answer4 := ""

	// todo что то сделать с этими ссылками в статичных текстах
	userReplyPlease := "Будь ласка, дай відповідь на питання вище!"
	userWelcomeMsg := "Привіт, зараз я поставлю тобі кілька запитань!"
	userAlreadyDoneMsg := "Ваша заявку вже було розглянуто, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyiv"
	userWaitingMsg := "Наразі ваша заявка на розгляді, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyiv"
	userRejectMsg := "Вашу заявку було відхилено, для інформації зв'яжіться з адміністрацією. @fclubkyiv"
	userDoneReguestMsg := "Дякуємо, найближчим часом ви отримаєте посилання на чат. Якщо протягом тривалого часу ви не отримали посилання - зв'яжіться з адміністрацією - @fclubkyiv."
	userBannedMsg := "Ваша заявка була заблокована, якщо виникли питання - зв'яжіться з адміністрацією. @fclubkyiv"

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

	// test git branches
	for update := range updates {
		if update.Message != nil { // If we got a message
			// Пропускаем сообщения если они в из суперчата
			if update.Message.Chat.ID == SupergroupF30Id {
				continue
			}

			user, err := getUser(db, update.Message.Chat.ID)
			if err != nil {
				log.Println("Error getting user:", err)
				return
			}

			// Проверяем статус пользователя
			switch user.Status {
			case statusAccepted:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, userAlreadyDoneMsg)
				msg.ParseMode = parseModeHTMl
				bot.Send(msg)
			case statusRejected:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, userRejectMsg)
				msg.ParseMode = parseModeHTMl
				bot.Send(msg)
			case statusBanned:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, userBannedMsg)
				msg.ParseMode = parseModeHTMl
				bot.Send(msg)
			case statusWaiting:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, userWaitingMsg)
				msg.ParseMode = parseModeHTMl
				bot.Send(msg)
			case statusNew:
				switch user.State {
				case stateInitial:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
					msg.Text = userWelcomeMsg
					bot.Send(msg)
					msg.Text = askUserName
					bot.Send(msg)
					user.State = stateName
					updateUser(db, user)
				case stateName:
					update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
					userMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
					if update.Message.Text == "" {
						// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
						userMsg.Text = userReplyPlease
						bot.Send(userMsg)
						// todo не уверен что тут нужно continue
						continue
					}

					// todo в каждом сообщении нужно убирать смайлы и проверять не пустая ли строка
					// Отправляем пользователю следующий вопрос
					user.Name = update.Message.Text
					userMsg.Text = askUserCity
					bot.Send(userMsg)

					// Обновляем состояние пользователя
					user.State = stateCity
					updateUser(db, user)
				case stateCity:
					update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
					userMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
					if update.Message.Text == "" {
						// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
						userMsg.Text = userReplyPlease
						bot.Send(userMsg)
						// todo не уверен что тут нужно continue
						continue
					}

					// todo в каждом сообщении нужно убирать смайлы и проверять не пустая ли строка
					// Отправляем пользователю следующий вопрос
					user.City = update.Message.Text
					userMsg.Text = askUserCar
					bot.Send(userMsg)

					// Обновляем состояние пользователя
					user.State = stateCar
					updateUser(db, user)
				case stateCar:
					update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
					userMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
					if update.Message.Text == "" {
						// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
						userMsg.Text = userReplyPlease
						bot.Send(userMsg)
						// todo не уверен что тут нужно continue
						continue
					}

					// todo в каждом сообщении нужно убирать смайлы и проверять не пустая ли строка
					// Отправляем пользователю следующий вопрос
					user.Car = update.Message.Text
					userMsg.Text = askUserEngine
					bot.Send(userMsg)

					// Обновляем состояние пользователя
					user.State = stateEngine
					updateUser(db, user)
				case stateEngine:
					update.Message.Text = gomoji.RemoveEmojis(update.Message.Text)
					userMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
					if update.Message.Text == "" {
						// Если не ответ не пришел в нормальном формате, просим ещё раз ответить
						userMsg.Text = userReplyPlease
						bot.Send(userMsg)
						// todo не уверен что тут нужно continue
						continue
					}

					// todo в каждом сообщении нужно убирать смайлы и проверять не пустая ли строка
					// Отправляем пользователю следующий вопрос
					user.Engine = update.Message.Text
					userMsg.Text = askUserPhoto
					bot.Send(userMsg)

					// Обновляем состояние пользователя
					user.State = statePhoto
					updateUser(db, user)
				case statePhoto:
					if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
						// Получаем первое фото из слайса
						photoID := (update.Message.Photo)[1].FileID

						// Добавляем фото в фото пользователя
						user.Photos = append(user.Photos, photoID)

						// Отправляем сообщение пользователю
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Фото було успішно додано, завантаж ще, або натисни <b>Готово</b>.")
						msg.ParseMode = parseModeHTMl
						msg.ReplyMarkup = doneButton
						bot.Send(msg)

						// Обновляем пользователя в базе данных
						updateUser(db, user)
					} else if update.Message.Text == "Готово👌" {
						// Проверяем что бы у пользователя было загружено хоть одно фото
						if len(user.Photos) == 0 {
							// Если фото нет - отправляем уведомление
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ви не завантажили жодного фото!")
							msg.ReplyMarkup = doneButton
							bot.Send(msg)
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
							update.Message.From.ID)

						// Сообщение администратору
						adminMsg := tgbotapi.NewMessage(adminChatId, adminMsgText)
						adminMsg.ReplyMarkup = requestButtons
						bot.Send(adminMsg)

						// Формируем галерею с комментарием
						files := make([]interface{}, len(user.Photos))
						caption := fmt.Sprintf("ChatID: %d", update.Message.Chat.ID)
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
							adminChatId,
							files,
						)

						if _, err := bot.SendMediaGroup(cfg); err != nil {
							log.Panic(err)
						}

						// Отправляем сообщение пользователю
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, userDoneReguestMsg)
						msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						bot.Send(msg)

						// Сбрасываем состояние пользователя
						user.State = stateCompleted
						user.Status = statusWaiting
						user.Photos = nil
						updateUser(db, user)
					} else {
						// Просим пользователя загрузить фото
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, askUserPhoto)
						bot.Send(msg)
					}
				}
			}
		} else if update.CallbackQuery != nil {
			// разбиваем сообщение на котором висят кнопки (сама заявка админа) на массив
			s := strings.Fields(update.CallbackQuery.Message.Text)

			// В нашем случае последний элемент массива будет chat_id (string)
			strUserID := s[len(s)-1]

			// Преобразовываем строку в число и получаем числовой `chat_id` пользователя отправившего заявку
			userChatID, _ := strconv.ParseInt(strUserID, 10, 64)

			user, err := getUser(db, userChatID)
			if err != nil {
				log.Println("Error getting user:", err)
				return
			}

			adminMsg := tgbotapi.NewMessage(adminChatId, "")
			switch user.Status {
			case statusAccepted:
				adminMsg.Text = fmt.Sprintf("Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>Прийнято</b>.", userChatID)
				adminMsg.ParseMode = parseModeHTMl
				bot.Send(adminMsg)
				// todo не уверен что нужно `continue`
				continue
			case statusRejected:
				adminMsg.Text = fmt.Sprintf("Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>Відхилено</b>.", userChatID)
				adminMsg.ParseMode = parseModeHTMl
				bot.Send(adminMsg)
				// todo не уверен что нужно `continue`
				continue
			case statusBanned:
				adminMsg.Text = fmt.Sprintf("Користувач був розглянутий! \n Поточний статус користувача з ID: %d - <b>Заблоковано</b>.", userChatID)
				adminMsg.ParseMode = parseModeHTMl
				bot.Send(adminMsg)
				// todo не уверен что нужно `continue`
				continue
			case statusWaiting:
				// Получаем данные из колбека
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
				userMsg := tgbotapi.NewMessage(userChatID, "")
				// todo переменная выше уже объявлена
				adminMsg := tgbotapi.NewMessage(adminChatId, "")

				// Действия админа по отношению к заявке
				switch callback.Text {
				case callbackAccept:
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
					userMsg.Text = "Привіт!\nТвої відповіді стосовно вступу в <b>F-club Kyiv</b> були оброблені нашою командою. Ознайомся з простими умовами спілкування в нашому клубі та приєднуйся до нас! \n\n1. Поважай інших учасників. Нецензурна лайка, цькування, використання непристойних стікерів - заборонено(але якщо це в тему, то всі розуміють😂)\n2. Не влаштовуємо «Барахолку»! Ти можешь запропонувати, якщо в тебе є щось корисне для автомобіля, чи будь що, але не треба про це писати кожного дня і робити рекламні оголошення. \n3. Якщо ти хочеш запропонувати свої послугу(сто, детейлінг, автомийки, итд) - повідом про це адміністрації і зробіть гарне оголошення разом - це все безкоштовно !! \n 4. Ми розуміємо, що зараз без цього ніяк, але маємо про це попросити - якомога менше суперечок стосовно політики. Ми всі підтримуємо Україну і не шукаємо зради!\n 5. Стосовно використання GIF , ми не проти цього, але не треба постити дуже багато, один за одним! \n 6. Май повагу до інших власників автомобілів, не у кожного така гарна машина, як в тебе!  \n\nМаєш бажання отримати клубний стікер на авто чи номерну рамку - відпиши на це повідомлення\U0001FAE1\n\nТримай посилання, для вступу в чат!\n     P.s.Не забудь привітатися з нових товаришами, та розповісти який в тебе автомобіль!\n\n\n\nДонати для розвитку!(За бажанням) \n\nF-Club Kyiv \n\n🎯Ціль: 100 000.00 ₴\n\n🔗Посилання на банку\nhttps://send.monobank.ua/jar/S87zLF6xL\n\n💳Номер картки банки\n5375 4112 0304 9692"
					userMsg.ParseMode = parseModeHTMl
					bot.Send(userMsg)

					userMsg.Text = fmt.Sprintf("Ось ваше <a href=\"%s\">%s</a>", chatInviteLink.InviteLink, chatInviteLink.Name)
					userMsg.ParseMode = parseModeHTMl
					bot.Send(userMsg)

					// todo Обновляем статусы пользователя (принять в группу)
					user.Status = statusAccepted
					updateUser(db, user)

					// Ответное сообщение администратору
					adminMsg.Text = fmt.Sprintf("Користувача з <b>ChatID: %d</b> підтверджено, посилання на вступ до групи надіслано!", userChatID)
					adminMsg.ParseMode = parseModeHTMl
					bot.Send(adminMsg)
				case callbackReject:
					// Обновляем статус пользователя
					user.Status = statusRejected
					updateUser(db, user)

					// Отравляем уведомление пользователю
					userMsg.Text = userRejectMsg
					userMsg.ParseMode = parseModeHTMl
					bot.Send(userMsg)

					// todo вынести в константу
					// Отправляем уведомление админу
					adminMsg.Text = "Користувач був успішно відхилений!"
					bot.Send(adminMsg)
				case callbackBanned:
					// Обновляем статус пользователя
					user.Status = statusBanned
					updateUser(db, user)

					// Отравляем уведомление пользователю
					userMsg.Text = userBannedMsg
					userMsg.ParseMode = parseModeHTMl
					bot.Send(userMsg)

					// todo вынести в константу
					// todo переменная выше уже объявлена
					// Отправляем уведомление админу
					adminMsg.Text = "Користувач був успішно заблокованний!"
					bot.Send(adminMsg)
				}
			}
		}
	}
}

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal StringArray value: %v", value)
	}

	return json.Unmarshal(b, &a)
}

func getUser(db *gorm.DB, telegramID int64) (*User, error) {
	var user User
	if err := db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = User{
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

func updateUser(db *gorm.DB, user *User) {
	if err := db.Save(user).Error; err != nil {
		log.Printf("Error updating user: %s", err)
	}
}
