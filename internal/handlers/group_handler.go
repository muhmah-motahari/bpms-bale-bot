package handlers

import (
	service "bbb/internal/services"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GroupHandler handles commands and callbacks related to groups.
type GroupHandler struct {
	groupService        service.GroupService
	userService         service.UserService
	groupBuilderService *service.GroupBuilderService
}

// NewGroupHandler creates a new GroupHandler.
func NewGroupHandler(
	groupService service.GroupService,
	userService service.UserService,
	groupBuilderService *service.GroupBuilderService,
) *GroupHandler {
	return &GroupHandler{
		groupService:        groupService,
		userService:         userService,
		groupBuilderService: groupBuilderService,
	}
}

// HandleGroupCommands handles message-based commands for group management.
func (h *GroupHandler) HandleGroupCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	switch {
	case update.Message.Text == "گروه جدید":
		h.groupBuilderService.StartGroup(userID)
		sendMessage(chatID, "لطفا نام گروه را وارد کنید")
		return

	case update.Message.Text == "عضویت در گروه":
		h.groupBuilderService.StartJoinGroup(userID)
		sendMessage(chatID, "لطفا کد پیوستن به گروه را بفرستید")
		return

	case update.Message.Text == "لیست گروه ها":
		groups, err := h.groupService.GetGroupsByOwnerID(userID)
		if err != nil {
			sendMessage(chatID, "خطا در دریافت لیست گروه‌ها. لطفا دوباره تلاش کنید.")
			return
		}

		if len(groups) == 0 {
			sendMessage(chatID, "هیچ گروهی وجود ندارد.")
			return
		}

		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, group := range groups {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(group.Name, fmt.Sprintf("view_group_%d", group.ID)),
			}
			keyboard = append(keyboard, row)
		}

		msg := tgbotapi.NewMessage(chatID, "لیست گروه‌ها:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
		if _, errSend := bot.Send(msg); errSend != nil { // Keep bot.Send for specific inline keyboard
			log.Printf("Error sending group list: %v", errSend)
		}
		return

	case strings.HasPrefix(update.Message.Text, "پیوستن به گروه:"):
		joinKey := strings.TrimPrefix(update.Message.Text, "پیوستن به گروه:")
		joinKey = strings.TrimSpace(joinKey)

		if err := h.groupService.JoinGroup(userID, joinKey); err != nil {
			sendMessage(chatID, "خطا در پیوستن به گروه. لطفا کلید پیوست را بررسی کنید یا مطمئن شوید قبلا عضو نشده‌اید: "+err.Error())
			return
		}
		sendMessage(chatID, "با موفقیت به گروه پیوستید!")
		return
	}

	// Check if user is in the middle of creating a group
	builder, exists := h.groupBuilderService.GetBuilder(userID)
	if exists {
		if builder.IsJoining {
			// User is in the process of joining a group
			if !h.groupBuilderService.SetName(userID, update.Message.Text) {
				sendMessage(chatID, "خطا در پردازش کد پیوستن.")
				return
			}

			joinKey, success := h.groupBuilderService.CompleteJoinGroup(userID)
			if !success {
				sendMessage(chatID, "خطا در تکمیل فرآیند پیوستن به گروه.")
				return
			}

			if err := h.groupService.JoinGroup(userID, joinKey); err != nil {
				sendMessage(chatID, "خطا در پیوستن به گروه. لطفا کد پیوست را بررسی کنید یا مطمئن شوید قبلا عضو نشده‌اید: "+err.Error())
				return
			}
			sendMessage(chatID, "با موفقیت به گروه پیوستید!")
			return
		} else {
			// User is in the process of creating a group
			if !h.groupBuilderService.SetName(userID, update.Message.Text) {
				sendMessage(chatID, "خطا در تنظیم نام گروه.")
				return
			}

			group, success := h.groupBuilderService.CompleteGroup(userID)
			if !success {
				sendMessage(chatID, "خطا در تکمیل ایجاد گروه.")
				return
			}

			if err := h.groupService.CreateGroup(group); err != nil {
				sendMessage(chatID, "خطا در ایجاد گروه. لطفا دوباره تلاش کنید.")
				return
			}

			sendMessage(chatID, fmt.Sprintf("گروه با موفقیت ساخته شد!\nکلید عضویت به گروه %s\nافرادی که می‌خواهید در این گروه عضو شوند این پیام را برایشان ارسال کنید:\n\nجهت عضویت در گروه %s ، به بازو @bpmss پیام زیر را ارسال کنید:\nپیوستن به گروه: %s", group.Name, group.Name, group.JoinKey))
			return
		}
	}
}

// HandleGroupCallback handles callback queries for group actions.
func (h *GroupHandler) HandleGroupCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackMsg := ""

	if strings.HasPrefix(data, "view_group_") {
		groupIDStr := strings.TrimPrefix(data, "view_group_")
		groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه گروه.")
			callbackMsg = "شناسه نامعتبر"
		} else {
			group, err := h.groupService.GetGroupByID(uint(groupID))
			if err != nil {
				sendMessage(chatID, "خطا در دریافت اطلاعات گروه.")
				callbackMsg = "خطا در گروه"
			} else {
				members, err := h.groupService.GetGroupMembers(group.ID)
				if err != nil {
					sendMessage(chatID, "خطا در دریافت اعضای گروه.")
					callbackMsg = "خطا در اعضا"
				} else {
					var membersList strings.Builder
					membersList.WriteString(fmt.Sprintf("اطلاعات گروه %s:\n\n", group.Name))
					membersList.WriteString(fmt.Sprintf("کلید پیوستن به گروه: %s\n\n", group.JoinKey))
					membersList.WriteString("اعضای گروه:\n")
					if len(members) == 0 {
						membersList.WriteString("(این گروه فعلا عضوی ندارد)\n")
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
					callbackMsg = "اطلاعات گروه نمایش داده شد"
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
