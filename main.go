package main

import (
	"log"
	"os"
	"strconv"

	"notes-ai-bot/bot"
	"notes-ai-bot/llm"
	"notes-ai-bot/speechkit"
	"notes-ai-bot/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Инициализация логгера
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Загружаем переменные окружения из .env (если есть)
	_ = godotenv.Load()

	// 1) Синхронная загрузка VOSK ДО всего остального
	modelPath := os.Getenv("VOSK_MODEL_PATH")
	if modelPath == "" {
		modelPath = "models/vosk-model-small-ru-0.22"
	}
	if err := speechkit.Init(modelPath); err != nil {
		logger.Fatalf("❌ Ошибка инициализации VOSK: %v", err)
	}
	defer speechkit.Close()

	// 2) Инициализация БД
	db, err := storage.NewStorage("notes.db")
	if err != nil {
		logger.Fatalf("❌ Ошибка открытия БД: %v", err)
	}
	defer db.Close()

	// 3) Инициализация LLM
	llmClient, err := llm.NewClient(logger)
	if err != nil {
		logger.Fatalf("❌ Ошибка инициализации LLM-клиента: %v", err)
	}

	// 4) Параллелизм обработки сообщений Telegram
	concurrency := 8
	if v := os.Getenv("TELEGRAM_CONCURRENCY"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 {
			concurrency = n
		}
	}
	logger.Printf("⚙️ Параллелизм обработки сообщений Telegram: %d", concurrency)

	// 5) Инициализация бота
	b, err := bot.NewBot(logger, db, llmClient, bot.WithConcurrency(concurrency))
	if err != nil {
		logger.Fatalf("❌ Ошибка инициализации Telegram-бота: %v", err)
	}

	// 6) Запуск бота
	if err := b.Start(); err != nil {
		logger.Fatalf("❌ Ошибка запуска бота: %v", err)
	}
}
