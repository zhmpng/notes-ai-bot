## 🔧 FAQ — Частые вопросы и решения

### ❓ `exec: "ffmpeg": executable file not found in %PATH%`
**Причина:** `ffmpeg` не найден в переменной `PATH`.

**Решение:**
1) Установи FFmpeg (portable сборка под Windows):
    - Официальные сборки: https://www.gyan.dev/ffmpeg/builds/
    - Альтернатива: https://github.com/BtbN/FFmpeg-Builds/releases  
      Распакуй в, например, `C:\ffmpeg\bin` (внутри должен лежать `ffmpeg.exe`).

2) Добавь в `PATH`:
    - **GoLand → Run/Debug Configuration → Environment**:
      ```
      PATH=%PATH%;C:\msys64\mingw64\bin;C:\ffmpeg\bin
      ```
    - **Или PowerShell на время текущей сессии**:
      ```powershell
      $Env:PATH += ";C:\ffmpeg\bin"
      ```

3) Проверь из той же среды (GoLand/PowerShell), откуда запускаешь бота:
```powershell
   where ffmpeg
```

4. Временно диагностировать в коде:

```go
   if p, err := exec.LookPath("ffmpeg"); err != nil {
       log.Printf("ffmpeg не найден: %v\nPATH=%s", err, os.Getenv("PATH"))
   } else {
       log.Printf("ffmpeg найден: %s", p)
   }
```

> Альтернатива без PATH: жёсткий путь в коде
> `exec.Command(`C:\ffmpeg\bin\ffmpeg.exe`, ...)`

---

### ❓ `build constraints exclude all Go files in .../speechkit`

**Причина:** CGO отключён или не настроены компиляторы/флаги для VOSK.

**Решение:**

* В **GoLand → Run/Debug Configuration → Environment** добавь:

  ```
  CGO_ENABLED=1;
  CC=C:\msys64\mingw64\bin\gcc.exe;
  CXX=C:\msys64\mingw64\bin\g++.exe;
  CGO_CFLAGS=-IC:/Github/notes-ai-bot/vosk_lib;
  CGO_LDFLAGS=-LC:/Github/notes-ai-bot/vosk_lib -lvosk;
  PATH=%PATH%;C:\msys64\mingw64\bin;C:\ffmpeg\bin
  ```
* Убедись, что пути к `vosk_lib` корректны (диск/директории).
* Проверь, что используешь **64-битный** MinGW (w64) под 64-битную Go.

---

### ❓ При запуске: `libvosk.dll` не найден / приложение не стартует

**Причина:** динамическая библиотека VOSK не доступна в `PATH` или рядом с бинарником.

**Решение:**

* Положи `libvosk.dll` рядом с `notes-ai-bot.exe` **или** добавь путь к папке с DLL в `PATH`, например:

  ```
  PATH=%PATH%;C:\Github\notes-ai-bot\vosk_lib
  ```
* Перезапусти GoLand/PowerShell, чтобы обновился PATH.

---

### ❓ Код ошибки 0xc000007b (или «Неверный формат образа»)

**Причина:** несоответствие архитектур (32/64-бит).

**Решение:**

* Убедись, что все компоненты x64:
  Go (amd64), MinGW-w64 (x86\_64), `libvosk.dll` (64-бит), сборка Go (`GOARCH=amd64`).
* В GoLand: File → Settings → Go → GOROOT соответствует 64-битной Go.

---

### ❓ Голосовое сообщение конвертируется, но распознаётся плохо

**Рекомендации:**

* Мы уже принудительно конвертируем в `pcm_s16le`, `-ar 16000`, `-ac 1` — это правильно для VOSK.
* Для «шумных» записей попробуй нормализацию:

  ```powershell
  ffmpeg -i input.ogg -af "dynaudnorm" -acodec pcm_s16le -ar 16000 -ac 1 out.wav
  ```
* Переход на «большую» модель может улучшить качество (см. сайт VOSK).

---

### ❓ Где брать модель VOSK и как указать путь?

* Скачай русскую модель (например, `vosk-model-small-ru-0.22`) со страницы:
  [https://alphacephei.com/vosk/models](https://alphacephei.com/vosk/models)
* Распакуй в `models/vosk-model-small-ru-0.22`
* В `.env` укажи:

  ```env
  VOSK_MODEL_PATH=models/vosk-model-small-ru-0.22
  ```

> Важно: путь указывает **на папку модели**, а не на файл.

---

### ❓ Как сделать, чтобы переменные CGO не прописывать руками каждый раз?

**Раз и навсегда в Go:**

```powershell
go env -w CGO_ENABLED=1
go env -w CC="C:\msys64\mingw64\bin\gcc.exe"
go env -w CXX="C:\msys64\mingw64\bin\g++.exe"
go env -w CGO_CFLAGS="-IC:/Github/notes-ai-bot/vosk_lib"
go env -w CGO_LDFLAGS="-LC:/Github/notes-ai-bot/vosk_lib -lvosk"
```

**В GoLand** всё равно укажи `PATH`, потому что `go env` не меняет системный PATH:

```
PATH=%PATH%;C:\msys64\mingw64\bin;C:\ffmpeg\bin
```

---

### ❓ Я добавил PATH в системе, но GoLand всё равно «не видит» ffmpeg/gcc

**Причина:** IDE запускалась до изменения переменных среды.

**Решение:**

* Полностью перезапусти **GoLand**.
* Альтернатива: укажи `PATH` **в конфигурации Run/Debug** — это надёжнее и не зависит от системного PATH.

---

### ❓ Как изменить/ограничить количество параллельных обработчиков сообщений?

* В `.env`:

  ```env
  TELEGRAM_CONCURRENCY=8
  ```
* Выше — больше параллелизма, но выше нагрузка (VOSK и LLM — ресурсоёмкие).

---

### ❓ Как убрать «болтливость» логов VOSK в консоли?

* В `speechkit.Init` можно вызвать (при желании):

  ```c
  C.vosk_set_log_level(0);
  ```
* Или через переменную среды (если поддерживается в вашей сборке).
  Низкий уровень логов удобен на проде, но на этапе настройки лучше оставить как есть.

---

### ❓ Как быстро проверить, что всё собрано правильно?

* Из консоли **той же среды**, что и запуск бота:

  ```powershell
  where ffmpeg
  where gcc
  go env | findstr CGO
  ```
* Запусти:

  ```powershell
  go run main.go
  ```

  В логе увидишь:

    * `🔄 Загрузка модели VOSK ...`
    * `✅ Модель VOSK успешно загружена`
    * `🚀 Запуск Telegram-бота`
    * и т.д.

---

### ❓ Как установить FFmpeg на Windows (шаги)

1. Скачай **release full** (или gpl/shared) сборку под Windows x64.
2. Распакуй в `C:\ffmpeg` → внутри должна быть папка `bin\ffmpeg.exe`.
3. Добавь в `PATH` (GoLand или системно):

   ```
   C:\ffmpeg\bin
   ```
4. Проверь:

   ```powershell
   where ffmpeg
   ffmpeg -version
   ```
