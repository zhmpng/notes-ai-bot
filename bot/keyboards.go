package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// categories определяет фиксированные категории заметок
var categories = []string{"список дел", "список покупок", "запомни", "идеи/мысли", "прочее"}

// GetMainKeyboard возвращает главное меню в виде пользовательской клавиатуры
func GetMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Просмотр заметок"),
			tgbotapi.NewKeyboardButton("Удаление заметок"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Скрыть"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// SendCategoriesMenu отправляет меню категорий в виде пользовательской клавиатуры
func SendCategoriesMenu(b *Bot, chatID int64, text string, actionPrefix string) {
	var rows [][]tgbotapi.KeyboardButton

	for _, cat := range categories {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(fmt.Sprintf("%s %s", actionPrefix, cat)),
		)
		rows = append(rows, row)
	}

	// Добавляем кнопку "Назад"
	backRow := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Назад"),
	)
	rows = append(rows, backRow)

	// Создаём клавиатуру
	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.ResizeKeyboard = true // Подстраивает размер клавиатуры под контент

	SendMessageWithKeyboard(b, chatID, text, keyboard)
}
