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
	"log"
	"os"
	"path/filepath"
	"sync"
	"unsafe"
)

var (
	model     *C.VoskModel
	initOnce  sync.Once
	closeOnce sync.Once
	initErr   error
)

// Init загружает VOSK-модель синхронно. Вызвать нужно 1 раз перед началом работы бота.
// Путь к модели передается в виде строки, например: "models/vosk-model-small-ru-0.22".
// Если путь не указан, вернет ошибку initErr.
// Если модель уже загружена, повторный вызов ничего не сделает.
func Init(modelDir string) error {
	initOnce.Do(func() {
		if modelDir == "" {
			initErr = fmt.Errorf("путь к модели VOSK не указан")
			return
		}
		cpath := C.CString(filepath.Clean(modelDir))
		defer C.free(unsafe.Pointer(cpath))

		log.Printf("🔄 Загрузка модели VOSK из: %s ...", modelDir)
		m := C.vosk_model_new(cpath)
		if m == nil {
			initErr = fmt.Errorf("не удалось загрузить модель VOSK из: %s", modelDir)
			return
		}
		model = m
		log.Printf("✅ Модель VOSK успешно загружена")
	})
	return initErr
}

// Close освобождает модель (по завершении приложения).
func Close() {
	closeOnce.Do(func() {
		if model != nil {
			log.Printf("🔻 Освобождение ресурсов модели VOSK")
			C.vosk_model_free(model)
			model = nil
		}
	})
}

// Transcribe — синхронная транскрипция wav-файла (16kHz mono s16le).
// Путь к файлу передается в виде строки, например: "audio.wav".
// Возвращает текстовую расшифровку или ошибку.
func Transcribe(wavPath string) (string, error) {
	if model == nil {
		return "", fmt.Errorf("модель VOSK не инициализирована — вызовите speechkit.Init до запуска бота")
	}

	log.Printf("🎤 Начало транскрипции: %s", wavPath)
	f, err := os.Open(wavPath)
	if err != nil {
		return "", fmt.Errorf("не удалось открыть WAV-файл: %w", err)
	}
	defer f.Close()

	// 16000 Гц — под конвертацию из utils.go
	rec := C.vosk_recognizer_new(model, 16000.0)
	if rec == nil {
		return "", fmt.Errorf("не удалось создать распознаватель VOSK")
	}
	defer C.vosk_recognizer_free(rec)

	buf := make([]byte, 4000)
	for {
		n, er := f.Read(buf)
		if n > 0 {
			C.vosk_recognizer_accept_waveform(rec, (*C.char)(unsafe.Pointer(&buf[0])), C.int(n))
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			return "", fmt.Errorf("ошибка чтения WAV-файла: %w", er)
		}
	}

	finalJSON := C.GoString(C.vosk_recognizer_final_result(rec))
	var jres map[string]interface{}
	if err := json.Unmarshal([]byte(finalJSON), &jres); err != nil {
		return "", fmt.Errorf("ошибка разбора финального результата VOSK (JSON): %w", err)
	}
	text, _ := jres["text"].(string)
	log.Printf("📝 Результат транскрипции: %q", text)
	return text, nil
}
