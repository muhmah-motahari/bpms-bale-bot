package handlers

import (
	"bbb/internal/models"
	service "bbb/internal/services"
	"fmt"
	"strconv"
	"strings"
	"time"

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
			taskExecID, err := strconv.ParseUint(strings.TrimPrefix(data, "complete_task_"), 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه وظیفه.")
				bot.Send(msg)
				return
			}

			if isFinal, _ := h.taskService.IsFinalTask(uint(taskExecID)); isFinal {
				processExecution, _ := h.processService.GetProcessExecutionByID(uint(taskExecID))
				processExecution.Status = models.ProcessExecutionStatusCompleted
				completedTime := time.Now()
				processExecution.CompletedAt = &(completedTime)
				if err := h.processService.UpdateProcessExecution(processExecution); err != nil {
					msg := tgbotapi.NewMessage(chatID, "خطا در تکمیل فرایند")
					bot.Send(msg)
					return
				}
				return
			}

			err = h.taskService.CompleteTask(uint(taskExecID), userID)
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

			msg := tgbotapi.NewMessage(chatID, taskDetails)
			bot.Send(msg)
			return
		}

		if strings.HasPrefix(data, "select_process_") {
			processID, err := strconv.ParseUint(strings.TrimPrefix(data, "select_process_"), 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه فرآیند.")
				bot.Send(msg)
				return
			}

			// Start task creation flow using task builder
			h.taskBuilderService.StartTask(userID)
			if !h.taskBuilderService.SetProcess(userID, uint(processID)) {
				msg := tgbotapi.NewMessage(chatID, "خطا در شروع ایجاد وظیفه.")
				bot.Send(msg)
				return
			}

			msg := tgbotapi.NewMessage(chatID, "لطفا عنوان وظیفه را وارد کنید:")
			bot.Send(msg)
			return
		}

		if strings.HasPrefix(data, "add_prerequisite_") {
			prerequisiteID, err := strconv.ParseUint(strings.TrimPrefix(data, "add_prerequisite_"), 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه وظیفه پیش‌نیاز.")
				bot.Send(msg)
				return
			}

			if !h.taskBuilderService.AddPrerequisite(userID, uint(prerequisiteID)) {
				msg := tgbotapi.NewMessage(chatID, "خطا در افزودن پیش‌نیاز.")
				bot.Send(msg)
				return
			}

			msg := tgbotapi.NewMessage(chatID, "آیا وظیفه پیش‌نیاز دیگری دارد؟ (بله/خیر)")
			bot.Send(msg)
			return
		}

		if strings.HasPrefix(data, "select_group_") {
			groupID, err := strconv.ParseUint(strings.TrimPrefix(data, "select_group_"), 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه گروه.")
				bot.Send(msg)
				return
			}

			if !h.taskBuilderService.SetGroup(userID, uint(groupID)) {
				msg := tgbotapi.NewMessage(chatID, "خطا در تنظیم گروه.")
				bot.Send(msg)
				return
			}

			// Ask if this is a final task
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("بله", "set_final_true"),
					tgbotapi.NewInlineKeyboardButtonData("خیر", "set_final_false"),
				),
			)

			msg := tgbotapi.NewMessage(chatID, "آیا این وظیفه پایانی است؟")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			return
		}

		if strings.HasPrefix(data, "set_final_") {
			isFinal := strings.HasSuffix(data, "true")
			if !h.taskBuilderService.SetIsFinal(userID, isFinal) {
				msg := tgbotapi.NewMessage(chatID, "خطا در تنظیم وضعیت نهایی وظیفه.")
				bot.Send(msg)
				return
			}

			// Complete task creation
			task, prerequisites, success := h.taskBuilderService.CompleteTask(userID)
			if !success {
				msg := tgbotapi.NewMessage(chatID, "خطا در تکمیل ایجاد وظیفه.")
				bot.Send(msg)
				return
			}

			// Save task
			if err := h.taskService.CreateTask(task); err != nil {
				msg := tgbotapi.NewMessage(chatID, "خطا در ذخیره وظیفه.")
				bot.Send(msg)
				return
			}

			// Add prerequisites
			for _, prerequisiteID := range prerequisites {
				if err := h.taskService.AddPrerequisite(task.ID, prerequisiteID); err != nil {
					msg := tgbotapi.NewMessage(chatID, "خطا در افزودن پیش‌نیازها.")
					bot.Send(msg)
					return
				}
			}

			msg := tgbotapi.NewMessage(chatID, "وظیفه با موفقیت ایجاد شد.")
			bot.Send(msg)
			return
		}
	}

	if update.Message != nil {
		// Check if user is in task creation flow
		if builder, exists := h.taskBuilderService.GetBuilder(update.Message.From.ID); exists {
			switch builder.CurrentStep {
			case "title":
				if !h.taskBuilderService.SetTitle(update.Message.From.ID, update.Message.Text) {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در تنظیم عنوان وظیفه.")
					bot.Send(msg)
					return
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "لطفا توضیحات وظیفه را وارد کنید:")
				bot.Send(msg)
				return

			case "description":
				if !h.taskBuilderService.SetDescription(update.Message.From.ID, update.Message.Text) {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در تنظیم توضیحات وظیفه.")
					bot.Send(msg)
					return
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "آیا وظیفه پیش‌نیاز دارد؟ (بله/خیر)")
				bot.Send(msg)
				return

			case "prerequisites":
				if update.Message.Text == "بله" {
					// Get available tasks for prerequisites
					tasks, err := h.taskService.GetTasksByProcessID(builder.ProcessID)
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در دریافت وظایف.")
						bot.Send(msg)
						return
					}

					var keyboard [][]tgbotapi.InlineKeyboardButton
					for _, task := range tasks {
						row := []tgbotapi.InlineKeyboardButton{
							tgbotapi.NewInlineKeyboardButtonData(task.Title, fmt.Sprintf("add_prerequisite_%d", task.ID)),
						}
						keyboard = append(keyboard, row)
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "لطفا وظیفه پیش‌نیاز را انتخاب کنید:")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
					bot.Send(msg)
					return
				} else if update.Message.Text == "خیر" {
					if !h.taskBuilderService.SetHasMorePrerequisites(update.Message.From.ID, false) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در تنظیم پیش‌نیازها.")
						bot.Send(msg)
						return
					}

					// Get available groups
					groups, err := h.groupService.GetAllGroups()
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در دریافت گروه‌ها.")
						bot.Send(msg)
						return
					}

					var keyboard [][]tgbotapi.InlineKeyboardButton
					for _, group := range groups {
						row := []tgbotapi.InlineKeyboardButton{
							tgbotapi.NewInlineKeyboardButtonData(group.Name, fmt.Sprintf("select_group_%d", group.ID)),
						}
						keyboard = append(keyboard, row)
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "لطفا گروه مورد نظر را انتخاب کنید:")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
					bot.Send(msg)
					return
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "لطفا 'بله' یا 'خیر' را وارد کنید.")
					bot.Send(msg)
					return
				}
			}
		}

		if update.Message.Text == "وظایف من" {
			taskExecutions, err := h.taskService.GetUserTasks(update.Message.From.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "خطا در دریافت وظایف شما.")
				bot.Send(msg)
				return
			}

			if len(taskExecutions) == 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "شما هیچ وظیفه‌ای ندارید.")
				bot.Send(msg)
				return
			}

			var keyboard [][]tgbotapi.InlineKeyboardButton
			for _, taskExecution := range taskExecutions {
				row := []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData(taskExecution.Task.Title, fmt.Sprintf("view_task_%d", taskExecution.ID)),
				}
				if taskExecution.Status == models.TaskStatusAssigned {
					row = append(row, tgbotapi.NewInlineKeyboardButtonData("تکمیل", fmt.Sprintf("complete_task_%d", taskExecution.ID)))
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
