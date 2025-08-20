package bot

import (
	"fmt"
	"strconv"
	"strings"

	"notes-ai-bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleKeyboard обрабатывает нажатия на кнопки клавиатуры
func HandleKeyboard(b *Bot, msg *tgbotapi.Message, db *storage.Storage) bool {
	text := msg.Text
	userID := msg.From.ID
	chatID := msg.Chat.ID

	switch {
	case text == "Просмотр заметок":
		SendCategoriesMenu(b, chatID, "Выберите категорию для просмотра:", "🔍")
		return true
	case text == "Удаление заметок":
		SendCategoriesMenu(b, chatID, "Выберите категорию для удаления:", "❌")
		return true
	case text == "Скрыть":
		SendMessageWithoutKeyboard(b, chatID, "Клавиатура скрыта.")
		return true
	case text == "Назад":
		SendMainMenu(b, chatID)
		return true
	case strings.HasPrefix(text, "🔍 "):
		noteType := strings.TrimPrefix(text, "🔍 ")
		sendNotesList(b, chatID, userID, noteType, db, false)
		return true
	case strings.HasPrefix(text, "❌ "):
		noteType := strings.TrimPrefix(text, "❌ ")
		sendNotesList(b, chatID, userID, noteType, db, true)
		return true
	case strings.HasPrefix(text, "⛔ "):
		idStr := strings.TrimPrefix(text, "⛔ ")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			b.logger.Printf("❌ Неверный ID заметки: %v", err)
			SendMessage(b, chatID, "Неверный ID.")
			return true
		}
		err = db.DeleteNote(userID, id)
		if err != nil {
			b.logger.Printf("❌ Ошибка при удалении заметки: %v", err)
			SendMessage(b, chatID, "Ошибка при удалении или заметка не найдена.")
			return true
		}
		SendMessageWithoutKeyboard(b, chatID, fmt.Sprintf("Заметка с ID %d удалена.", id))
		SendMainMenu(b, chatID)
		return true
	}
	return false
}
