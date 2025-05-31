package handlers

import (
	"bbb/internal/models"
	service "bbb/internal/services"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskHandler struct {
	taskService        service.TaskService
	taskBuilderService *service.TaskBuilderService
	processService     service.ProcessService
	teamService        service.TeamService
}

func NewTaskHandler(
	taskService service.TaskService,
	taskBuilderService *service.TaskBuilderService,
	processService service.ProcessService,
	teamService service.TeamService,
) *TaskHandler {
	return &TaskHandler{
		taskService:        taskService,
		taskBuilderService: taskBuilderService,
		processService:     processService,
		teamService:        teamService,
	}
}

func (h *TaskHandler) HandleTaskCreation(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if update.Message.Text == "وظیفه جدید" {
		processes, err := h.processService.GetProcessesByUserID(userID)
		if err != nil {
			sendMessage(chatID, "خطا در دریافت فرآیندها. لطفا دوباره تلاش کنید.")
			return
		}

		if len(processes) == 0 {
			sendMessage(chatID, "شما هیچ فرآیندی ندارید. ابتدا یک فرآیند ایجاد کنید.")
			return
		}

		var keyboardRows [][]tgbotapi.InlineKeyboardButton
		for _, process := range processes {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(process.Name, fmt.Sprintf("select_process_%d", process.ID)),
			}
			keyboardRows = append(keyboardRows, row)
		}

		msg := tgbotapi.NewMessage(chatID, "لطفا فرآیند مورد نظر را انتخاب کنید:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
		if _, errSend := bot.Send(msg); errSend != nil {
			log.Printf("Error sending process selection for task creation: %v", errSend)
		}
		return
	}

	builder, exists := h.taskBuilderService.GetBuilder(userID)
	if !exists || update.Message.Text == "" {
		return
	}

	switch builder.CurrentStep {
	case "title":
		if !h.taskBuilderService.SetTitle(userID, update.Message.Text) {
			sendMessage(chatID, "خطا در تنظیم عنوان وظیفه.")
			return
		}
		sendMessage(chatID, "لطفا توضیحات وظیفه را وارد کنید:")
	case "description":
		if !h.taskBuilderService.SetDescription(userID, update.Message.Text) {
			sendMessage(chatID, "خطا در تنظیم توضیحات وظیفه.")
			return
		}
		targetProcessTasks, err := h.taskService.GetTasksByProcessID(builder.ProcessID)
		if err != nil || len(targetProcessTasks) == 0 {
			sendMessage(chatID, "هیچ وظیفه دیگری در این فرایند برای انتخاب به عنوان پیش‌نیاز یافت نشد. آیا این وظیفه پیش‌نیاز دارد؟ (بله/خیر - فعلا فقط خیر پشتیبانی میشود یا متن 'skip')")
			return
		}

		var prereqKeyboardRows [][]tgbotapi.InlineKeyboardButton
		for _, task := range targetProcessTasks {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(task.Title, fmt.Sprintf("add_prerequisite_%d", task.ID)),
			}
			prereqKeyboardRows = append(prereqKeyboardRows, row)
		}
		prereqKeyboardRows = append(prereqKeyboardRows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("اتمام افزودن پیش‌نیازها", "done_prerequisites")))

		msg := tgbotapi.NewMessage(chatID, "لطفا وظایف پیش‌نیاز را انتخاب کنید (یا اتمام را بزنید):")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(prereqKeyboardRows...)
		if _, errSend := bot.Send(msg); errSend != nil {
			log.Printf("Error sending prerequisite selection: %v", errSend)
		}

	case "prerequisites":
		responseText := strings.ToLower(update.Message.Text)
		if responseText == "خیر" || responseText == "no" || responseText == "skip" || responseText == "تمام" {
			if !h.taskBuilderService.SetHasMorePrerequisites(userID, false) {
				sendMessage(chatID, "خطا در پردازش اتمام پیش‌نیازها.")
				return
			}
			teams, err := h.teamService.GetTeamsByOwnerID(userID)
			if err != nil || len(teams) == 0 {
				sendMessage(chatID, "تیمی برای تخصیص وظیفه یافت نشد. لطفا ابتدا یک تیم ایجاد کنید.")
				return
			}
			var teamKeyboardRows [][]tgbotapi.InlineKeyboardButton
			for _, team := range teams {
				row := []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData(team.Name, fmt.Sprintf("select_team_%d", team.ID)),
				}
				teamKeyboardRows = append(teamKeyboardRows, row)
			}
			msg := tgbotapi.NewMessage(chatID, "لطفا تیم مسئول این وظیفه را انتخاب کنید:")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(teamKeyboardRows...)
			if _, errSend := bot.Send(msg); errSend != nil {
				log.Printf("Error sending team selection: %v", errSend)
			}
		} else if responseText == "بله" || responseText == "yes" {
			sendMessage(chatID, "لطفا شناسه وظیفه پیش‌نیاز بعدی را وارد کنید یا از لیست بالا انتخاب کنید (اگر نمایش داده شده بود).")
		} else {
			if prereqID, err := strconv.ParseUint(responseText, 10, 64); err == nil {
				if !h.taskBuilderService.AddPrerequisite(userID, uint(prereqID)) {
					sendMessage(chatID, "خطا در افزودن پیش‌نیاز با شناسه.")
				} else {
					sendMessage(chatID, "پیش‌نیاز افزوده شد. آیا وظیفه پیش‌نیاز دیگری دارد؟ (بله/خیر) یا شناسه بعدی را وارد کنید.")
				}
			} else {
				sendMessage(chatID, "پاسخ نامعتبر. لطفا 'بله'، 'خیر'، 'skip' یا شناسه وظیفه پیش‌نیاز را وارد کنید.")
			}
		}
	default:
		break
	}
}

