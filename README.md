# 📝 Notes AI Bot  

> Telegram-бот для создания и управления заметками с поддержкой распознавания речи 🎤 и генерации текста 🤖  

---

## 📖 Описание  

`Notes AI Bot` — это удобный ассистент, который позволяет:  
- ✍️ Создавать и управлять заметками прямо в Telegram  
- 🎤 Добавлять голосовые заметки (бот преобразует их в текст с помощью **VOSK**)  
- 🤖 Использовать LLM для интеллектуальной обработки текста (генерация, резюме и др.)  
- 📂 Хранить все заметки в SQLite базе данных  

---

## 🏗 Архитектура  

```
├── main.go                   # Точка входа
├── bot/                      # Логика работы с Telegram API
│   ├── bot.go                # Инициализация и запуск бота
│   ├── handlers.go           # Обработка сообщений
│   ├── audio\_processing.go  # Работа с голосовыми сообщениями
│   └── utils.go              # Утилиты (скачивание файлов, ffmpeg и т.д.)
├── speechkit/                # Интеграция с VOSK (распознавание речи)
│   └── speechkit.go
├── llm/                      # Работа с LLM API
├── storage/                  # Хранилище заметок (SQLite)
├── models/                   # Каталог для VOSK-модели
└── .env                      # Конфигурация окружения
```

---

## ⚙️ Используемые технологии  

- [Go 1.21+](https://go.dev/) — основной язык разработки  
- [Telegram Bot API](https://core.telegram.org/bots/api) — взаимодействие с Telegram
- [VOSK](https://alphacephei.com/vosk/) — офлайн-распознавание речи  
- [SQLite](https://www.sqlite.org/) — локальное хранилище заметок  
- [FFmpeg](https://ffmpeg.org/) — конвертация аудио  

---

## 👤 Пользовательские сценарии  

- 📌 Отправка текстового сообщения → Добавление текстовой заметки  
- 🎙️ Отправка голосового сообщения → конвертация в текст → добавление текстовой заметки
- 📖 Просмотр списка заметок  
- 🗑️ Удаление заметок 
- 🤖 Интеллектуальная обработка текста через LLM  
### Будущие доработки:
- 🔍 Поиск заметок по ключевым словам
- 📅 Просмотр заметок за определённый период
- 📂 Экспорт заметок в текстовый файл
- 📥 Импорт заметок из текстового файла
- 📊 Статистика по заметкам (количество, частота использования и т.д.)
- 🔒 Защита заметок паролем
- 📑 Создание категорий для заметок
- 📅 Напоминания о важных заметках
- 📜 История изменений заметок
- 

---

## 🛠 Установка  

### 1. Скачивание репозитория  

```bash
git clone https://github.com/yourname/notes-ai-bot.git
cd notes-ai-bot
````

### 2. Установка зависимостей

```bash
go mod tidy
```

### 3. Скачивание VOSK модели

Перейдите на [страницу моделей VOSK](https://alphacephei.com/vosk/models) и скачайте русскую модель:

```
vosk-model-small-ru-0.22
```

Распакуйте её в папку `models/` так, чтобы путь был:

```
models/vosk-model-small-ru-0.22
```

---

## ⚡ Настройка окружения

Создайте файл `.env` в корне проекта:

```env
TELEGRAM_TOKEN=ваш_токен_бота
TELEGRAM_CONCURRENCY=8
VOSK_MODEL_PATH=models/vosk-model-small-ru-0.22
LLM_URL=http://127.0.0.1:1234
LLM_MODEL=local
PROMPT=
```

* `TELEGRAM_TOKEN` — токен бота от [BotFather](https://t.me/BotFather)
* `VOSK_MODEL_PATH` — путь к папке с моделью
* `TELEGRAM_CONCURRENCY` — сколько сообщений обрабатывается параллельно
* `LLM_URL` — Адрес LLM API (если используется)
* `LLM_MODEL` — Модель LLM (например, `local` для локального сервера)
* `PROMPT` — Промпт для LLM (если используется)

---

## 🖥 Настройка GoLand IDE

1. Откройте проект в **GoLand**
2. Перейдите: `Run → Edit Configurations`
3. В вашем `Go Build` конфиге укажите:

    * **Working Directory**: путь к корню проекта (`notes-ai-bot`)
    * **Environment variables**:

      ```
      CGO_ENABLED=1;
      CC=C:\msys64\mingw64\bin\gcc.exe;
      CXX=C:\msys64\mingw64\bin\g++.exe;
      CGO_CFLAGS=-IZ:/Repos/notes-ai-bot/vosk_lib;
      CGO_LDFLAGS=-LZ:/Repos/notes-ai-bot/vosk_lib -lvosk;
      PATH=%PATH%;C:\msys64\mingw64\bin;C:\ffmpeg\bin
      ```
4. Запустите проект кнопкой ▶ или через `Debug` 🐞

---

## 🚀 Запуск

### Через Go:

```bash
go run main.go
```

### Через бинарник:

```bash
go build -o notes-ai-bot.exe
./notes-ai-bot.exe
```

---

## ✅ Пример логов при запуске

```
🔄 Загрузка модели VOSK из: models/vosk-model-small-ru-0.22 ...
✅ Модель VOSK успешно загружена
🚀 Запуск Telegram-бота (@my_notes_bot)
🔄 Обработка обновлений запущена (параллельность: 8)
📩 Сообщение от @username (123456): тип=voice
📝 Результат транскрипции: "Привет, это тестовая заметка"
```

---

## 📦 Что ещё нужно

* Установленный [FFmpeg](https://ffmpeg.org/download.html) (чтобы конвертировать OGG → WAV)
* Установленный MinGW (`gcc.exe` / `g++.exe`) для сборки с CGO

---

## ❤️ Благодарности

* [VOSK](https://alphacephei.com/vosk/) за офлайн-распознавание речи
* [Go Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api)

---

## 🛡 Лицензия

Этот проект распространяется под лицензией MIT. Используйте и дорабатывайте свободно!


