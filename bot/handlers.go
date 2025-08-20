package bot

import (
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleMessage главная функция обработки сообщений
func HandleMessage(b *Bot, msg *tgbotapi.Message) {
	// Проверяем команды
	if HandleCommands(b, msg) {
		return
	}

	// Проверяем клавиатурные кнопки
	if HandleKeyboard(b, msg, b.db) {
		return
	}

	// Если текст начинается с префиксов клавиатуры, пропускаем обработку как заметки
	text := msg.Text
	if strings.HasPrefix(text, "🔍 ") || strings.HasPrefix(text, "❌ ") || strings.HasPrefix(text, "⛔ ") ||
		text == "Просмотр заметок" || text == "Удаление заметок" || text == "Скрыть" || text == "Назад" {
		return
	}

	// Обработка аудио, если есть
	transcribedText, err := ProcessAudio(b, msg, os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		SendMessage(b, msg.Chat.ID, err.Error())
		return
	}
	if transcribedText != "" {
		text = transcribedText
	}

	// Обработка текстовой заметки
	ProcessNote(b, msg, text, b.llmClient, b.db)
}
