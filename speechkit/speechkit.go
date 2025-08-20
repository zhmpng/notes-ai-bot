package speechkit

/*
#cgo CFLAGS: -IC:/Repos/notes-ai-bot/vosk_lib
#cgo LDFLAGS: -LC:/Repos/notes-ai-bot/vosk_lib -lvosk
#include "vosk_api.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	vosk "github.com/alphacep/vosk-api/go"
)

// Transcribe распознаёт речь из WAV-файла с помощью локальной модели Vosk
func Transcribe(filename string) (string, error) {
	// Загрузка модели (папка "model" должна существовать)
	model, err := vosk.NewModel("Z:/vosk-model-ru-0.42/vosk-model-ru-0.42")
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки модели: %v", err)
	}

	// Создание распознавателя с sampleRate 16 kHz
	sampleRate := 16000.0
	rec, err := vosk.NewRecognizer(model, sampleRate)
	if err != nil {
		return "", fmt.Errorf("ошибка создания распознавателя: %v", err)
	}
	defer rec.Free()
	rec.SetWords(1)

	// Открытие WAV-файла
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	buf := make([]byte, 4096)

	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("ошибка чтения файла: %v", err)
		}

		if rec.AcceptWaveform(buf[:n]) != 0 {
			// Можно вывести промежуточный результат
			fmt.Println(rec.Result())
		}
	}

	// Финальный результат
	finalJSON := rec.FinalResult()
	var jres map[string]interface{}
	if err := json.Unmarshal([]byte(finalJSON), &jres); err != nil {
		return "", fmt.Errorf("ошибка разбора JSON: %v", err)
	}

	text, _ := jres["text"].(string)
	return text, nil
}
