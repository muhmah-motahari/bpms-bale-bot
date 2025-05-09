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
		StartTaskExecution(taskID uint) error
		GetTaskByID(taskID uint) (*models.Task, error)
		GetTasksByProcessID(processID uint) ([]models.Task, error)
		AssignTask(taskExecutionID uint, userID int64) error
		CompleteTask(taskExecutionID uint, userID int64) error
		GetUserTasks(userID int64) ([]models.TaskExecution, error)
		AddPrerequisite(taskID uint, prerequisiteID uint) error
		GetTaskPrerequisites(taskID uint) ([]uint, error)
		IsFinalTask(taskID uint) (bool, error)
	}

	taskService struct {
		repo         repository.TaskRepository
		groupService GroupService
		bot          *tgbotapi.BotAPI
	}
)

func NewTaskService(repo repository.TaskRepository, groupService GroupService, bot *tgbotapi.BotAPI) TaskService {
	return &taskService{
		repo:         repo,
		groupService: groupService,
		bot:          bot,
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
		return errors.New("task execution not found")
	}

	if taskExecution.Status != models.TaskStatusPending {
		return errors.New("task execution is not available for assignment")
	}

	taskExecution.Status = models.TaskStatusAssigned
	taskExecution.UserID = &userID
	now := time.Now()
	taskExecution.AssignedAt = &now

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

func (s *taskService) StartTaskExecution(taskID uint) error {
	if err := s.repo.StartTaskExecution(taskID); err != nil {
		return fmt.Errorf("error starting task execution: %v", err)
	}

	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("error getting task: %v", err)
	}
	if task == nil {
		return errors.New("task not found")
	}

	if task.GroupID == nil {
		return errors.New("task has no group assigned")
	}

	group, err := s.groupService.GetGroupByID(*task.GroupID)
	if err != nil {
		return fmt.Errorf("error getting group: %v", err)
	}
	if group == nil {
		return errors.New("group not found")
	}

	members, err := s.groupService.GetGroupMembers(group.ID)
	if err != nil {
		return fmt.Errorf("error getting group members: %v", err)
	}
	if len(members) == 0 {
		return errors.New("group has no members")
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
		if _, err := s.bot.Send(msg); err != nil {
			return fmt.Errorf("error sending message to user %d: %v", member.ID, err)
		}
	}
	return nil
}

func (s *taskService) IsFinalTask(taskID uint) (bool, error) {
	task, err := s.GetTaskByID(taskID)
	if err != nil {
		return false, err
	}
	return task.IsFinal, nil
}
