package bot

import (
	"fmt"
	"os"
	"path/filepath"

	speechkit "notes-ai-bot/speechkit"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ProcessAudio обрабатывает голосовые или видео заметки, возвращает транскрибированный текст
func ProcessAudio(b *Bot, msg *tgbotapi.Message, token string) (string, error) {
	var fileID string
	if msg.Voice != nil {
		fileID = msg.Voice.FileID
	} else if msg.VideoNote != nil {
		fileID = msg.VideoNote.FileID
	} else {
		return "", nil // Не аудио
	}

	// Скачивание файла
	file, err := b.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		b.logger.Printf("❌ Ошибка получения файла: %v", err)
		return "", fmt.Errorf("ошибка при скачивании голосового сообщения")
	}

	downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, file.FilePath)
	oggPath := filepath.Join(os.TempDir(), file.FileID+".ogg")
	wavPath := filepath.Join(os.TempDir(), file.FileID+".wav")

	// Скачивание OGG
	err = DownloadFile(downloadURL, oggPath)
	if err != nil {
		b.logger.Printf("❌ Ошибка скачивания файла ogg: %v", err)
		return "", fmt.Errorf("ошибка при скачивании")
	}

	// Конвертация в WAV
	err = ConvertOggToWav(oggPath, wavPath)
	if err != nil {
		b.logger.Printf("❌ Ошибка при конвертации в формат wav: %v", err)
		return "", fmt.Errorf("ошибка при конвертации аудио")
	}

	// Транскрипция
	transcribedText, err := speechkit.Transcribe(wavPath)
	if err != nil {
		b.logger.Printf("❌ Ошибка при транскрипции аудио: %v", err)
		return "", fmt.Errorf("ошибка при расшифровке голоса")
	}

	// Удаление временных файлов
	os.Remove(oggPath)
	os.Remove(wavPath)

	return transcribedText, nil
}
