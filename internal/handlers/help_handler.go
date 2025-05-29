package handlers

import (
	"bbb/configs"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HelpHandler handles the help command
type HelpHandler struct {
	env configs.Env
}

// NewHelpHandler creates a new HelpHandler
func NewHelpHandler(env configs.Env) *HelpHandler {
	return &HelpHandler{
		env: env,
	}
}

// HandleHelpCommand handles the help command
func (h *HelpHandler) HandleHelpCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}

	if update.Message.Text == "راهنما" {
		// Forward the message from channel
		forward := tgbotapi.NewForward(update.Message.Chat.ID, h.env.HelpMessageChatID, h.env.HelpMessageID)
		if _, err := bot.Send(forward); err != nil {
			log.Printf("Error forwarding help message: %v", err)
			sendMessage(update.Message.Chat.ID, "متاسفانه در ارسال راهنما مشکلی پیش آمده. لطفا دوباره تلاش کنید.")
		}
	}
}
