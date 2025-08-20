package speechkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gorilla/websocket"
)

type Client struct {
	wsURL string
}

func NewClient() (*Client, error) {
	wsURL := os.Getenv("VOSK_SERVER")
	if wsURL == "" {
		return nil, fmt.Errorf("VOSK_SERVER не установлен в переменных окружения")
	}
	return &Client{wsURL: wsURL}, nil
}

type voskResponse struct {
	Result string `json:"text"`
}

func (c *Client) Transcribe(wavPath string) (string, error) {
	// Открываем WAV-файл
	file, err := os.Open(wavPath)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	// Устанавливаем WebSocket соединение
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(c.wsURL, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка подключения к WebSocket: %v", err)
	}
	defer conn.Close()

	// Читаем и отправляем аудиоданные по частям
	buffer := make([]byte, 1024*16) // Буфер 16KB
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("ошибка чтения файла: %v", err)
		}
		if n == 0 {
			// Конец файла
			if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"eof": 1}`)); err != nil {
				return "", fmt.Errorf("ошибка отправки EOF: %v", err)
			}
			break
		}

		// Отправляем аудиоданные как бинарное сообщение
		if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
			return "", fmt.Errorf("ошибка отправки аудиоданных: %v", err)
		}
	}

	// Получаем ответ от сервера
	var transcription string
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return "", fmt.Errorf("ошибка чтения ответа: %v", err)
		}

		var response voskResponse
		if err := json.Unmarshal(message, &response); err != nil {
			return "", fmt.Errorf("ошибка разбора ответа: %v", err)
		}

		if response.Result != "" {
			transcription += response.Result + " "
		}

		// Если получен финальный результат
		if bytes.Contains(message, []byte(`"eof": 1`)) {
			break
		}
	}

	// Проверяем результат транскрипции
	if transcription == "" {
		return "", fmt.Errorf("ошибка получения ответа транскрипции")
	}

	return transcription, nil
}
