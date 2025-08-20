package bot

import (
	"errors"
	"log"
	"os"

	"notes-ai-bot/llm"
	"notes-ai-bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// ErrNoToken возвращается, если TELEGRAM_TOKEN не установлен.
var ErrNoToken = errors.New("TELEGRAM_TOKEN не установлен")

// Bot представляет Telegram-бота и его состояние.
type Bot struct {
	bot       *tgbotapi.BotAPI // Telegram bot API клиент
	logger    *log.Logger      // Логгер (дублирует вывод в файл и консоль)
	db        *storage.Storage // Хранилище заметок
	llmClient *llm.Client      // Клиент для работы с LLM
}

// NewBot создаёт и инициализирует нового бота с загрузкой конфигурации из .env.
func NewBot(logger *log.Logger, db *storage.Storage, llmClient *llm.Client) (*Bot, error) {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		logger.Printf("⚠️ Не удалось загрузить .env файл: %v", err)
	}

	// Получаем токен бота из переменной окружения
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		logger.Printf("❌ TELEGRAM_TOKEN не установлен")
		return nil, ErrNoToken
	}

	// Инициализируем Telegram bot API
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Printf("❌ Ошибка инициализации Telegram Bot API: %v", err)
		return nil, err
	}

	// Отключаем отладочный режим
	bot.Debug = false
	logger.Printf("✅ Авторизован как %s", bot.Self.UserName)

	// Возвращаем новый экземпляр бота
	return &Bot{
		bot:       bot,
		logger:    logger,
		db:        db,
		llmClient: llmClient,
	}, nil
}

// Send отправляет сообщение через Telegram Bot API.
func (b *Bot) Send(msg tgbotapi.Chattable) (tgbotapi.Message, error) {
	message, err := b.bot.Send(msg)
	if err != nil {
		b.logger.Printf("❌ Ошибка отправки сообщения: %v", err)
		return tgbotapi.Message{}, err
	}
	return message, nil
}

// Start запускает бота, настраивая получение обновлений через polling.
func (b *Bot) Start() error {
	// Настраиваем канал обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u) // GetUpdatesChan возвращает только канал

	// Запускаем обработку обновлений
	b.StartPolling(updates)
	return nil
}

// StartPolling запускает обработку входящих обновлений Telegram.
func (b *Bot) StartPolling(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		HandleMessage(b, update.Message)
	}
}
