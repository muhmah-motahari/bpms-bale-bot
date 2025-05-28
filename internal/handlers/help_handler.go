package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HelpHandler handles the help command
type HelpHandler struct {
	channelChatID  int64
	channelMessage int
}

// NewHelpHandler creates a new HelpHandler
func NewHelpHandler() *HelpHandler {
	return &HelpHandler{
		channelChatID:  5750547246,
		channelMessage: 814,
	}
}

// HandleHelpCommand handles the help command
func (h *HelpHandler) HandleHelpCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}

	if update.Message.Text == "راهنما" {
		// Forward the message from channel
		forward := tgbotapi.NewForward(update.Message.Chat.ID, h.channelChatID, h.channelMessage)
		if _, err := bot.Send(forward); err != nil {
			log.Printf("Error forwarding help message: %v", err)
			sendMessage(update.Message.Chat.ID, "متاسفانه در ارسال راهنما مشکلی پیش آمده. لطفا دوباره تلاش کنید.")
		}
	}
}
