package handlers

import (
	// Keep for other handlers if used, or remove if not used in this file
	"bbb/internal/models"
	service "bbb/internal/services"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ProcessHandler handles incoming commands and messages related to processes.
type ProcessHandler struct {
	processService        service.ProcessService
	processBuilderService *service.ProcessBuilderService
	processExecService    service.ProcessExecutionService // Retained as it's a distinct service
	taskService           service.TaskService
}

const (
	processAlreadyExists = "Process with this name already exists. Please choose a different name."
)

// NewProcessHandler creates a new ProcessHandler.
func NewProcessHandler(
	processService service.ProcessService,
	processBuilderService *service.ProcessBuilderService,
	processExecService service.ProcessExecutionService, // Retained
	taskService service.TaskService,
) *ProcessHandler {
	return &ProcessHandler{
		processService:        processService,
		processBuilderService: processBuilderService,
		processExecService:    processExecService, // Retained
		taskService:           taskService,
	}
}

// HandleProcessCreation handles messages related to creating a new process.
func (h *ProcessHandler) HandleProcessCreation(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	// Check if the user is starting a new process creation
	if update.Message.Text == "فرایند جدید" { // "New Process" in Farsi
		h.processBuilderService.StartProcess(userID)
		sendMessage(chatID, "لطفا نام فرایند را وارد کنید") // "Please enter the process name"
		return
	}

	// Check if the user is in the middle of creating a process
	builder, exists := h.processBuilderService.GetBuilder(userID)
	if !exists {
		return // Not in a process creation flow
	}

	switch builder.CurrentStep {
	case "name":
		if h.processBuilderService.SetProcessName(userID, update.Message.Text) {
			sendMessage(chatID, "توضیحات فرایند را وارد کنید") // "Enter the process description"
		}
	case "description":
		if h.processBuilderService.SetProcessDescription(userID, update.Message.Text) {
			process, success := h.processBuilderService.CompleteProcess(userID) // Returns *models.Process
			if success {
				err := h.processService.CreateProcess(process) // Expects *models.Process
				if err != nil {
					sendMessage(chatID, "خطا در ایجاد فرآیند. لطفا دوباره تلاش کنید.") // "Error creating process. Please try again."
					return
				}
				response := fmt.Sprintf("فرایند با موفقیت ایجاد شد!\\nنام: %s\\nتوضیحات: %s", process.Name, process.Description) // "Process created successfully! Name: %s Description: %s"
				sendMessage(chatID, response)
			}
		}
	}
}

// HandleProcessExecution handles messages related to executing a process.
func (h *ProcessHandler) HandleProcessExecution(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if update.Message.Text == "شروع فرایند" { // "Start Process" in Farsi
		processes, err := h.processService.GetProcessesByUserID(userID)
		if err != nil {
			sendMessage(chatID, "خطا در دریافت فرایندها. لطفا دوباره تلاش کنید.")
			return
		}
		if len(processes) == 0 {
			sendMessage(chatID, "شما هیچ فرایندی برای اجرا ندارید.")
			return
		}

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
		msg := tgbotapi.NewMessage(chatID, "لطفا فرایند را برای اجرا انتخاب کنید:")
		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil { // Keep bot.Send for messages with specific inline keyboards
			log.Printf("Error sending process selection message: %v", err)
		}
	}
}

// HandleProcessCommands handles general process-related commands.
func (h *ProcessHandler) HandleProcessCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if update.Message.Text == "فرایند ها" { // "Processes" in Farsi
		processes, err := h.processService.GetProcessesByUserID(userID)
		if err != nil {
			sendMessage(chatID, "خطا در دریافت فرایندها. لطفا دوباره تلاش کنید.")
			return
		}
		if len(processes) == 0 {
			sendMessage(chatID, "شما هیچ فرایندی ندارید.")
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
		if _, err := bot.Send(msg); err != nil { // Keep bot.Send for specific inline keyboards
			log.Printf("Error sending process list: %v", err)
		}
		return
	}
}

// HandleProcessCallback handles callback queries for processes.
func (h *ProcessHandler) HandleProcessCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID
	// userID := update.CallbackQuery.From.ID // Not directly used in the restored logic for start_process, but good to have if needed

	if strings.HasPrefix(data, "start_process_") {
		processIDStr := strings.TrimPrefix(data, "start_process_")
		processID, err := strconv.ParseUint(processIDStr, 10, 32)
		if err != nil {
			sendMessage(chatID, "خطای داخلی: شناسه فرایند نامعتبر است.")
			log.Printf("Error parsing processID from callback: %v", err)
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "شناسه نامعتبر")
			bot.Request(callbackAns)
			return
		}

		// Start a new process execution using h.processService
		execution, err := h.processService.StartProcessExecution(uint(processID))
		if err != nil {
			sendMessage(chatID, "خطا در شروع فرایند. لطفا دوباره تلاش کنید.")
			log.Printf("Error starting process execution: %v", err)
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "خطا در شروع")
			bot.Request(callbackAns)
			return
		}

		// Get all tasks for the process
		tasks, err := h.taskService.GetTasksByProcessID(uint(processID))
		if err != nil {
			sendMessage(chatID, "خطا در دریافت وظایف اولیه فرایند.")
			log.Printf("Error getting tasks by processID %d: %v", processID, err)
			// execution might still be valid, but no tasks to start.
			// Decide if to rollback or just inform user.
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "خطا در وظایف")
			bot.Request(callbackAns)
			return
		}

		var firstTasks []models.Task
		for _, task := range tasks {
			preTaskIDs, err := h.taskService.GetTaskPrerequisites(task.ID)
			if err != nil {
				sendMessage(chatID, fmt.Sprintf("خطا در بررسی پیش‌نیازهای وظیفه %s.", task.Title))
				log.Printf("Error getting prerequisites for task %d: %v", task.ID, err)
				// Potentially skip this task or halt process? For now, log and continue.
				continue
			}
			if len(preTaskIDs) == 0 {
				firstTasks = append(firstTasks, task)
			}
		}

		if len(firstTasks) == 0 {
			sendMessage(chatID, "فرایند شروع شد، اما هیچ وظیفه اولیه‌ای برای آن تعریف نشده یا پیش‌نیازها مشکل دارند.")
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "وظیفه اولیه نیست")
			bot.Request(callbackAns)
			return
		}

		var startedTasksInfo strings.Builder
		startedTasksCount := 0
		for _, task := range firstTasks {
			// StartTaskExecution now also notifies users for the task.
			// It returns models.TaskExecution
			taskExecution, err := h.taskService.StartTaskExecution(execution.ID, task.ID)
			if err != nil {
				// Log error, maybe inform user specifically about this task failing to start
				log.Printf("Error starting task execution for task %d in process exec %d: %v", task.ID, execution.ID, err)
				sendMessage(chatID, fmt.Sprintf("خطا در شروع وظیفه %s. %s", task.Title, err.Error()))
				continue // Try to start other initial tasks
			}

			// Add the task execution to the pending list of the process execution
			if err := h.processService.AddPendingTask(execution.ID, taskExecution.ID); err != nil {
				log.Printf("Error adding pending task %d to process exec %d: %v", taskExecution.ID, execution.ID, err)
				sendMessage(chatID, fmt.Sprintf("خطا در ثبت وظیفه %s به عنوان وظیفه در انتظار.", task.Title))
				// This is more critical, as the task started but isn't tracked by process properly.
				// Potentially needs rollback or admin alert.
				continue
			}
			startedTasksInfo.WriteString(fmt.Sprintf("- %s\n", task.Title))
			startedTasksCount++
		}

		if startedTasksCount > 0 {
			responseMsg := fmt.Sprintf("فرایند با شناسه اجرای %d شروع شد.\nوظایف اولیه زیر آغاز شدند و به گروه‌های مربوطه اطلاع داده شد:\n%s", execution.ID, startedTasksInfo.String())
			sendMessage(chatID, responseMsg)
		} else {
			sendMessage(chatID, fmt.Sprintf("فرایند با شناسه اجرای %d شروع شد، اما هیچ وظیفه اولیه‌ای با موفقیت آغاز نشد.", execution.ID))
		}
		callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "فرایند شروع شد")
		bot.Request(callbackAns)

	} else if strings.HasPrefix(data, "view_process_") {
		processID, err := strconv.ParseUint(strings.TrimPrefix(data, "view_process_"), 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه فرایند.")
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "شناسه نامعتبر")
			bot.Request(callbackAns)
			return
		}

		tasks, err := h.taskService.GetTasksByProcessID(uint(processID))
		if err != nil {
			sendMessage(chatID, "خطا در دریافت وظایف فرایند.")
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "خطا در وظایف")
			bot.Request(callbackAns)
			return
		}

		if len(tasks) == 0 {
			sendMessage(chatID, "این فرایند هیچ وظیفه‌ای ندارد.")
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "بدون وظیفه")
			bot.Request(callbackAns)
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
		if _, errBot := bot.Send(msg); errBot != nil {
			log.Printf("Error sending task list for process: %v", errBot)
		}
		callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "وظایف نمایش داده شد")
		bot.Request(callbackAns)

	} else if strings.HasPrefix(data, "view_task_") { // This was in original process_handler, might belong to task_handler
		taskID, err := strconv.ParseUint(strings.TrimPrefix(data, "view_task_"), 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه وظیفه.")
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "شناسه وظیفه نامعتبر")
			bot.Request(callbackAns)
			return
		}

		task, err := h.taskService.GetTaskByID(uint(taskID))
		if err != nil {
			sendMessage(chatID, "خطا در دریافت اطلاعات وظیفه.")
			callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "خطا در دریافت وظیفه")
			bot.Request(callbackAns)
			return
		}

		taskDetails := fmt.Sprintf("عنوان: %s\\nتوضیحات: %s",
			task.Title,
			task.Description,
		)
		sendMessage(chatID, taskDetails)
		callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "جزئیات وظیفه")
		bot.Request(callbackAns)
	} else {
		log.Printf("ProcessHandler received unhandled callback data: %s from chatID %d", data, chatID)
		callbackAns := tgbotapi.NewCallback(update.CallbackQuery.ID, "عملیات نامشخص")
		bot.Request(callbackAns)
	}
}

// No specific inline keyboards defined here for now, as main.go handles the persistent keyboard.
// If HandleProcessCreation callback for confirmation was still here, its keyboard would be defined here.
