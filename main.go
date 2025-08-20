package main

import (
	"log"
	"os"

	"notes-ai-bot/bot"
	"notes-ai-bot/llm"
	"notes-ai-bot/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Инициализация логгера
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Загрузка переменных окружения
	if err := godotenv.Load(".env"); err != nil {
		logger.Fatalf("❌ Не удалось загрузить .env: %v", err)
	}

	// Инициализация хранилища
	db, err := storage.NewStorage("notes.db")
	if err != nil {
		logger.Fatalf("❌ Ошибка инициализации DB: %v", err)
	}
	defer db.Close()

	// Инициализация LLM клиента
	llmClient, err := llm.NewClient(logger)
	if err != nil {
		logger.Fatalf("❌ Ошибка инициализации LLM client: %v", err)
	}

	// Инициализация бота
	bot, err := bot.NewBot(logger, db, llmClient)
	if err != nil {
		logger.Fatalf("❌ Ошибка инициализации телеграм бота: %v", err)
	}

	// Запуск бота
	if err := bot.Start(); err != nil {
		logger.Fatalf("❌ Ошибка запуска бота: %v", err)
	}
}
