package handlers

import (
	"bbb/internal/models"
	service "bbb/internal/services"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type GroupHandler struct {
	groupService service.GroupService
	userService  service.UserService
}

func NewGroupHandler(
	groupService service.GroupService,
	userService service.UserService,
) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		userService:  userService,
	}
}

func (h *GroupHandler) HandleGroupCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if update.Message.Text == "گروه جدید" {
		msg := tgbotapi.NewMessage(chatID, " لطفا نام گروه را وارد کنید.\nتوجه کنید که حتما به صورت زیر پیام خود را ارسال کنید:\n\n*نام گروه: تکنسین کنترل توربین*")
		bot.Send(msg)
		return
	}

	if update.Message.Text == "لیست گروه ها" {
		groups, err := h.groupService.GetAllGroups()
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در دریافت لیست گروه‌ها. لطفا دوباره تلاش کنید.")
			bot.Send(msg)
			return
		}

		if len(groups) == 0 {
			msg := tgbotapi.NewMessage(chatID, "هیچ گروهی وجود ندارد.")
			bot.Send(msg)
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
		bot.Send(msg)
		return
	}

	if strings.HasPrefix(update.Message.Text, "نام گروه:") {
		groupName := strings.TrimPrefix(update.Message.Text, "نام گروه:")
		groupName = strings.TrimSpace(groupName)

		group := models.Group{
			Name: groupName,
		}

		if err := h.groupService.CreateGroup(&group); err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در ایجاد گروه. لطفا دوباره تلاش کنید.")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(" گروه با موفقیت ساخته شد!\n کلید عضویت به گروه %s\n افرادی که میخواهید در این گروه عضو شوند این پیام را برایشان ارسال کنید:\n\n جهت عضویت در گروه %s ، به بازو @bpmss پیام زیر را ارسال کنید:\n پیوستن به گروه: %s\n\n گروه با موفقیت ایجاد شد!\nکلید پیوستن به گروه: %s", group.Name, group.Name, group.JoinKey, group.JoinKey))
		bot.Send(msg)
		return
	}

	if strings.HasPrefix(update.Message.Text, "پیوستن به گروه:") {
		joinKey := strings.TrimPrefix(update.Message.Text, "پیوستن به گروه:")
		joinKey = strings.TrimSpace(joinKey)

		user, err := h.userService.GetUserByID(userID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در پیدا کردن کاربر. لطفا دوباره تلاش کنید.")
			bot.Send(msg)
			return
		}

		if err := h.groupService.JoinGroup(user.ID, joinKey); err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در پیوستن به گروه. لطفا کلید پیوست را بررسی کنید.")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "با موفقیت به گروه پیوستید!")
		bot.Send(msg)
		return
	}
}

func (h *GroupHandler) HandleGroupCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	if strings.HasPrefix(data, "view_group_") {
		groupID, err := strconv.ParseUint(strings.TrimPrefix(data, "view_group_"), 10, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه گروه.")
			bot.Send(msg)
			return
		}

		group, err := h.groupService.GetGroupByID(uint(groupID))
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در دریافت اطلاعات گروه.")
			bot.Send(msg)
			return
		}

		members, err := h.groupService.GetGroupMembers(group.ID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در دریافت اعضای گروه.")
			bot.Send(msg)
			return
		}

		var membersList strings.Builder
		membersList.WriteString(fmt.Sprintf("اطلاعات گروه %s:\n\n", group.Name))
		membersList.WriteString(fmt.Sprintf("کلید پیوستن به گروه: %s\n\n", group.JoinKey))
		membersList.WriteString("اعضای گروه:\n")
		for i, member := range members {
			membersList.WriteString(fmt.Sprintf("%d. %s %s (@%s)\n",
				i+1,
				member.FirstName,
				member.LastName,
				member.Username))
		}

		msg := tgbotapi.NewMessage(chatID, membersList.String())
		bot.Send(msg)
		return
	}
}
