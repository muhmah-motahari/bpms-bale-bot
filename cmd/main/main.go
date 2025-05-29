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
	groupRepo   repository.GroupRepository   = repository.NewGroupRepository(db)

	// Bot
	bot *tgbotapi.BotAPI

	// Services
	userService             = service.NewUserService(userRepo)
	groupService            = service.NewGroupService(groupRepo, userRepo)
	processService          = service.NewProcessService(processRepo)
	processBuilderService   = service.NewProcessBuilderService()
	processExecutionService = service.NewProcessExecutionService(processRepo, taskRepo, groupRepo)
	taskBuilderService      = service.NewTaskBuilderService()
	taskService             service.TaskService
	groupBuilderService     = service.NewGroupBuilderService()

	// Handlers
	groupHandler   *handlers.GroupHandler
	taskHandler    *handlers.TaskHandler
	processHandler *handlers.ProcessHandler
	helpHandler    *handlers.HelpHandler
	startHandler   *handlers.StartHandler
)

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("فرایند جدید"), tgbotapi.NewKeyboardButton("شروع فرایند"), tgbotapi.NewKeyboardButton("فرایند ها")),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("وظیفه جدید"), tgbotapi.NewKeyboardButton("وظایف من"), tgbotapi.NewKeyboardButton("لیست گروه ها")),
	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("گروه جدید"), tgbotapi.NewKeyboardButton("عضویت در گروه"), tgbotapi.NewKeyboardButton("راهنما")),
)

func init() {
	APIEndpoint := "https://tapi.bale.ai/bot%s/%s"
	var err error
	bot, err = tgbotapi.NewBotAPIWithAPIEndpoint(env.Token, APIEndpoint)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Initialize taskService with bot
	taskService = service.NewTaskService(taskRepo, groupService, processService, bot)

	// Initialize handlers
	groupHandler = handlers.NewGroupHandler(groupService, userService, groupBuilderService)
	taskHandler = handlers.NewTaskHandler(taskService, taskBuilderService, processService, groupService)
	processHandler = handlers.NewProcessHandler(processService, processBuilderService, processExecutionService, taskService)
	helpHandler = handlers.NewHelpHandler(env)
	startHandler = handlers.NewStartHandler()
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
			groupHandler.HandleGroupCommands(bot, update, sendMessageWithKeyboard)
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
			groupHandler.HandleGroupCallback(bot, update, sendCallbackMessageWithKeyboard)
		}
	}
}
