package service

import (
	"bbb/internal/models"
	"bbb/internal/repository"
	"errors"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type (
	TaskService interface {
		CreateTask(task *models.Task) error
		StartTaskExecution(processExecutionID, taskID uint) (models.TaskExecution, error)
		GetTaskByID(taskID uint) (*models.Task, error)
		GetTasksByProcessID(processID uint) ([]models.Task, error)
		AssignTask(taskExecutionID uint, userID int64) error
		CompleteTask(taskExecutionID uint, userID int64) error
		GetUserTasks(userID int64) ([]models.TaskExecution, error)
		AddPrerequisite(taskID uint, prerequisiteID uint) error
		GetTaskPrerequisites(taskID uint) ([]uint, error)
		IsFinalTask(taskID uint) (bool, error)
		GetDependentTasks(taskID uint) ([]models.Task, error)
		GetTaskExecutionByID(taskExecutionID uint) (*models.TaskExecution, error)
	}

	taskService struct {
		repo           repository.TaskRepository
		groupService   GroupService
		processService ProcessService
		bot            *tgbotapi.BotAPI
	}
)

func NewTaskService(repo repository.TaskRepository, groupService GroupService, processService ProcessService, bot *tgbotapi.BotAPI) TaskService {
	return &taskService{
		repo:           repo,
		groupService:   groupService,
		processService: processService,
		bot:            bot,
	}
}

func (s *taskService) CreateTask(task *models.Task) error {
	return s.repo.Save(task)
}

func (s *taskService) GetTaskByID(taskID uint) (*models.Task, error) {
	return s.repo.GetByID(taskID)
}

func (s *taskService) GetTasksByProcessID(processID uint) ([]models.Task, error) {
	return s.repo.GetByProcessID(processID)
}

func (s *taskService) AssignTask(taskExecutionID uint, userID int64) error {
	taskExecution, err := s.repo.GetTaskExecutionByID(taskExecutionID)
	if err != nil {
		return err
	}
	if taskExecution == nil {
		return errors.New("وظیفه‌ی درحال اجرا با این شناسه یافت نشد!")
	}

	if taskExecution.Status != models.TaskStatusPending {
		return errors.New("وظیفه را فرد دیگری به عهده گرفت‌:(")
	}

	taskExecution.Status = models.TaskStatusAssigned
	taskExecution.UserID = &userID
	now := time.Now()
	taskExecution.AssignedAt = &now

	// Move task from pending to in-progress
	processExecution, err := s.processService.GetProcessExecutionByID(taskExecution.ProcessExecutionID)
	if err != nil {
		return err
	}

	// Remove from pending
	for i, id := range processExecution.PendingTaskExecutionIDs {
		if id == taskExecutionID {
			processExecution.PendingTaskExecutionIDs = append(processExecution.PendingTaskExecutionIDs[:i], processExecution.PendingTaskExecutionIDs[i+1:]...)
			break
		}
	}

	// Add to in-progress
	processExecution.InProgressTaskExecutionIDs = append(processExecution.InProgressTaskExecutionIDs, taskExecutionID)

	if err := s.processService.UpdateProcessExecution(processExecution); err != nil {
		return err
	}

	return s.repo.UpdateTaskExecution(taskExecution)
}

func (s *taskService) CompleteTask(taskExecutionID uint, userID int64) error {
	taskExecution, err := s.repo.GetTaskExecutionByID(taskExecutionID)
	if err != nil {
		return err
	}
	if taskExecution == nil {
		return errors.New("task execution not found")
	}

	if taskExecution.Status != models.TaskStatusAssigned || taskExecution.UserID == nil || *taskExecution.UserID != userID {
		return errors.New("task is not assigned to you")
	}

	taskExecution.Status = models.TaskStatusCompleted
	now := time.Now()
	taskExecution.CompletedAt = &now

	// Move task from in-progress to completed
	processExecution, err := s.processService.GetProcessExecutionByID(taskExecution.ProcessExecutionID)
	if err != nil {
		return err
	}

	// Remove from in-progress
	for i, id := range processExecution.InProgressTaskExecutionIDs {
		if id == taskExecutionID {
			processExecution.InProgressTaskExecutionIDs = append(processExecution.InProgressTaskExecutionIDs[:i], processExecution.InProgressTaskExecutionIDs[i+1:]...)
			break
		}
	}

	// Add to completed
	processExecution.CompletedTaskExecutionIDs = append(processExecution.CompletedTaskExecutionIDs, taskExecutionID)

	if err := s.processService.UpdateProcessExecution(processExecution); err != nil {
		return err
	}

	return s.repo.UpdateTaskExecution(taskExecution)
}

func (s *taskService) GetUserTasks(userID int64) ([]models.TaskExecution, error) {
	return s.repo.GetTaskExecutionsByUserID(userID)
}

func (s *taskService) AddPrerequisite(taskID uint, prerequisiteID uint) error {
	return s.repo.AddPrerequisite(taskID, prerequisiteID)
}

func (s *taskService) GetTaskPrerequisites(taskID uint) ([]uint, error) {
	return s.repo.GetPrerequisites(taskID)
}

func (s *taskService) StartTaskExecution(processExecutionID, taskID uint) (models.TaskExecution, error) {
	// Get prerequisites
	preTaskIDs, err := s.GetTaskPrerequisites(taskID)
	if err != nil {
		return models.TaskExecution{}, fmt.Errorf("error getting prerequisites: %v", err)
	}

	// Get process execution
	processExecution, err := s.processService.GetProcessExecutionByID(processExecutionID)
	if err != nil {
		return models.TaskExecution{}, fmt.Errorf("error getting process execution: %v", err)
	}

	// Check if all prerequisites are completed
	for _, preTaskID := range preTaskIDs {
		found := false
		for _, TEID := range processExecution.CompletedTaskExecutionIDs {
			te, err := s.GetTaskExecutionByID(TEID)
			if err != nil {
				return models.TaskExecution{}, fmt.Errorf("error getting task executions: %v", err)
			}
			if te.TaskID == preTaskID {
				found = true
				break
			}
		}
		if !found {
			return models.TaskExecution{}, fmt.Errorf("prerequisite task %d is not completed in this process execution", preTaskID)
		}
	}

	// Start the task execution
	taskExecution := models.TaskExecution{
		TaskID:             taskID,
		ProcessExecutionID: processExecutionID,
		Status:             models.TaskStatusPending,
	}

	if err := s.repo.SaveTaskExecution(&taskExecution); err != nil {
		return models.TaskExecution{}, fmt.Errorf("error starting task execution: %v", err)
	}

	// Add to pending tasks
	processExecution.PendingTaskExecutionIDs = append(processExecution.PendingTaskExecutionIDs, taskExecution.ID)
	if err := s.processService.UpdateProcessExecution(processExecution); err != nil {
		return models.TaskExecution{}, fmt.Errorf("error updating process execution: %v", err)
	}

	// Get task details for notification
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return models.TaskExecution{}, fmt.Errorf("error getting task: %v", err)
	}
	if task == nil {
		return models.TaskExecution{}, errors.New("task not found")
	}

	if task.GroupID == nil {
		return models.TaskExecution{}, errors.New("task has no group assigned")
	}

	group, err := s.groupService.GetGroupByID(*task.GroupID)
	if err != nil {
		return models.TaskExecution{}, fmt.Errorf("error getting group: %v", err)
	}
	if group == nil {
		return models.TaskExecution{}, errors.New("group not found")
	}

	members, err := s.groupService.GetGroupMembers(group.ID)
	if err != nil {
		return models.TaskExecution{}, fmt.Errorf("error getting group members: %v", err)
	}
	if len(members) == 0 {
		return models.TaskExecution{}, errors.New("group has no members")
	}

	taskMsg := fmt.Sprintf("وظیفه با اطلاعات زیر فعال شده است، اگر تمایل دارید که انجام دهید اعلام کنید.\n\nعنوان: %s\nتوضیحات: %s",
		task.Title, task.Description)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("به عهده گرفتن وظیفه", fmt.Sprintf("take_task_%d", taskExecution.ID)),
		),
	)

	for _, member := range members {
		msg := tgbotapi.NewMessage(member.ID, taskMsg)
		msg.ReplyMarkup = keyboard
		if _, err := s.bot.Send(msg); err != nil {
			return models.TaskExecution{}, fmt.Errorf("error sending message to user %d: %v", member.ID, err)
		}
	}
	return taskExecution, nil
}

func (s *taskService) IsFinalTask(taskID uint) (bool, error) {
	task, err := s.GetTaskByID(taskID)
	if err != nil {
		return false, err
	}
	return task.IsFinal, nil
}

func (s *taskService) GetDependentTasks(taskID uint) ([]models.Task, error) {
	return s.repo.GetDependentTasks(taskID)
}

func (s *taskService) GetTaskExecutionByID(taskExecutionID uint) (*models.TaskExecution, error) {
	return s.repo.GetTaskExecutionByID(taskExecutionID)
}