func (h *TaskHandler) HandleCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, sendMessage func(chatID int64, text string)) {
	if update.CallbackQuery == nil {
		return
	}
	data := update.CallbackQuery.Data
	userID := update.CallbackQuery.From.ID
	chatID := update.CallbackQuery.Message.Chat.ID
	var callbackMsg string

	switch {
	case strings.HasPrefix(data, "select_process_"):
		processID, err := strconv.ParseUint(strings.TrimPrefix(data, "select_process_"), 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه فرآیند.")
			callbackMsg = "خطا در شناسه"
			break
		}
		h.taskBuilderService.StartTask(userID)
		if !h.taskBuilderService.SetProcess(userID, uint(processID)) {
			sendMessage(chatID, "خطا در شروع ایجاد وظیفه برای این فرآیند.")
			callbackMsg = "خطا در تنظیم فرآیند"
			break
		}
		sendMessage(chatID, "لطفا عنوان وظیفه را وارد کنید:")
		callbackMsg = "فرآیند انتخاب شد"

	case strings.HasPrefix(data, "add_prerequisite_"):
		prerequisiteID, err := strconv.ParseUint(strings.TrimPrefix(data, "add_prerequisite_"), 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه وظیفه پیش‌نیاز.")
			callbackMsg = "خطا در شناسه پیش‌نیاز"
			break
		}
		if !h.taskBuilderService.AddPrerequisite(userID, uint(prerequisiteID)) {
			sendMessage(chatID, "خطا در افزودن پیش‌نیاز.")
			callbackMsg = "خطا در افزودن پیش‌نیاز"
		} else {
			sendMessage(chatID, "پیش‌نیاز افزوده شد. برای افزودن مورد بعدی انتخاب کنید یا 'اتمام' را بزنید.")
			callbackMsg = "پیش‌نیاز افزوده شد"
		}

	case data == "done_prerequisites":
		if !h.taskBuilderService.SetHasMorePrerequisites(userID, false) {
			sendMessage(chatID, "خطا در پردازش اتمام پیش‌نیازها.")
			callbackMsg = "خطا"
			break
		}
		teams, err := h.teamService.GetAllTeams()
		if err != nil || len(teams) == 0 {
			sendMessage(chatID, "تیم برای تخصیص وظیفه یافت نشد. لطفا ابتدا یک تیم ایجاد کنید.")
			callbackMsg = "تیمی یافت نشد"
			break
		}
		var teamKeyboardRows [][]tgbotapi.InlineKeyboardButton
		for _, team := range teams {
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(team.Name, fmt.Sprintf("select_team_%d", team.ID)),
			}
			teamKeyboardRows = append(teamKeyboardRows, row)
		}
		msg := tgbotapi.NewMessage(chatID, "لطفا تیم مسئول این وظیفه را انتخاب کنید:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(teamKeyboardRows...)
		if _, errSend := bot.Send(msg); errSend != nil {
			log.Printf("Error sending team selection: %v", errSend)
		}
		callbackMsg = "انتخاب تیم"

	case strings.HasPrefix(data, "select_team_"):
		teamID, err := strconv.ParseUint(strings.TrimPrefix(data, "select_team_"), 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه تیم.")
			callbackMsg = "خطای شناسه تیم"
			break
		}
		if !h.taskBuilderService.SetTeam(userID, uint(teamID)) {
			sendMessage(chatID, "خطا در تنظیم تیم برای وظیفه.")
			callbackMsg = "خطا در تنظیم تیم"
			break
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("بله", "set_final_true"),
				tgbotapi.NewInlineKeyboardButtonData("خیر", "set_final_false"),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "آیا این وظیفه پایانی است؟")
		msg.ReplyMarkup = keyboard
		if _, errSend := bot.Send(msg); errSend != nil {
			log.Printf("Error sending final task confirmation: %v", errSend)
		}
		callbackMsg = "تیم انتخاب شد"

	case strings.HasPrefix(data, "set_final_"):
		isFinal := strings.HasSuffix(data, "true")
		if !h.taskBuilderService.SetIsFinal(userID, isFinal) {
			sendMessage(chatID, "خطا در تنظیم وضعیت نهایی وظیفه.")
			callbackMsg = "خطا در وضعیت نهایی"
			break
		}
		task, prerequisites, success := h.taskBuilderService.CompleteTask(userID)
		if !success {
			sendMessage(chatID, "خطا در تکمیل ایجاد وظیفه.")
			callbackMsg = "خطا در تکمیل وظیفه"
			break
		}
		if err := h.taskService.CreateTask(task); err != nil {
			sendMessage(chatID, fmt.Sprintf("خطا در ذخیره وظیفه: %s", err.Error()))
			callbackMsg = "خطا در ذخیره وظیفه"
			break
		}
		for _, prereqID := range prerequisites {
			if err := h.taskService.AddPrerequisite(task.ID, prereqID); err != nil {
				sendMessage(chatID, fmt.Sprintf("خطا در افزودن پیش‌نیاز %d به وظیفه %d: %s", prereqID, task.ID, err.Error()))
				log.Printf("Error adding prerequisite %d to task %d: %v", prereqID, task.ID, err)
			}
		}
		sendMessage(chatID, fmt.Sprintf("وظیفه '%s' با موفقیت ایجاد شد.", task.Title))
		callbackMsg = "وظیفه ایجاد شد"

	case strings.HasPrefix(data, "take_task_"):
		taskExecutionID, err := strconv.ParseUint(strings.TrimPrefix(data, "take_task_"), 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه وظیفه در حال اجرا.")
			callbackMsg = "خطای شناسه"
			break
		}
		err = h.taskService.AssignTask(uint(taskExecutionID), userID)
		if err != nil {
			sendMessage(chatID, "خطا در به عهده گرفتن وظیفه: "+err.Error())
			callbackMsg = "خطا در تخصیص"
			break
		}
		var keyboardRows [][]tgbotapi.InlineKeyboardButton
		row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(
			"تکمیل وظیفه",
			fmt.Sprintf("complete_task_%d", taskExecutionID),
		)}
		keyboardRows = append(keyboardRows, row)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
		msg := tgbotapi.NewMessage(chatID, "وظیفه با موفقیت به شما اختصاص داده شد.\n\nهنگامی که وظیفه را انجام دادید روی دکمه «تکمیل وظیفه» کلیک کنید.")
		msg.ReplyMarkup = keyboard
		if _, errSend := bot.Send(msg); errSend != nil {
			log.Printf("Error sending complete task button: %v", errSend)
		}
		callbackMsg = "وظیفه تخصیص داده شد"

	case strings.HasPrefix(data, "complete_task_"):
		taskExecID, err := strconv.ParseUint(strings.TrimPrefix(data, "complete_task_"), 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه وظیفه در حال اجرا.")
			callbackMsg = "خطای شناسه"
			break
		}

		taskExec, err := h.taskService.GetTaskExecutionByID(uint(taskExecID))
		if err != nil || taskExec == nil {
			sendMessage(chatID, "خطا: اطلاعات اجرای وظیفه یافت نشد.")
			callbackMsg = "اجرای وظیفه یافت نشد"
			break
		}

		err = h.taskService.CompleteTask(uint(taskExecID), userID)
		if err != nil {
			sendMessage(chatID, "خطا در تکمیل وظیفه: "+err.Error())
			callbackMsg = "خطا در تکمیل"
			break
		}
		sendMessage(chatID, "وظیفه با موفقیت تکمیل شد.")
		callbackMsg = "وظیفه تکمیل شد"

		fmt.Println("taskExec.Task: ", taskExec.Task)
		fmt.Println("taskExec.Task.Process: ", taskExec.Task.Process)
		fmt.Println("taskExec.Task.Process.UserID: ", taskExec.Task.Process.UserID)
		ownerID := taskExec.Task.Process.UserID
		if ownerID != chatID {
			sendMessage(ownerID, fmt.Sprintf(`
				☑️ اعلان تکمیل وظیفه
					- فرایند: %s
					- شماره فرایند اجرایی: %s
					- وظیفه: %s
					- انجام دهنده: %s %s
					- تاریخ: %s
				`,
				taskExec.Task.Process.Name,
				taskExec.ProcessExecutionID,
				taskExec.Task.Title,
				taskExec.User.FirstName,
				taskExec.User.LastName,
				taskExec.CompletedAt))
		}

		if isFinal, _ := h.taskService.IsFinalTask(taskExec.TaskID); isFinal {
			processExec, _ := h.processService.GetProcessExecutionByID(taskExec.ProcessExecutionID)
			processExec.Status = models.ProcessExecutionStatusCompleted
			completedTime := time.Now()
			processExec.CompletedAt = &completedTime
			if errUpdate := h.processService.UpdateProcessExecution(processExec); errUpdate != nil {
				sendMessage(chatID, "خطا در بروزرسانی وضعیت نهایی فرایند.")
				log.Printf("Error updating process execution status to completed: %v", errUpdate)
			} else {
				sendMessage(chatID, "فرایند والد نیز با موفقیت تکمیل شد.")
				if ownerID != chatID {
					sendMessage(ownerID, fmt.Sprintf(
						`✅ اعلان تکمیل فرایند
							- فرایند: %s
							- شماره فرایند اجرایی: %s
							- تاریخ: %s
						`,
						taskExec.Task.Process.Name,
						taskExec.ProcessExecutionID,
						processExec.CompletedAt))
				}
			}
		}

		DependentTasks, depErr := h.taskService.GetDependentTasks(taskExec.TaskID)
		if depErr != nil {
			log.Printf("Error getting dependent tasks for task %d: %v", taskExec.TaskID, depErr)
		} else {
			for _, task := range DependentTasks {
				if _, startErr := h.taskService.StartTaskExecution(taskExec.ProcessExecutionID, task.ID); startErr != nil {
					sendMessage(chatID, fmt.Sprintf("خطا در شروع وظیفه وابسته %s: %s", task.Title, startErr.Error()))
					log.Printf("Error starting dependent task %d: %v", task.ID, startErr)
				}
			}
		}
		break

	case strings.HasPrefix(data, "view_task_"):
		taskID, err := strconv.ParseUint(strings.TrimPrefix(data, "view_task_"), 10, 64)
		if err != nil {
			sendMessage(chatID, "خطا در پردازش شناسه وظیفه.")
			callbackMsg = "خطای شناسه"
			break
		}
		task, err := h.taskService.GetTaskByID(uint(taskID))
		if err != nil {
			sendMessage(chatID, "خطا در دریافت اطلاعات وظیفه.")
			callbackMsg = "خطا در دریافت وظیفه"
			break
		}
		taskDetails := fmt.Sprintf("عنوان: %s\nتوضیحات: %s", task.Title, task.Description)
		sendMessage(chatID, taskDetails)
		callbackMsg = "جزئیات وظیفه"
	}

	if callbackMsg != "" {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, callbackMsg)
		bot.Request(callback)
	}
}
