package bot

import (
	"fmt"
	"notes-ai-bot/storage"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendMessage отправляет простое сообщение
func SendMessage(b *Bot, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.Send(msg); err != nil {
		b.logger.Printf("❌ Ошибка отправки сообщения: %v", err)
	}
}

// SendMessageWithKeyboard отправляет сообщение с пользовательской клавиатурой
func SendMessageWithKeyboard(b *Bot, chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = &keyboard
	if _, err := b.Send(msg); err != nil {
		b.logger.Printf("❌ Ошибка отправки сообщения с клавиатурой: %v", err)
	}
}

// SendMessageWithoutKeyboard отправляет сообщение без клавиатуры
func SendMessageWithoutKeyboard(b *Bot, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err := b.Send(msg); err != nil {
		b.logger.Printf("❌ Ошибка отправки сообщения без клавиатуры: %v", err)
	}
}

// SendMessageWithMenu отправляет подтверждение с главным меню
func SendMessageWithMenu(b *Bot, chatID int64, text string) {
	SendMessageWithKeyboard(b, chatID, text, GetMainKeyboard())
}

// SendMainMenu отправляет главное меню
func SendMainMenu(b *Bot, chatID int64) {
	SendMessageWithKeyboard(b, chatID, "Выберите действие:", GetMainKeyboard())
}

// sendNotesList отправляет список заметок категории (с кнопками удаления, если deleteMode=true)
func sendNotesList(b *Bot, chatID int64, userID int64, noteType string, db *storage.Storage, deleteMode bool) {
	notes, err := db.GetNotes(userID, noteType)
	if err != nil {
		b.logger.Printf("❌ Ошибка получения заметок: %v", err)
		SendMessage(b, chatID, "❌ Ошибка при получении списка заметок.")
		return
	}
	if len(notes) == 0 {
		SendMessage(b, chatID, "⚠️ Нет заметок в категории "+noteType+".")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Заметки в категории %s:\n", noteType))
	var rows [][]tgbotapi.KeyboardButton

	for _, note := range notes {
		// Форматируем дату в читаемый вид: YYYY-MM-DD HH:MM:SS
		formattedDate := note.CreatedAt.Format("2006-01-02 15:04:05")
		// Выводим поля на отдельных строках
		sb.WriteString(fmt.Sprintf("ID: %d\nТекст: %s\nДата: %s\n", note.ID, note.Text, formattedDate))
		if deleteMode {
			rows = append(rows, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(fmt.Sprintf("⛔ %d", note.ID)),
			))
		}
	}

	if deleteMode {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		))
		keyboard := tgbotapi.NewReplyKeyboard(rows...)
		keyboard.ResizeKeyboard = true
		SendMessageWithKeyboard(b, chatID, sb.String(), keyboard)
	} else {
		SendMessage(b, chatID, sb.String())
	}

	// Показываем главное меню после, если не в режиме удаления
	if !deleteMode {
		SendMainMenu(b, chatID)
	}
}
