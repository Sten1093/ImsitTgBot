package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"tgbot/database"
	"tgbot/parser"
)

func Bot() {
	err := godotenv.Load("./resources/metadata/token/test_bot_token_example.env")
	if err != nil {
		log.Fatal("Ошибка загрузки токена")
	}

	botToken := os.Getenv("BOT_TOKEN")
	dbPath := "./resources/data/users.sqlite"

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	userDAO := database.NewUserDAO(dbPath)
	defer userDAO.Close()

	users := make(map[int64]*database.User) // Локальный кеш пользователей

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		user, exists := users[chatID]

		if !exists {
			dbUser, err := userDAO.GetUser(chatID)
			if err == nil && dbUser != nil {
				user = dbUser
			} else {
				user = &database.User{ID: chatID, State: "hello"}
				err := userDAO.SaveUser(user)
				if err != nil {
					log.Printf("Ошибка при сохранении нового пользователя: %s", err)
				}
			}
			users[chatID] = user
		}

		log.Printf("Получено сообщение от %s: %s", update.Message.From.UserName, update.Message.Text)

		switch user.State {

		// выбор расписания/корпуса/препода
		case "hello":
			users[chatID] = user
			if update.Message.Text == "🗓Расписание🗓" {
				if user.EducationLevel == "" {
					user.State = "waiting_for_education"
					sendMessage(bot, chatID, "Выбери форму обучения", createEducationKeyboard)
				} else {
					user.State = "waiting_for_course"
					sendMessage(bot, chatID, "Выбери курс:", createCourseKeyboard)
				}
			} else if update.Message.Text == "👱‍♂️Найти препода👱" {
				user.State = "teacher"
				sendMessage(bot, chatID, "Напиши фамилию преподоваеля в формате (Иванов)", nil)
			} else if update.Message.Text == "🏢Найти корпус🏫" {
				user.State = "corpus_info"
				sendMessage(bot, chatID, "Напиши номер корпуса", createCorpusNum)
			} else if update.Message.Text == "/start" {
				sendMessage(bot, chatID, "Привет, Я бот для помощи тебе в твоем обучении!", createHelloKeyboard)
			} else {
				sendMessage(bot, chatID, "Используй клавиатуру", createHelloKeyboard)
			}

		case "waiting_for_education":
			user.EducationLevel = update.Message.Text
			switch user.EducationLevel {
			case "Высшее", "Среднее":
				user.State = "waiting_for_course"
				sendMessage(bot, chatID, "Выбери курс:", createCourseKeyboard)
			case "⬅️Назад":
				user.State = "hello"
				sendMessage(bot, chatID, "Попробуем снова", createHelloKeyboard)
			default:
				sendMessage(bot, chatID, "Используй для этого клавиатуру", createEducationKeyboard)
			}
			// Сохраняем пользователя в базе данных после изменения образования
			user.UserName = update.Message.From.UserName
			err := userDAO.SaveUser(user)
			if err != nil {
				log.Printf("Ошибка при сохранении пользователя в БД: %s", err)
			}
			users[chatID] = user

		//выбор курса
		case "waiting_for_course":
			user.Course = update.Message.Text
			if update.Message.Text == "⬅️Назад" {
				sendMessage(bot, chatID, "Попробуем еще раз", createHelloKeyboard)
				user.State = "hello"
			} else if user.Course == "🤓 1 курс" || user.Course == "😎 2 курс" || user.Course == "🧐 3 курс" || user.Course == "🎓 4 курс" {
				sendMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel))
				user.State = "select_group"
			} else {
				sendMessage(bot, chatID, "Нажми кнопочку на клавиатуре", createCourseKeyboard)
			}
		// выбор группы
		case "select_group":
			user.Group = update.Message.Text

			if update.Message.Text == "⬅️Назад" {
				user.State = "waiting_for_course"
				sendMessage(bot, chatID, "Выбери курс:", createCourseKeyboard)
			} else {
				if user.Format == "" {
					sendMessage(bot, chatID, "Выбери формат вывода", createPrintKeyboard)
					user.State = "select_format"
				} else {
					schedule := parser.Tab(user.Group, user.Format, user.EducationLevel)
					sendMessage(bot, chatID, schedule, createBackKeyboard)
					user.State = "waiting_for_return"
				}

			}
		// выбор формата вывода
		case "select_format":
			if update.Message.Text == "⬅️Назад" {
				user.State = "select_group"
				sendMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel))
			} else {
				user.Format = update.Message.Text
				schedule := parser.Tab(user.Group, user.Format, user.EducationLevel)
				sendMessage(bot, chatID, schedule, createBackKeyboard)
				user.State = "waiting_for_return"
				user.UserName = update.Message.From.UserName

				err := userDAO.SaveUser(user)
				if err != nil {
					log.Printf("Ошибка при сохранении пользователя в БД: %s", err)
				}
				// Удаляем пользователя из кеша, чтобы не хранить лишние данные
				delete(users, chatID)
			}
		// ожидание выбора возврата
		case "waiting_for_return":
			switch update.Message.Text {
			case "📚 Курс":
				user.State = "waiting_for_course"
				sendMessage(bot, chatID, "Выберите курс:", createCourseKeyboard)
			case "🏫 Группа":
				user.State = "select_group"
				sendMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel))
			case "📋 Вывод":
				user.State = "select_format"
				sendMessage(bot, chatID, "Выберите формат вывода:", createPrintKeyboard)
			case "🎓Образование":
				user.State = "waiting_for_education"
				sendMessage(bot, chatID, "Выберите форму обучения:", createEducationKeyboard)
			case "〽️Начало":
				user.State = "hello"
				sendMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard)
			default:
				sendMessage(bot, chatID, "Нажми кнопку на клавиатуре", createBackKeyboard)
			}

		// вывод учителя и его расписания
		case "teacher":
			surname := update.Message.Text
			// получение учителя из списка
			teacher := parser.FindTeacher(surname)

			if teacher == nil || teacher.Picture == "" {
				user.State = "hello"
				users[chatID] = user
				sendMessage(bot, chatID, "Преподователь "+surname+" не найден", createHelloKeyboard)
				break
			}
			// получение его пары на данный момент времени
			lesson, _ := parser.FindCurrentLessons(teacher.FileName)

			handleMediaGroupInfo(bot, chatID, teacher.Surname+teacher.Name+teacher.Text+lesson, teacher.Picture, "")
			sendMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard)
			user.State = "hello"
		// нужны фотки и описание корпусов
		case "corpus_info":
			switch update.Message.Text {
			case "1":
				handleMediaGroupInfo(bot, chatID, "Эот наш главный корпус\nОриентиром тут послужит огромная парковка(курилка)\nАдрес: Зиповская, д.5", "AgACAgIAAxkBAAIMGGezl2jvC2ayVe3mLgMYwXFasFu2AAIJ6jEbElWZSSNHfcCmQC5DAQADAgADeQADNgQ", "AgACAgIAAxkBAAIMGmezl3wqE1b3QeT8-dcyzK_3fV1uAAIK6jEbElWZSeHvxteiCh53AQADAgADeQADNgQ")
			case "2":
				handleMediaGroupInfo(bot, chatID, "Второй корпус или (Сбербанк)\nНаходиться на пересечении зиповской и московской. А наш ориентиир это компьютерный клуб Fenix\nАдрес: Зиповская 8", "AgACAgIAAxkBAAIMHGezl4EBvKI8oDrhd-1BowtoOdaXAAIL6jEbElWZSXNaIW1dEbN3AQADAgADeQADNgQ", "AgACAgIAAxkBAAIMHmezl4TKiWxTBBCpxje-a0mrZrFdAAIM6jEbElWZSanwqQY42zwPAQADAgADeQADNgQ")
			case "3":
				handleMediaGroupInfo(bot, chatID, "Третий корпус\nНаходиться за трамвайными путями по правой стороне\nАдрес: Зиповская 12", "AgACAgIAAxkBAAIMIGezl4ilhBlCPHXLOjJDgw0Jsl1WAAIN6jEbElWZSYM8NJomN7cjAQADAgADeQADNgQ", "AgACAgIAAxkBAAIMImezl4vq1hZcGzxjsHc2OhkR2T6GAAIO6jEbElWZSXtrNqSuGZE8AQADAgADeQADNgQ")
			case "4":
				handleMediaGroupInfo(bot, chatID, "Четвертый корпус\nНаш ориентиир это SubWay а точнее слева от него\nАдрес: Зиповская 5/2", "AgACAgIAAxkBAAIMJGezl41-B1WQR58j6hmA2UofC5KBAAIP6jEbElWZSR0J-RkHYC4uAQADAgADeQADNgQ", "AgACAgIAAxkBAAIMKGezl5MOOzeENlVFHs_4OBT7FDxxAAIQ6jEbElWZSVYDwTwZvotcAQADAgADeQADNgQ")
			case "5":
				handleMediaGroupInfo(bot, chatID, "Пятый корпус (школа)\nНаходиться в пристройке бывшей гимназии имсит. Но только не путай наш вход с торца а не главный\nАдрес: Зиповская 8", "AgACAgIAAxkBAAIMKmezl6KaBTFwpQ7yg2dfZmblKod6AAIR6jEbElWZSVPbb9zz5qnyAQADAgADeQADNgQ", "AgACAgIAAxkBAAIMLGezl6ZXWMGczOSPJM_u07L8yWMfAAIS6jEbElWZSYIvWM7Sy1obAQADAgADeQADNgQ")
			case "6":
				handleMediaGroupInfo(bot, chatID, "Шестой корпус(Дизайнеры)\nнаходиьтся за углом от мфц напротив входа в главный корпус\nАдрес: Зиповская 5к1", "AgACAgIAAxkBAAIMLmezl6keJdC81C60nlfGHnMTGCA2AAIT6jEbElWZSX8RQz9bs9sUAQADAgADeQADNgQ", "AgACAgIAAxkBAAIMMGezl6yQZVNzQilooWyezBIjyW8XAAIU6jEbElWZSRLQzWyvE7vuAQADAgADeQADNgQ")
			case "7":
				handleMediaGroupInfo(bot, chatID, "Седьмой корпус\nНаходиьтся сразу за главным", "AgACAgIAAxkBAAIMsGe0YjrPKP6gsXW3DirbQL5nnHggAAKl6DEbcu2oSfG8npWC0fCrAQADAgADeQADNgQ", "AgACAgIAAxkBAAIMrme0YddRwEUphXHrn57bARP2EVE0AAKd6DEbcu2oSW71UXvpQVQ3AQADAgADeQADNgQ")
			case "8":
				handleMediaGroupInfo(bot, chatID, "Восьмой корпус\nнаходитьтся слева от корпуса пять в здании гимназии\nАдрес: Зиповская 3", "AgACAgIAAxkBAAIMKmezl6KaBTFwpQ7yg2dfZmblKod6AAIR6jEbElWZSVPbb9zz5qnyAQADAgADeQADNgQ", "AgACAgIAAxkBAAIMNGezl7O4RORY0x7-mcw3Oxb1t0yIAAIV6jEbElWZSSclYr8g5AsbAQADAgADeQADNgQ")
			case "〽️Начало":
				user.State = "hello"
				sendMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard)
			default:
				sendMessage(bot, chatID, "Нажми цифру на клавиатуре", nil)
			}
		}
	}
}

