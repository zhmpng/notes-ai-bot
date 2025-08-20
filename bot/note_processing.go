package bot

import (
	"fmt"

	"notes-ai-bot/llm"
	"notes-ai-bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ProcessNote классифицирует, обрабатывает и сохраняет заметку
func ProcessNote(b *Bot, msg *tgbotapi.Message, text string, llmClient *llm.Client, db *storage.Storage) {
	if text == "" {
		return
	}

	userID := msg.From.ID
	chatID := msg.Chat.ID

	// Классификация и структурирование через LLM
	noteType, noteMessage, err := llmClient.ClassifyAndProcessNote(text, b.logger)
	if err != nil {
		b.logger.Printf("❌ Ошибка при классификации заметки: %v", err)
		SendMessage(b, chatID, "Ошибка при классификации или обработке заметки.")
		return
	}

	// Сохранение заметки
	id, err := db.AddNote(userID, noteType, noteMessage)
	if err != nil {
		b.logger.Printf("❌ Ошибка при сохранении заметки: %v", err)
		SendMessage(b, chatID, "Ошибка при сохранении заметки.")
		return
	}

	// Подтверждение и показ меню
	SendMessageWithMenu(b, chatID, fmt.Sprintf("Заметка сохранена! Тип: %s, ID: %d\nТекст: %s", noteType, id, noteMessage))
}
