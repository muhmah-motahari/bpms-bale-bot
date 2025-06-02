package handlers

import (
	service "bbb/internal/services"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TeamHandler handles commands and callbacks related to teams.
type TeamHandler struct {
	teamService        service.TeamService
	userService        service.UserService
	teamBuilderService *service.TeamBuilderService
}

// NewTeamHandler creates a new TeamHandler.
func NewTeamHandler(
	teamService service.TeamService,
	userService service.UserService,
	teamBuilderService *service.TeamBuilderService,
) *TeamHandler {
	return &TeamHandler{
		teamService:        teamService,
		userService:        userService,
		teamBuilderService: teamBuilderService,
	}
}

// HandleTeamCommands handles message-based commands for team management.
func (h *TeamHandler) HandleTeamCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	switch {
	case update.Message.Text == "تیم جدید":
		h.teamBuilderService.StartTeam(userID)
		sendMessage(chatID, "لطفا نام تیم را وارد کنید")
		return

	case update.Message.Text == "عضویت در تیم":
		h.teamBuilderService.StartJoinTeam(userID)
		sendMessage(chatID, "لطفا کد پیوستن به تیم را بفرستید")
		return

	case update.Message.Text == "لیست تیم ها":
		teams, err := h.teamService.GetTeamsByOwnerID(userID)
		if err != nil {
			sendMessage(chatID, "خطا در دریافت لیست تیم‌ها. لطفا دوباره تلاش کنید.")
			return
		}

		if len(teams) == 0 {
			sendMessage(chatID, "هیچ تیمی وجود ندارد.")
			return
		}

		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, team := range teams {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(team.Name, fmt.Sprintf("view_team_%d", team.ID)),
			}
			keyboard = append(keyboard, row)
		}

		msg := tgbotapi.NewMessage(chatID, "لیست تیم‌ها:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
		if _, errSend := bot.Send(msg); errSend != nil { // Keep bot.Send for specific inline keyboard
			log.Printf("Error sending team list: %v", errSend)
		}
		return

	case strings.HasPrefix(update.Message.Text, "پیوستن به تیم:"):
		joinKey := strings.TrimPrefix(update.Message.Text, "پیوستن به تیم:")
		joinKey = strings.TrimSpace(joinKey)

		if err := h.teamService.JoinTeam(userID, joinKey); err != nil {
			sendMessage(chatID, "خطا در پیوستن به تیم. لطفا کلید پیوست را بررسی کنید یا مطمئن شوید قبلا عضو نشده‌اید: "+err.Error())
			return
		}
		sendMessage(chatID, "با موفقیت به تیم پیوستید!")
		return
	}

	// Check if user is in the middle of creating a team
	builder, exists := h.teamBuilderService.GetBuilder(userID)
	if exists {
		if builder.IsJoining {
			// User is in the process of joining a team
			if !h.teamBuilderService.SetName(userID, update.Message.Text) {
				sendMessage(chatID, "خطا در پردازش کد پیوستن.")
				return
			}

			joinKey, success := h.teamBuilderService.CompleteJoinTeam(userID)
			if !success {
				sendMessage(chatID, "خطا در تکمیل فرآیند پیوستن به تیم.")
				return
			}

			if err := h.teamService.JoinTeam(userID, joinKey); err != nil {
				sendMessage(chatID, "خطا در پیوستن به تیم. لطفا کد پیوست را بررسی کنید یا مطمئن شوید قبلا عضو نشده‌اید: "+err.Error())
				return
			}
			sendMessage(chatID, "با موفقیت به تیم پیوستید!")
			return
		} else {
			// User is in the process of creating a team
			if !h.teamBuilderService.SetName(userID, update.Message.Text) {
				sendMessage(chatID, "خطا در تنظیم نام تیم.")
				return
			}

			team, success := h.teamBuilderService.CompleteTeam(userID)
			if !success {
				sendMessage(chatID, "خطا در تکمیل ایجاد تیم.")
				return
			}

			if err := h.teamService.CreateTeam(team); err != nil {
				sendMessage(chatID, "خطا در ایجاد تیم. لطفا دوباره تلاش کنید.")
				return
			}

			sendMessage(chatID, fmt.Sprintf("تیم با موفقیت ساخته شد!\nکلید عضویت به تیم \"%s\"\nافرادی که می‌خواهید در این تیم عضو شوند این پیام را برایشان ارسال کنید:\n\nجهت عضویت در تیم \"%s\" ، به بازو @%s پیام \"عضویت در تیم\" را ارسال کنید و کلید زیر را وارد کنید:\n %s", team.Name, team.Name, bot.Self.UserName, team.JoinKey))
			return
		}
	}
}

// HandleTeamCallback handles callback queries for team actions.
func (h *TeamHandler) HandleTeamCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackMsg := ""

	if strings.HasPrefix(data, "view_team_") {
		teamIDStr := strings.TrimPrefix(data, "view_team_")
		teamID, err := strconv.ParseUint(teamIDStr, 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه تیم.")
			callbackMsg = "شناسه نامعتبر"
		} else {
			team, err := h.teamService.GetTeamByID(uint(teamID))
			if err != nil {
				sendMessage(chatID, "خطا در دریافت اطلاعات تیم.")
				callbackMsg = "خطا در تیم"
			} else {
				members, err := h.teamService.GetTeamMembers(team.ID)
				if err != nil {
					sendMessage(chatID, "خطا در دریافت اعضای تیم.")
					callbackMsg = "خطا در اعضا"
				} else {
					var membersList strings.Builder
					membersList.WriteString(fmt.Sprintf("اطلاعات تیم %s:\n\n", team.Name))
					membersList.WriteString(fmt.Sprintf("کلید پیوستن به تیم: %s\n\n", team.JoinKey))
					membersList.WriteString("اعضای تیم:\n")
					if len(members) == 0 {
						membersList.WriteString("(این تیم فعلا عضوی ندارد)\n")
					} else {
						for i, member := range members {
							membersList.WriteString(fmt.Sprintf("%d. %s %s (@%s)\n",
								i+1,
								member.FirstName,
								member.LastName,
								member.Username))
						}
					}
					sendMessage(chatID, membersList.String())
					callbackMsg = "اطلاعات تیم نمایش داده شد"
				}
			}
		}
	}
	// Answer the callback query
	if callbackMsg != "" {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, callbackMsg)
		if _, err := bot.Request(callback); err != nil {
			log.Printf("Error answering callback query: %v", err)
		}
	}
}
