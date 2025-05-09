package handlers

import (
	"bbb/internal/models"
	service "bbb/internal/services"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ProcessHandler struct {
	processService        service.ProcessService
	processBuilderService *service.ProcessBuilderService
	processExecService    service.ProcessExecutionService
	taskService           service.TaskService
}

func NewProcessHandler(
	processService service.ProcessService,
	processBuilderService *service.ProcessBuilderService,
	processExecService service.ProcessExecutionService,
	taskService service.TaskService,
) *ProcessHandler {
	return &ProcessHandler{
		processService:        processService,
		processBuilderService: processBuilderService,
		processExecService:    processExecService,
		taskService:           taskService,
	}
}

func (h *ProcessHandler) HandleProcessCreation(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if update.Message.Text == "فرایند جدید" {
		h.processBuilderService.StartProcess(userID)
		msg := tgbotapi.NewMessage(chatID, "لطفا نام فرایند را وارد کنید")
		bot.Send(msg)
		return
	}

	builder, exists := h.processBuilderService.GetBuilder(userID)
	if !exists {
		return
	}

	switch builder.CurrentStep {
	case "name":
		if h.processBuilderService.SetProcessName(userID, update.Message.Text) {
			msg := tgbotapi.NewMessage(chatID, "توضیحات فرایند را وارد کنید")
			bot.Send(msg)
		}
	case "description":
		if h.processBuilderService.SetProcessDescription(userID, update.Message.Text) {
			process, success := h.processBuilderService.CompleteProcess(userID)
			if success {
				err := h.processService.CreateProcess(process)
				if err != nil {
					msg := tgbotapi.NewMessage(chatID, "خطا در ایجاد فرآیند. لطفا دوباره تلاش کنید.")
					bot.Send(msg)
					return
				}
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("فرایند با موفقیت ایجاد شد!\nنام: %s\nتوضیحات: %s",
					process.Name, process.Description))
				bot.Send(msg)
			}
		}
	}
}

func (h *ProcessHandler) HandleProcessExecution(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if update.Message.Text == "شروع فرایند" {
		if processes, err := h.processService.GetProcessesByUserID(userID); err == nil {
			var keyboardRows [][]tgbotapi.InlineKeyboardButton
			for _, process := range processes {
				row := []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData(
						process.Name,
						fmt.Sprintf("start_process_%d", process.ID),
					),
				}
				keyboardRows = append(keyboardRows, row)
			}
			keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
			msg := tgbotapi.NewMessage(chatID, "لطفا فرایند را انتخاب کنید:")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}
		return
	}

}

func (h *ProcessHandler) HandleProcessCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if update.Message.Text == "فرایند ها" {
		processes, err := h.processService.GetProcessesByUserID(userID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در دریافت فرایندها. لطفا دوباره تلاش کنید.")
			bot.Send(msg)
			return
		}

		if len(processes) == 0 {
			msg := tgbotapi.NewMessage(chatID, "شما هیچ فرایندی ندارید.")
			bot.Send(msg)
			return
		}

		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, process := range processes {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(process.Name, fmt.Sprintf("view_process_%d", process.ID)),
			}
			keyboard = append(keyboard, row)
		}

		msg := tgbotapi.NewMessage(chatID, "فرایندهای شما:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
		bot.Send(msg)
		return
	}
}

func (h *ProcessHandler) HandleProcessCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	if strings.HasPrefix(data, "start_process_") {
		processID, _ := strconv.ParseUint(strings.TrimPrefix(data, "start_process_"), 10, 32)

		// Start a new process execution
		execution, err := h.processService.StartProcessExecution(uint(processID))
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در شروع فرایند. لطفا دوباره تلاش کنید.")
			bot.Send(msg)
			return
		}

		// Get the first task
		tasks, err := h.taskService.GetTasksByProcessID(uint(processID))
		var firstTasks []models.Task
		for _, task := range tasks {
			preTasks, _ := h.taskService.GetTaskPrerequisites(task.ID)
			if len(preTasks) == 0 {
				firstTasks = append(firstTasks, task)
			}
		}
		if err != nil || len(firstTasks) == 0 {
			msg := tgbotapi.NewMessage(chatID, "خطا در دریافت وظایف فرایند.")
			bot.Send(msg)
			return
		}

		// Add the first task to pending tasks
		for _, task := range firstTasks {
			if err := h.processService.AddPendingTask(execution.ID, task.ID); err != nil {
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("خطا در اضافه کردن وظیفه %v.", task.Title))
				bot.Send(msg)
				return
			}
			if err := h.taskService.StartTaskExecution(task.ID); err != nil {
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("خطا در شروع وظیفه %v.", task.Title))
				bot.Send(msg)
				return
			}
		}

		// Create message with all initial tasks
		var tasksInfo strings.Builder
		tasksInfo.WriteString(fmt.Sprintf("فرایند با موفقیت شروع شد!\nشناسه اجرا: %d\n\nوظایف اولیه:\n", execution.ID))
		for i, task := range firstTasks {
			tasksInfo.WriteString(fmt.Sprintf("%d. %s\n   توضیحات: %s\n", i+1, task.Title, task.Description))
		}

		msg := tgbotapi.NewMessage(chatID, tasksInfo.String())
		bot.Send(msg)
	}

	if strings.HasPrefix(data, "view_process_") {
		processID, err := strconv.ParseUint(strings.TrimPrefix(data, "view_process_"), 10, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در پردازش شناسه فرایند.")
			bot.Send(msg)
			return
		}

		tasks, err := h.taskService.GetTasksByProcessID(uint(processID))
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در دریافت وظایف فرایند.")
			bot.Send(msg)
			return
		}

		if len(tasks) == 0 {
			msg := tgbotapi.NewMessage(chatID, "این فرایند هیچ وظیفه‌ای ندارد.")
			bot.Send(msg)
			return
		}

		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, task := range tasks {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(task.Title, fmt.Sprintf("view_task_%d", task.ID)),
			}
			keyboard = append(keyboard, row)
		}

		msg := tgbotapi.NewMessage(chatID, "وظایف این فرایند:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
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
}
