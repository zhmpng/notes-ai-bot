# Notes AI Bot — интеллектуальный помощник заметок (текст + голос)

Бот помогает превращать поток мыслей в лаконичные, структурированные заметки.  
Он **принимает текст и голосовые сообщения**, делает **распознавание речи (Vosk, офлайн)**, затем **классифицирует заметку по типу** и формирует её по **шаблону**, выделяя суть.

---

## 🚀 Возможности

- ✍️ **Текст → заметка**: понимает свободный текст и переводит в аккуратную заметку по шаблону.
- 🗣️ **Голос → текст → заметка**: принимает голосовые в Telegram, конвертирует в WAV 16 кГц и распознаёт офлайн через Vosk.
- 🧠 **Классификация по типам**: определяет тип (задача, идея, решение, заметка-встречи, дневник и т.п.) и применяет нужный шаблон.
- ✂️ **Выжимка смысла**: из «литра воды» извлекает «стакан сути» — кратко и по делу.
- 🧩 **Шаблоны и логика по типам**: у каждого типа свои поля и правила заполнения.
- 💾 **Хранилище**: записи можно сохранять в выбранное хранилище (см. `storage/`).

> В репозитории присутствуют директории `bot/` (Telegram-логика), `speechkit/` (Vosk), `llm/` (классификация/шаблоны), `models/`, `storage/`, `vosk_lib/` — см. структуру ниже.

---

## 📦 Структура проекта

```
notes-ai-bot/
 ├── bot/           # Telegram-бот: загрузка файлов, обработка команд/сообщений
 ├── llm/           # Классификация, суммаризация, применение шаблонов
 ├── models/        # Описания типов/шаблонов заметок, вспомогательные структуры
 ├── speechkit/     # Интеграция с Vosk (CGO + libvosk), конвертация и распознавание
 ├── storage/       # Сохранение заметок (файлы/БД/интеграции — в зависимости от реализации)
 ├── vosk_lib/      # Windows-библиотеки Vosk: libvosk.dll, vosk_api.h, *.lib, зависимости
 ├── main.go        # Точка входа
 ├── go.mod         # Go-модуль
 └── README.md
```

---

## 🛠 Требования

- **Go 1.24+**
- **Windows**: MSYS2 (mingw-w64) для `gcc` (нужно для CGO)
- **Vosk**: Windows library (`vosk-win64-*.zip`) + русская модель (например, `vosk-model-ru-0.42` или «small» вариант)
- (Опционально) **FFmpeg** — если используете внешнюю утилиту для конвертации аудио
- Токен Telegram-бота

---

## ⚙️ Переменные окружения

Обязательные/возможные (в зависимости от конфигурации проекта):

```bash
# Telegram
TELEGRAM_BOT_TOKEN=123456789:AA...

# Путь к модели Vosk (рекомендуется)
VOSK_MODEL_PATH=C:/vosk-model-ru-0.42

# --- CGO / libvosk для Windows ---
# (указывает компилятору, где искать vosk_api.h и libvosk.dll/lib)
CGO_ENABLED=1
CGO_CFLAGS=-IC:/Repos/notes-ai-bot/vosk_lib
CGO_LDFLAGS=-LC:/Repos/notes-ai-bot/vosk_lib -lvosk

# (опционально) Ускорители LLM/провайдеры — если в llm/ используется внешний API
# OPENAI_API_KEY=sk-...
# ANTHROPIC_API_KEY=...
# LLM_PROVIDER=openai|anthropic|local
```

> В `speechkit/` путь к модели может быть захардкожен. Рекомендуется читать его из `VOSK_MODEL_PATH` (если ещё не сделано).

---

## 🧩 Установка зависимостей (Windows)

1) Установите [MSYS2](https://www.msys2.org/) и пакет компилятора:
```bash
pacman -S mingw-w64-x86_64-gcc
```

2) Добавьте mingw64 в PATH (PowerShell):
```powershell
$Env:PATH = $Env:PATH + ";C:\msys64\mingw64\bin"
```

3) Разместите файлы из `vosk-win64-*.zip` в `./vosk_lib/` проекта:
```
vosk_lib/
 ├── libvosk.dll
 ├── libvosk.lib
 ├── vosk_api.h
 ├── libgcc_s_seh-1.dll
 ├── libstdc++-6.dll
 └── libwinpthread-1.dll
```

