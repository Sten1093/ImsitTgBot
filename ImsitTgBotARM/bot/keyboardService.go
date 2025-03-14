package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgbot/parser"
)

const firstCourseName = "🤓 1 курс"
const seconndCourseName = "😎 2 курс"
const thirdCourseName = "🧐 3 курс"
const fourthCourseName = "🎓 4 курс"
const backName = "⬅️Назад"

// клавиатуры
func createHelloKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🗓Расписание🗓"),
			tgbotapi.NewKeyboardButton("👱‍♂️Найти препода👱"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏢Найти корпус🏫"),
		),
	)
}
func createBackKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📚 Курс"),
			tgbotapi.NewKeyboardButton("🏫 Группа"),
			tgbotapi.NewKeyboardButton("📋 Вывод"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎓Образование"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("〽️Начало"),
		),
	)
}
func createPrintKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🌞 День 🙋‍♂️"),
			tgbotapi.NewKeyboardButton("📅 Неделя"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(backName),
		),
	)
}
func createCourseKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(firstCourseName),
			tgbotapi.NewKeyboardButton(seconndCourseName),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(thirdCourseName),
			tgbotapi.NewKeyboardButton(fourthCourseName),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(backName),
		),
	)
}
func createCorpusNum() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("1"),
			tgbotapi.NewKeyboardButton("2"),
			tgbotapi.NewKeyboardButton("3"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("4"),
			tgbotapi.NewKeyboardButton("5"),
			tgbotapi.NewKeyboardButton("6"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("7"),
			tgbotapi.NewKeyboardButton("8"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("〽️Начало"),
		),
	)
}
func createGroupKeyboardCourseById(id int) tgbotapi.ReplyKeyboardMarkup {
	groups := parser.GetGroups()

	// Создание двумерного среза для хранения строк клавиатуры
	var keyboard [][]tgbotapi.KeyboardButton

	// Количество кнопок в строке
	const buttonsPerRow = 3

	// Временная строка для накопления кнопок
	var row []tgbotapi.KeyboardButton

	// Проходим по каждой группе
	for _, group := range groups {
		// Добавляем только те группы, у которых id == 1
		if group.ID == id {
			row = append(row, tgbotapi.NewKeyboardButton(group.TGName))

			// Если накопили нужное количество кнопок, добавляем строку
			if len(row) == buttonsPerRow {
				keyboard = append(keyboard, row)
				row = []tgbotapi.KeyboardButton{}
			}
		}
	}

	// Добавляем оставшиеся кнопки, если они есть
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	// Добавляем строку с кнопкой "⬅️Назад"
	keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("⬅️Назад"),
	))

	// Возвращаем клавиатуру
	return tgbotapi.NewReplyKeyboard(keyboard...)
}
func createEducationKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Высшее"),
			tgbotapi.NewKeyboardButton("Среднее"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⬅️Назад"),
		),
	)
}
