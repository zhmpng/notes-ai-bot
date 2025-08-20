package speechkit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

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
	Text    string `json:"text"`
	Partial string `json:"partial"`
}

func (c *Client) Transcribe(wavPath string) (string, error) {
	// Открываем WAV-файл
	file, err := os.Open(wavPath)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	// Подключаемся к WebSocket
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second
	fmt.Println("Попытка подключения к:", c.wsURL)
	conn, resp, err := dialer.Dial(c.wsURL, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка подключения к WebSocket: %v, HTTP статус: %v", err, resp)
	}
	defer conn.Close()

	// Устанавливаем таймауты
	conn.SetReadDeadline(time.Now().Add(300 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(300 * time.Second))

	// Отправляем config
	config := map[string]interface{}{
		"config": map[string]interface{}{
			"sample_rate": 16000,
			"language":    "ru-RU",
		},
	}
	cfgBytes, _ := json.Marshal(config)
	fmt.Println("Отправка конфигурации:", string(cfgBytes))
	if err := conn.WriteMessage(websocket.TextMessage, cfgBytes); err != nil {
		return "", fmt.Errorf("ошибка отправки config: %v", err)
	}

	// Отправляем аудиофайл по кускам
	buffer := make([]byte, 4000)
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("ошибка чтения файла: %v", err)
		}
		if n == 0 {
			break
		}

		fmt.Printf("Отправлено %d байт аудиоданных\n", n)
		if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
			return "", fmt.Errorf("ошибка отправки аудиоданных: %v", err)
		}
	}

	// Отправляем EOF
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Отправка EOF")
	if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"eof":1}`)); err != nil {
		return "", fmt.Errorf("ошибка отправки EOF: %v", err)
	}

	// Читаем ответы
	var transcription string
	timeout := time.After(30 * time.Second)
	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("таймаут ожидания ответа от сервера")
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseInternalServerErr) {
					return "", fmt.Errorf("соединение закрыто сервером: %v", err)
				}
				return "", fmt.Errorf("ошибка чтения ответа: %v", err)
			}

			fmt.Println("Получен ответ:", string(message))
			var response voskResponse
			if err := json.Unmarshal(message, &response); err != nil {
				return "", fmt.Errorf("ошибка разбора ответа: %v", err)
			}

			if response.Partial != "" {
				fmt.Println("partial:", response.Partial)
			}
			if response.Text != "" {
				fmt.Println("text:", response.Text)
				transcription = response.Text
				return transcription, nil
			}
		}
	}
}
