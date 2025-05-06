package handlers

import (
	"bbb/internal/models"
	service "bbb/internal/services"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskHandler struct {
	taskService        service.TaskService
	taskBuilderService *service.TaskBuilderService
	processService     service.ProcessService
	groupService       service.GroupService
}

func NewTaskHandler(
	taskService service.TaskService,
	taskBuilderService *service.TaskBuilderService,
	processService service.ProcessService,
	groupService service.GroupService,
) *TaskHandler {
	return &TaskHandler{
		taskService:        taskService,
		taskBuilderService: taskBuilderService,
		processService:     processService,
		groupService:       groupService,
	}
}

func (h *TaskHandler) HandleTaskCreation(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if update.Message.Text == "وظیفه جدید" {
		processes, err := h.processService.GetProcessesByUserID(userID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در دریافت فرآیندها. لطفا دوباره تلاش کنید.")
			bot.Send(msg)
			return
		}

		if len(processes) == 0 {
			msg := tgbotapi.NewMessage(chatID, "شما هیچ فرآیندی ندارید. ابتدا یک فرآیند ایجاد کنید.")
			bot.Send(msg)
			return
		}

		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, process := range processes {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(process.Name, fmt.Sprintf("select_process_%d", process.ID)),
			}
			keyboard = append(keyboard, row)
		}

		msg := tgbotapi.NewMessage(chatID, "لطفا فرآیند مورد نظر را انتخاب کنید:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
		bot.Send(msg)
		return
	}
}

func (h *TaskHandler) HandleTaskCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		data := update.CallbackQuery.Data
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Chat.ID

		if strings.HasPrefix(data, "take_task_") {
			taskID, err := strconv.ParseUint(strings.TrimPrefix(data, "take_task_"), 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه وظیفه.")
				bot.Send(msg)
				return
			}

			err = h.taskService.AssignTask(uint(taskID), userID)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در به عهده گرفتن وظیفه: "+err.Error())
				bot.Send(msg)
				return
			}

			msg := tgbotapi.NewMessage(chatID, "وظیفه با موفقیت به شما اختصاص داده شد.")
			bot.Send(msg)
			return
		}

		if strings.HasPrefix(data, "complete_task_") {
			taskID, err := strconv.ParseUint(strings.TrimPrefix(data, "complete_task_"), 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه وظیفه.")
				bot.Send(msg)
				return
			}

			err = h.taskService.CompleteTask(uint(taskID), userID)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در تکمیل وظیفه: "+err.Error())
				bot.Send(msg)
				return
			}

			msg := tgbotapi.NewMessage(chatID, "وظیفه با موفقیت تکمیل شد.")
			bot.Send(msg)
			return
		}

		if strings.HasPrefix(data, "view_task_") {
			taskID, err := strconv.ParseUint(strings.TrimPrefix(data, "view_task_"), 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه وظیفه.")
				bot.Send(msg)
				return
			}

			task, err := h.taskService.GetTaskByID(uint(taskID))
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در دریافت اطلاعات وظیفه.")
				bot.Send(msg)
				return
			}

			taskDetails := fmt.Sprintf("عنوان: %s\nتوضیحات: %s",
				task.Title,
				task.Description)

			if task.AssignedAt != nil {
				taskDetails += fmt.Sprintf("\nزمان به عهده گرفتن: %s", task.AssignedAt.Format("2006-01-02 15:04:05"))
			}

			if task.CompletedAt != nil {
				taskDetails += fmt.Sprintf("\nزمان تکمیل: %s", task.CompletedAt.Format("2006-01-02 15:04:05"))
			}

			msg := tgbotapi.NewMessage(chatID, taskDetails)
			bot.Send(msg)
			return
		}
	}

	if update.Message != nil {
		if update.Message.Text == "وظایف من" {
			tasks, err := h.taskService.GetUserTasks(update.Message.From.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در دریافت وظایف شما.")
				bot.Send(msg)
				return
			}

			if len(tasks) == 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "شما هیچ وظیفه‌ای ندارید.")
				bot.Send(msg)
				return
			}

			var keyboard [][]tgbotapi.InlineKeyboardButton
			for _, task := range tasks {
				row := []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData(task.Title, fmt.Sprintf("view_task_%d", task.ID)),
				}
				if task.Status == models.TaskStatusAssigned {
					row = append(row, tgbotapi.NewInlineKeyboardButtonData("تکمیل", fmt.Sprintf("complete_task_%d", task.ID)))
				}
				keyboard = append(keyboard, row)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "وظایف شما:")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
			bot.Send(msg)
			return
		}
	}
}

func (h *TaskHandler) NotifyGroupMembers(bot *tgbotapi.BotAPI, task *models.Task) error {
	if task.GroupID == nil {
		return fmt.Errorf("task has no group assigned")
	}

	group, err := h.groupService.GetGroupByID(*task.GroupID)
	if err != nil {
		return err
	}

	members, err := h.groupService.GetGroupMembers(group.ID)
	if err != nil {
		return err
	}

	taskMsg := fmt.Sprintf("وظیفه با اطلاعات زیر فعال شده است، اگر تمایل دارید که انجام دهید اعلام کنید.\n\nعنوان: %s\nتوضیحات: %s",
		task.Title, task.Description)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("به عهده گرفتن وظیفه", fmt.Sprintf("take_task_%d", task.ID)),
		),
	)

	for _, member := range members {
		msg := tgbotapi.NewMessage(member.ID, taskMsg)
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			return fmt.Errorf("error sending message to user %d: %v", member.ID, err)
		}
	}

	return nil
}
