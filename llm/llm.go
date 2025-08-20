package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"notes-ai-bot/models"
)

type Client struct {
	url   string // URL локального сервера LM Studio
	model string
}

type NoteResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewClient(logger *log.Logger) (*Client, error) {
	url := os.Getenv("LLM_URL")
	if url == "" {
		logger.Printf("❌ LM_URL не установлен в переменных окружения")
		return nil, fmt.Errorf("LM_URL не установлен в переменных окружения")
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		logger.Printf("❌ LM_MODEL не установлен в переменных окружения")
		return nil, fmt.Errorf("LM_MODEL не установлен в переменных окружения")
	}
	return &Client{url: url, model: model}, nil
}

func (c *Client) ClassifyAndProcessNote(text string, logger *log.Logger) (string, string, error) {
	promptTemplate := os.Getenv("PROMPT")
	if promptTemplate == "" {
		logger.Printf("❌ PROMPT не установлен в переменных окружения")
		return "", "", fmt.Errorf("PROMPT не установлен в переменных окружения")
	}

	// Декодируем \n в реальные переносы строк
	promptTemplate = strings.ReplaceAll(promptTemplate, `\n`, "\n")
	fullPrompt := fmt.Sprintf(promptTemplate, text)

	// Формируем запрос
	req := models.ChatRequest{
		Model: c.model,
		Messages: []models.ChatMessage{
			{Role: "user", Content: fullPrompt},
		},
		MaxTokens: 512, // Ограничиваем длину ответа
	}

	data, err := json.Marshal(req)
	if err != nil {
		logger.Printf("⚠️ Не удалось преобразовать запрос в JSON: %v", err)
		return "", "", fmt.Errorf("не удалось преобразовать запрос в JSON: %w", err)
	}
	logger.Printf("🧠 PROMPT в LM Studio:\n%s", fullPrompt)

	// Выполняем запрос
	resp, err := http.Post(fmt.Sprintf("%s/v1/chat/completions", c.url), "application/json", bytes.NewBuffer(data))
	if err != nil {
		logger.Printf("⚠️ Не удалось отправить запрос: %v", err)
		return "", "", fmt.Errorf("не удалось отправить запрос: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("⚠️ Не удалось прочитать тело ответа: %v", err)
		return "", "", fmt.Errorf("не удалось прочитать тело ответа: %w", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		logger.Printf("⚠️ Ошибка сервера: %s, тело ответа: %s", resp.Status, string(body))
		return "", "", fmt.Errorf("ошибка сервера: %s", resp.Status)
	}

	// Парсим ответ
	var result models.ChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Printf("⚠️ Не удалось разобрать ответ LLM: %v, сырое тело: %s", err, string(body))
		return "", "", fmt.Errorf("не удалось разобрать ответ LLM: %w", err)
	}

	if len(result.Choices) == 0 {
		logger.Printf("⚠️ В ответе LLM нет вариантов: %s", string(body))
		return "", "", fmt.Errorf("в ответе LLM нет вариантов")
	}

	// Очищаем ответ от <think> тегов
	answer := removeThinkBlock(result.Choices[0].Message.Content)
	logger.Printf("✅ Ответ LM Studio:\n%s", answer)

	// Парсим текст ответа как JSON
	var note NoteResponse
	if err := json.Unmarshal([]byte(answer), &note); err != nil {
		logger.Printf("⚠️ Не удалось разобрать ответ заметки: %v, сырой текст: %s", err, answer)
		return "", "", fmt.Errorf("не удалось разобрать ответ заметки: %w", err)
	}

	logger.Printf("✅ Распознанная заметка: type=%s, message=%s", note.Type, note.Message)
	return note.Type, note.Message, nil
}

func removeThinkBlock(text string) string {
	start := strings.Index(text, "<think>")
	end := strings.Index(text, "</think>")

	if start >= 0 && end > start {
		return strings.TrimSpace(text[:start] + text[end+len("</think>"):])
	}
	return text
}
