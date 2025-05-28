package handlers

import (
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HelpHandler handles help-related commands and messages
type HelpHandler struct{}

// NewHelpHandler creates a new HelpHandler
func NewHelpHandler() *HelpHandler {
	return &HelpHandler{}
}

// HandleHelpCommand handles the help command
func (h *HelpHandler) HandleHelpCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}

	if update.Message.Text == "راهنما" {
		// Then send the help image
		imagePath := filepath.Join("assets", "images", "help.png")
		photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FilePath(imagePath))
		photo.Caption = "نمودار جریان کار در سیستم BPMS"
		if _, err := bot.Send(photo); err != nil {
			// If image sending fails, send a message to the user
			sendMessage(update.Message.Chat.ID, "متأسفانه در ارسال تصویر راهنما مشکلی پیش آمد.")
		}

		helpText := `راهنمای استفاده از ربات BPMS:

1️⃣ *مدیریت گروه‌ها*
• گروه جدید: ایجاد یک گروه جدید برای تعریف مسئولیت‌ها
• لیست گروه ها: مشاهده گروه‌های موجود
• پس از ایجاد گروه، ربات یک کلید عضویت به شما می‌دهد که می‌توانید آن را به افراد مورد نظر بدهید تا در گروه عضو شوند

2️⃣ *مدیریت فرایندها*
• فرایند جدید: ایجاد یک فرایند جدید
• فرایند ها: مشاهده لیست فرایندها
• شروع فرایند: اجرای یک فرایند موجود (هر فرایند می‌تواند چندین بار به صورت همزمان اجرا شود)

3️⃣ *مدیریت وظایف*
• وظیفه جدید: ایجاد وظیفه جدید در یک فرایند
• وظایف من: مشاهده وظایف محول شده به شما

*نحوه کار سیستم:*
1. ابتدا گروه‌های مورد نیاز را ایجاد کنید و افراد را به آن‌ها اضافه کنید
2. یک فرایند جدید بسازید
3. برای فرایند، وظایف مورد نیاز را تعریف کنید:
   - برای هر وظیفه، گروه مسئول آن را مشخص کنید
   - می‌توانید پیش‌نیازهای هر وظیفه را تعریف کنید
   - وظایف می‌توانند به صورت متوالی یا موازی اجرا شوند
4. فرایند را اجرا کنید:
   - وظایف اولیه (بدون پیش‌نیاز) فعال می‌شوند
   - به اعضای گروه‌های مربوطه پیام ارسال می‌شود
   - اولین فردی که وظیفه را به عهده بگیرد، مسئول انجام آن می‌شود
   - پس از تکمیل وظیفه، وظایف وابسته به آن فعال می‌شوند
   - این روند تا تکمیل تمام وظایف فرایند ادامه می‌یابد`

		sendMessage(update.Message.Chat.ID, helpText)

	}
}