func sendMessage(api *tgbotapi.BotAPI, chatID int64, text string, keyboardFunc func() tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	if keyboardFunc != nil {
		msg.ReplyMarkup = keyboardFunc()
	}
	api.Send(msg)
}

func getGroupKeyboard(course, education string) func() tgbotapi.ReplyKeyboardMarkup {
	switch education {
	case "Высшее":
		switch course {
		case "🤓 1 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(1)
			}
		case "😎 2 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(2)
			}
		case "🧐 3 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(3)
			}
		case "🎓 4 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(4)
			}
		}

	case "Среднее":
		switch course {
		case "🤓 1 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(7)
			}
		case "😎 2 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(8)
			}
		case "🧐 3 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(9)
			}
		case "🎓 4 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(10)
			}
		}

	default:
		return nil
	}
	return nil
}

func handleMediaGroupInfo(api *tgbotapi.BotAPI, chatID int64, text, fileID1, fileID2 string) {
	media1 := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID1))
	media1.Caption = text
	var mediaGroup []interface{}
	mediaGroup = append(mediaGroup, media1)

	// Проверяем, передан ли второй fileID
	if fileID2 != "" {
		media2 := tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID2))
		mediaGroup = append(mediaGroup, media2)
	}
	mediGroup := tgbotapi.NewMediaGroup(chatID, mediaGroup)

	api.Send(mediGroup)
}
