package main

import (
	"bbb/configs"
	"bbb/internal/dto"
	"bbb/internal/handlers"
	"bbb/internal/repository"
	service "bbb/internal/services"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

var (
	env configs.Env = configs.NewEnv()
	db  *gorm.DB    = configs.SetUpDatabaseConnection(env)

	// Repositories
	processRepo repository.ProcessRepository = repository.NewProcessRepository(db)
	taskRepo    repository.TaskRepository    = repository.NewTaskRepository(db)
	userRepo    repository.UserRepository    = repository.NewUserRepository(db)
	teamRepo    repository.TeamRepository    = repository.NewTeamRepository(db)

	// Bot
	bot *tgbotapi.BotAPI

	// Services
	userService             = service.NewUserService(userRepo)
	teamService             = service.NewTeamService(teamRepo, userRepo)
	processService          = service.NewProcessService(processRepo)
	processBuilderService   = service.NewProcessBuilderService()
	processExecutionService = service.NewProcessExecutionService(processRepo, taskRepo, teamRepo)
	taskBuilderService      = service.NewTaskBuilderService()
	taskService             service.TaskService
	teamBuilderService      = service.NewTeamBuilderService()

	// Handlers
	teamHandler    *handlers.TeamHandler
	taskHandler    *handlers.TaskHandler
	processHandler *handlers.ProcessHandler
	helpHandler    *handlers.HelpHandler
	startHandler   *handlers.StartHandler
)

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("فرایند جدید"), tgbotapi.NewKeyboardButton("شروع فرایند"), tgbotapi.NewKeyboardButton("فرایند ها")),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("وظیفه جدید"), tgbotapi.NewKeyboardButton("وظایف من"), tgbotapi.NewKeyboardButton("لیست تیم ها")),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("تیم جدید"), tgbotapi.NewKeyboardButton("عضویت در تیم"), tgbotapi.NewKeyboardButton("راهنما")),
)

func init() {
	var err error
	bot, err = tgbotapi.NewBotAPIWithAPIEndpoint(env.Token, env.APIEndpoint)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Initialize taskService with bot
	taskService = service.NewTaskService(taskRepo, teamService, processService, bot)

	// Initialize handlers
	teamHandler = handlers.NewTeamHandler(teamService, userService, teamBuilderService)
	taskHandler = handlers.NewTaskHandler(taskService, taskBuilderService, processService, teamService)
	processHandler = handlers.NewProcessHandler(processService, processBuilderService, processExecutionService, taskService)
	helpHandler = handlers.NewHelpHandler(env, &mainKeyboard)
	startHandler = handlers.NewStartHandler(&mainKeyboard)
}

func main() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			// Generic message sender with main keyboard
			sendMessageWithKeyboard := func(chatID int64, text string) {
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ReplyMarkup = mainKeyboard
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending message with keyboard: %v", err)
				}
			}

			// Save or update user
			if err := userService.SaveOrUpdateUser(dto.Message{
				From: dto.User{
					ID:         update.Message.From.ID,
					First_name: update.Message.From.FirstName,
					Last_name:  update.Message.From.LastName,
					Username:   update.Message.From.UserName,
				},
				Chat: dto.Chat{ID: update.Message.Chat.ID, Title: update.Message.Chat.Title, Type: update.Message.Chat.Type},
			}); err != nil {
				log.Printf("Error saving/updating user: %v", err)
			}

			// Pass bot, update, and the sender function to handlers
			startHandler.HandleStartCommand(bot, update, sendMessageWithKeyboard)
			processHandler.HandleProcessCreation(bot, update, sendMessageWithKeyboard)
			processHandler.HandleProcessExecution(bot, update, sendMessageWithKeyboard)
			processHandler.HandleProcessCommands(bot, update, sendMessageWithKeyboard)
			taskHandler.HandleTaskCreation(bot, update, sendMessageWithKeyboard)
			teamHandler.HandleTeamCommands(bot, update, sendMessageWithKeyboard)
			helpHandler.HandleHelpCommand(bot, update, sendMessageWithKeyboard)

		} else if update.CallbackQuery != nil {
			// Generic message sender for callback responses (might also include main keyboard)
			sendCallbackMessageWithKeyboard := func(chatID int64, text string) {
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ReplyMarkup = mainKeyboard
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending callback message with keyboard: %v", err)
				}
			}

			// Pass bot, update, and the sender function to callback handlers
			processHandler.HandleProcessCallback(bot, update, sendCallbackMessageWithKeyboard)
			taskHandler.HandleCallbackQuery(bot, update, sendCallbackMessageWithKeyboard)
			teamHandler.HandleTeamCallback(bot, update, sendCallbackMessageWithKeyboard)
		}
	}
}
