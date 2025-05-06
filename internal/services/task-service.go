package service

import (
	"bbb/internal/models"
	"bbb/internal/repository"
	"errors"
	"time"
)

type (
	TaskService interface {
		CreateTask(task *models.Task) error
		GetTaskByID(taskID uint) (*models.Task, error)
		GetTasksByProcessID(processID uint) ([]models.Task, error)
		AssignTask(taskID uint, userID int64) error
		CompleteTask(taskID uint, userID int64) error
		GetUserTasks(userID int64) ([]models.Task, error)
		GetNextAvailableTasks(processID uint) ([]models.Task, error)
		AddPrerequisite(taskID uint, prerequisiteID uint) error
		GetTaskPrerequisites(taskID uint) ([]uint, error)
	}

	taskService struct {
		repo repository.TaskRepository
	}
)

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(task *models.Task) error {
	task.Status = models.TaskStatusPending
	return s.repo.Save(task)
}

func (s *taskService) GetTaskByID(taskID uint) (*models.Task, error) {
	return s.repo.GetByID(taskID)
}

func (s *taskService) GetTasksByProcessID(processID uint) ([]models.Task, error) {
	return s.repo.GetByProcessID(processID)
}

func (s *taskService) AssignTask(taskID uint, userID int64) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if task.Status != models.TaskStatusPending {
		return errors.New("task is not available for assignment")
	}

	task.Status = models.TaskStatusAssigned
	task.AssigneeID = &userID
	now := time.Now()
	task.AssignedAt = &now

	return s.repo.Save(task)
}

func (s *taskService) CompleteTask(taskID uint, userID int64) error {
	task, err := s.repo.GetByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	if task.Status != models.TaskStatusAssigned || task.AssigneeID == nil || *task.AssigneeID != userID {
		return errors.New("task is not assigned to you")
	}

	task.Status = models.TaskStatusCompleted
	now := time.Now()
	task.CompletedAt = &now

	return s.repo.Save(task)
}

func (s *taskService) GetUserTasks(userID int64) ([]models.Task, error) {
	return s.repo.GetByAssigneeID(userID)
}

func (s *taskService) GetNextAvailableTasks(processID uint) ([]models.Task, error) {
	return s.repo.GetNextAvailableTasks(processID)
}

func (s *taskService) AddPrerequisite(taskID uint, prerequisiteID uint) error {
	return s.repo.AddPrerequisite(taskID, prerequisiteID)
}

func (s *taskService) GetTaskPrerequisites(taskID uint) ([]uint, error) {
	return s.repo.GetPrerequisites(taskID)
}
