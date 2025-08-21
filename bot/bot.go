package bot

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"notes-ai-bot/llm"
	"notes-ai-bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// ErrNoToken возвращается, если TELEGRAM_TOKEN не установлен.
var ErrNoToken = errors.New("TELEGRAM_TOKEN не установлен")

type Bot struct {
	api         *tgbotapi.BotAPI
	logger      *log.Logger
	db          *storage.Storage
	llmClient   *llm.Client
	concurrency int
}

type Option func(*Bot)

// WithConcurrency настраивает количество параллельных воркеров обработки сообщений.
func WithConcurrency(n int) Option {
	return func(b *Bot) { b.concurrency = n }
}

func NewBot(logger *log.Logger, db *storage.Storage, llmClient *llm.Client, opts ...Option) (*Bot, error) {
	_ = godotenv.Load()
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, ErrNoToken
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		api:         api,
		logger:      logger,
		db:          db,
		llmClient:   llmClient,
		concurrency: 8, // значение по умолчанию
	}
	for _, o := range opts {
		o(b)
	}

	return b, nil
}

func (b *Bot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.api.Send(c)
}

func (b *Bot) Start() error {
	b.logger.Printf("🚀 Запуск Telegram-бота (@%s)", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	// Грейсфул-завершение
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go b.startPollingAsync(updates)

	<-stop
	b.logger.Println("🛑 Остановка бота по сигналу")
	return nil
}

// Асинхронная обработка входящих обновлений с ограничением параллелизма.
func (b *Bot) startPollingAsync(updates tgbotapi.UpdatesChannel) {
	sem := make(chan struct{}, b.concurrency)
	b.logger.Printf("🔄 Обработка обновлений запущена (параллельность: %d)", b.concurrency)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		sem <- struct{}{}
		go func(u tgbotapi.Update) {
			defer func() { <-sem }()
			b.logger.Printf("📩 Сообщение от @%s (%d): тип=%s",
				u.Message.From.UserName, u.Message.From.ID, messageKind(u.Message))
			HandleMessage(b, u.Message)
		}(update)
	}
	b.logger.Println("ℹ️ Канал обновлений Telegram закрыт")
}

// Вспомогательная функция для красивых логов
func messageKind(m *tgbotapi.Message) string {
	switch {
	case m.Text != "":
		return "текст"
	case m.Voice != nil:
		return "voice"
	case m.Audio != nil:
		return "audio"
	case m.VideoNote != nil:
		return "videonote"
	default:
		return "другое"
	}
}
