package handlers

import (
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
}

func NewProcessHandler(
	processService service.ProcessService,
	processBuilderService *service.ProcessBuilderService,
	processExecService service.ProcessExecutionService,
) *ProcessHandler {
	return &ProcessHandler{
		processService:        processService,
		processBuilderService: processBuilderService,
		processExecService:    processExecService,
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

	if update.Message.Text == "شروع فرایند" {
		if processes, err := h.processService.GetProcessesByUserID(update.Message.From.ID); err == nil {
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

	if update.Message.Text == "تکمیل وظیفه" {
		var keyboardRows [][]tgbotapi.InlineKeyboardButton
		runningProcesses := h.processExecService.GetRunningProcessesByUserID(update.Message.From.ID)
		for processID := range runningProcesses {
			process, err := h.processService.GetProcessByID(processID)
			if err != nil {
				continue
			}
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					process.Name,
					fmt.Sprintf("complete_task_%d", processID),
				),
			}
			keyboardRows = append(keyboardRows, row)
		}
		if len(keyboardRows) == 0 {
			msg := tgbotapi.NewMessage(chatID, "هیچ فرایندی در حال اجرا نیست.")
			bot.Send(msg)
			return
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)
		msg := tgbotapi.NewMessage(chatID, "لطفا فرایند را انتخاب کنید:")
		msg.ReplyMarkup = keyboard
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
		if err := h.processExecService.StartProcess(uint(processID)); err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در شروع فرایند. لطفا دوباره تلاش کنید.")
			bot.Send(msg)
			return
		}

		currentTask, err := h.processExecService.GetCurrentTask(uint(processID))
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "خطا در دریافت وظیفه فعلی. لطفا دوباره تلاش کنید.")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("فرایند با موفقیت شروع شد!\nوظیفه فعلی: %s\nتوضیحات: %s",
				currentTask.Title, currentTask.Description))
		bot.Send(msg)
	}
}
