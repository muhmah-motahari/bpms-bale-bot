package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartHandler handles the /start command
type StartHandler struct {
	keyboard *tgbotapi.ReplyKeyboardMarkup
}

// NewStartHandler creates a new StartHandler
func NewStartHandler(keyboard *tgbotapi.ReplyKeyboardMarkup) *StartHandler {
	return &StartHandler{
		keyboard: keyboard,
	}
}

// HandleStartCommand handles the /start command
func (h *StartHandler) HandleStartCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}

	if update.Message.Text == "/start" {
		welcomeMessage := `به ربات مدیریت فرآیندهای کسب و کار خوش آمدید! 👋

برای شروع کار با ربات، می‌توانید از دستورات زیر استفاده کنید:

📋 دستورات اصلی:
• فرایند جدید - ایجاد یک فرآیند جدید
• فرایند ها - مشاهده لیست فرآیندهای شما
• شروع فرایند - شروع اجرای یک فرآیند
• وظیفه جدید - ایجاد یک وظیفه جدید
• تیم جدید - ایجاد یک تیم جدید
• تیم ها - مشاهده لیست تیم‌ها
• راهنما - مشاهده راهنمای کامل ربات

برای اطلاعات بیشتر می‌توانید از دستور راهنما استفاده کنید.`

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
		msg.ReplyMarkup = h.keyboard
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending welcome message: %v", err)
			sendMessage(update.Message.Chat.ID, "متاسفانه در ارسال پیام خوش‌آمدگویی مشکلی پیش آمده. لطفا دوباره تلاش کنید.")
		}
	}
}