4) Скачайте и распакуйте модель, например в `C:\vosk-model-ru-0.42\`.

5) Экспортируйте переменные окружения (PowerShell):
```powershell
$Env:CGO_ENABLED="1"
$Env:CGO_CFLAGS="-IC:\Repos\notes-ai-bot\vosk_lib"
$Env:CGO_LDFLAGS="-LC:\Repos\notes-ai-bot\vosk_lib -lvosk"
$Env:VOSK_MODEL_PATH="C:\vosk-model-ru-0.42"
$Env:PATH = $Env:PATH + ";C:\msys64\mingw64\bin"
```

6) Установите Go-зависимости:
```bash
go mod tidy
```

---

## 🏃 Быстрый старт (Windows, PowerShell)

```powershell
# 1) Клонируем
git clone https://github.com/zhmpng/notes-ai-bot
cd notes-ai-bot

# 2) Настраиваем окружение (см. раздел выше)
$Env:CGO_ENABLED="1"
$Env:CGO_CFLAGS="-I$PWD\vosk_lib"
$Env:CGO_LDFLAGS="-L$PWD\vosk_lib -lvosk"
$Env:VOSK_MODEL_PATH="C:\vosk-model-ru-0.42"
$Env:TELEGRAM_BOT_TOKEN="<ваш_токен>"

# 3) Сборка и запуск
go build -v
go run main.go
```

> При старте Vosk выводит длинные `LOG (...)` — это **нормально**: загружается большая модель (особенно `ru-0.42`). Для ускорения используйте «small»-модель и/или перенесите модель на SSD.

---

## 🐧 Запуск на Linux (альтернатива)

```bash
# Загрузка и распаковка libvosk для Linux x86_64
wget https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-linux-x86_64-0.3.45.zip
unzip vosk-linux-x86_64-0.3.45.zip
export VOSK_PATH=$PWD/vosk-linux-x86_64-0.3.45

# Пути для компоновки/инклюда
export CGO_ENABLED=1
export CGO_CPPFLAGS="-I $VOSK_PATH"
export CGO_LDFLAGS="-L $VOSK_PATH -lvosk -lpthread -ldl"
export LD_LIBRARY_PATH=$VOSK_PATH:$LD_LIBRARY_PATH

# Путь к модели
export VOSK_MODEL_PATH="$HOME/models/vosk-model-ru-0.42"

# Сборка/запуск
go mod tidy
go build -v
./notes-ai-bot
```

---

## 🔄 Конвейер обработки

1. **Получение сообщения** (текст/голос) в Telegram.
2. **Голос → WAV 16 кГц** (конвертация из OGG/Opus).
3. **Распознавание Vosk** (офлайн, через `speechkit/` и `libvosk`).
4. **Классификация** в `llm/` → выбор **типа заметки**.
5. **Шаблонизация**: заполнение полей по логике типа.
6. **Сохранение** в `storage/` (файлы/БД/интеграции — по реализации).
7. **Ответ пользователю**: краткая выжимка + оформленная заметка.

---

## 🧪 Проверка распознавания отдельно

Чтобы исключить ошибки окружения, можно вызвать распознавание напрямую (пример):  
```go
text, err := speechkit.Transcribe("path/to/file.wav")
if err != nil { /* обработать ошибку */ }
fmt.Println(text)
```
> Убедитесь, что WAV — 16 кГц, моно. Иначе приведите к нужному формату.

---

## 🧯 Разбор частых проблем

- **`fatal error: vosk_api.h: No such file or directory`**  
  CGO не видит заголовок. Проверьте `CGO_CFLAGS` и содержимое `vosk_lib/`.

- **`cgo: C compiler "gcc" not found`**  
  Установите MSYS2 и добавьте `C:\msys64\mingw64\bin` в `PATH`.

- **Долго «висит» на старте с логами Vosk**  
  Это загрузка модели. Используйте «small»-модель или перенесите модель на SSD/локальный диск.

- **Приложение не запускается: отсутствует DLL**  
  Убедитесь, что `libvosk.dll` лежит рядом с `.exe` или путь к нему есть в `PATH`.

---

## 📜 Лицензия

MIT. Подробности см. в `LICENSE`.

---

## 🙏 Благодарности

- [Vosk](https://alphacephei.com/vosk/) — офлайн ASR
- Сообщества Go и Telegram за отличные библиотеки и SDK
