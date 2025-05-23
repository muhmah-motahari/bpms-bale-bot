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

	db *gorm.DB = configs.SetUpDatabaseConnection(env)

	processRepo           repository.ProcessRepository   = repository.NewProcessRepository(db)
	processService        service.ProcessService         = service.NewProcessService(processRepo)
	processBuilderService *service.ProcessBuilderService = service.NewProcessBuilderService()

	taskRepo           repository.TaskRepository   = repository.NewTaskRepository(db)
	taskBuilderService *service.TaskBuilderService = service.NewTaskBuilderService()

	userRepo     repository.UserRepository  = repository.NewUserRepository(db)
	groupRepo    repository.GroupRepository = repository.NewGroupRepository(db)
	groupService service.GroupService       = service.NewGroupService(groupRepo, userRepo)

	userService = service.NewUserService(userRepo)

	processExecutionService = service.NewProcessExecutionService(processRepo, taskRepo, groupRepo)
)

func main() {
	APIEndpoint := "https://tapi.bale.ai/bot%s/%s"
	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(env.Token, APIEndpoint)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	taskService := service.NewTaskService(taskRepo, groupService, bot)

	// Initialize handlers
	processHandler := handlers.NewProcessHandler(processService, processBuilderService, processExecutionService, taskService)
	taskHandler := handlers.NewTaskHandler(taskService, taskBuilderService, processService, groupService)
	groupHandler := handlers.NewGroupHandler(groupService, userService)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Loop through each update.
	for update := range updates {
		// Check if we've gotten a message update.
		if update.Message != nil {
			// Save or update user
			if err := userService.SaveOrUpdateUser(dto.Message{
				From: dto.User{
					ID:         update.Message.From.ID,
					First_name: update.Message.From.FirstName,
					Last_name:  update.Message.From.LastName,
					Username:   update.Message.From.UserName,
				},
			}); err != nil {
				log.Printf("Error saving/updating user: %v", err)
			}

			// Handle process creation
			processHandler.HandleProcessCreation(bot, update)
			// Handle process execution
			processHandler.HandleProcessExecution(bot, update)
			// Handle process commands
			processHandler.HandleProcessCommands(bot, update)
			// Handle task creation
			taskHandler.HandleTaskCreation(bot, update)
			// Handle task commands
			taskHandler.HandleTaskCommands(bot, update)
			// Handle group commands
			groupHandler.HandleGroupCommands(bot, update)
		} else if update.CallbackQuery != nil {
			// Handle process callbacks
			processHandler.HandleProcessCallback(bot, update)
			// Handle task callbacks
			taskHandler.HandleTaskCommands(bot, update)
			// Handle group callbacks
			groupHandler.HandleGroupCallback(bot, update)
		}
	}
}
