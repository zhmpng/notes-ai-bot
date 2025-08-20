package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCommands обрабатывает команды типа /start, /menu и т.д.
func HandleCommands(b *Bot, msg *tgbotapi.Message) bool {
	if msg.IsCommand() {
		cmd := msg.Command()
		switch cmd {
		case "start", "menu":
			SendMainMenu(b, msg.Chat.ID)
		case "hide":
			SendMessageWithoutKeyboard(b, msg.Chat.ID, "Клавиатура скрыта.")
		default:
			SendMessage(b, msg.Chat.ID, "Неизвестная команда. Используйте /start, /menu или /hide.")
		}
		return true // Команда обработана, не продолжать
	}
	return false
}
