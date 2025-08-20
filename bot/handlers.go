package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"notes-ai-bot/llm"
	"notes-ai-bot/speechkit"
	"notes-ai-bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleMessage обрабатывает текстовые и голосовые сообщения
func HandleMessage(b *Bot, msg *tgbotapi.Message, db *storage.Storage, llmClient *llm.Client, speechClient *speechkit.Client, token string) {
	userID := msg.From.ID
	text := msg.Text

	// Обработка специальных команд
	if msg.IsCommand() {
		cmd := msg.Command()
		if cmd == "start" || cmd == "menu" {
			SendMainMenu(b, msg.Chat.ID)
			return
		}
		if cmd == "hide" {
			SendMessageWithoutKeyboard(b, msg.Chat.ID, "Клавиатура скрыта.")
			return
		}
		SendMessage(b, msg.Chat.ID, "Неизвестная команда. Используйте /start, /menu или /hide.")
		return
	}

	// Обработка пользовательской клавиатуры
	if text != "" {
		// Команды главного меню
		switch text {
		case "Просмотр заметок":
			SendCategoriesMenu(b, msg.Chat.ID, "Выберите категорию для просмотра:", "🔍")
		case "Удаление заметок":
			SendCategoriesMenu(b, msg.Chat.ID, "Выберите категорию для удаления:", "❌")
		case "Скрыть":
			SendMessageWithoutKeyboard(b, msg.Chat.ID, "Клавиатура скрыта.")
			return
		case "Назад":
			SendMainMenu(b, msg.Chat.ID)
			return
		}

		// Обработка категорий
		if strings.HasPrefix(text, "🔍 ") {
			noteType := strings.TrimPrefix(text, "🔍 ")
			sendNotesList(b, msg.Chat.ID, userID, noteType, db, false)
			return
		}
		if strings.HasPrefix(text, "❌ ") {
			noteType := strings.TrimPrefix(text, "❌ ")
			sendNotesList(b, msg.Chat.ID, userID, noteType, db, true)
			return
		}

		// Удаление конкретной заметки (например, "⛔ 1")
		if strings.HasPrefix(text, "⛔ ") {
			idStr := strings.TrimPrefix(text, "⛔ ")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				b.logger.Printf("❌ Неверный ID заметки: %v", err)
				SendMessage(b, msg.Chat.ID, "Неверный ID.")
				return
			}
			err = db.DeleteNote(userID, id)
			if err != nil {
				b.logger.Printf("❌ Ошибка при удалении заметки: %v", err)
				SendMessage(b, msg.Chat.ID, "Ошибка при удалении или заметка не найдена.")
				return
			}
			SendMessageWithoutKeyboard(b, msg.Chat.ID, fmt.Sprintf("Заметка с ID %d удалена.", id))
			SendMainMenu(b, msg.Chat.ID)
			return
		}
	}

	// Обработка голосовых сообщений
	if msg.Voice != nil || msg.VideoNote != nil {
		var fileID string
		if msg.Voice != nil {
			fileID = msg.Voice.FileID
		} else {
			fileID = msg.VideoNote.FileID
		}

		// Скачивание файла
		file, err := b.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
		if err != nil {
			b.logger.Printf("❌ Ошибка получения файла: %v", err)
			SendMessage(b, msg.Chat.ID, "Ошибка при скачивании голосового сообщения.")
			return
		}

		downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, file.FilePath)
		oggPath := filepath.Join(os.TempDir(), file.FileID+".ogg")
		wavPath := filepath.Join(os.TempDir(), file.FileID+".wav")

		// Скачивание OGG
		err = DownloadFile(downloadURL, oggPath)
		if err != nil {
			b.logger.Printf("❌ Ошибка скачивания файла ogg: %v", err)
			SendMessage(b, msg.Chat.ID, "Ошибка при скачивании.")
			return
		}

		// Конвертация в WAV
		err = ConvertOggToWav(oggPath, wavPath)
		if err != nil {
			b.logger.Printf("❌ Ошибка при конвертации в формат wav: %v", err)
			SendMessage(b, msg.Chat.ID, "Ошибка при конвертации аудио.")
			return
		}

		// Транскрипция
		transcribedText, err := speechClient.Transcribe(wavPath)
		if err != nil {
			b.logger.Printf("❌ Ошибка при транскрипции аудио: %v", err)
			SendMessage(b, msg.Chat.ID, "Ошибка при расшифровке голоса.")
			return
		}

		text = transcribedText
		// Удаление временных файлов
		os.Remove(oggPath)
		os.Remove(wavPath)
	}

	if text == "" {
		return
	}

	// Классификация и структурирование через LLM
	noteType, noteMessage, err := llmClient.ClassifyAndProcessNote(text, b.logger)
	if err != nil {
		b.logger.Printf("❌ Ошибка при классификации заметки: %v", err)
		SendMessage(b, msg.Chat.ID, "Ошибка при классификации или обработке заметки.")
		return
	}

	// Сохранение заметки
	id, err := db.AddNote(userID, noteType, noteMessage)
	if err != nil {
		b.logger.Printf("❌ Ошибка при сохранении заметки: %v", err)
		SendMessage(b, msg.Chat.ID, "Ошибка при сохранении заметки.")
		return
	}

	// Подтверждение и показ меню
	SendMessageWithMenu(b, msg.Chat.ID, fmt.Sprintf("Заметка сохранена! Тип: %s, ID: %d\nТекст: %s", noteType, id, noteMessage))
}
