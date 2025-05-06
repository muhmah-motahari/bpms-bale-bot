package service

import (
	"bbb/internal/models"
	"bbb/internal/repository"
	"errors"
	"time"
)

type ProcessService interface {
	CreateProcess(process *models.Process) error
	GetProcessByID(id uint) (*models.Process, error)
	GetProcessesByUserID(userID int64) ([]models.Process, error)
	GetAllProcesses() ([]models.Process, error)
	StartProcessExecution(processID uint) (*models.ProcessExecution, error)
	GetProcessExecutionByID(id uint) (*models.ProcessExecution, error)
	GetProcessExecutionsByProcessID(processID uint) ([]models.ProcessExecution, error)
	UpdateProcessExecution(execution *models.ProcessExecution) error
	GetPendingProcessExecutions() ([]models.ProcessExecution, error)
	AddPendingTask(executionID uint, taskID uint) error
	RemovePendingTask(executionID uint, taskID uint) error
}

type processService struct {
	repo repository.ProcessRepository
}

func NewProcessService(repo repository.ProcessRepository) ProcessService {
	return &processService{repo: repo}
}

func (s *processService) CreateProcess(process *models.Process) error {
	if process.Name == "" {
		return errors.New("process name is required")
	}

	if process.Description == "" {
		return errors.New("process description is required")
	}

	return s.repo.Save(*process)
}

func (s *processService) GetProcessByID(id uint) (*models.Process, error) {
	return s.repo.GetByID(id)
}

func (s *processService) GetProcessesByUserID(userID int64) ([]models.Process, error) {
	return s.repo.GetByUserID(userID)
}

func (s *processService) GetAllProcesses() ([]models.Process, error) {
	return s.repo.GetAll()
}

func (s *processService) StartProcessExecution(processID uint) (*models.ProcessExecution, error) {
	process, err := s.repo.GetByID(processID)
	if err != nil {
		return nil, err
	}
	if process == nil {
		return nil, errors.New("process not found")
	}

	execution := &models.ProcessExecution{
		ProcessID:      processID,
		Status:         models.ProcessExecutionStatusPending,
		PendingTaskIDs: make([]uint, 0),
		StartedAt:      time.Now(),
	}

	if err := s.repo.SaveProcessExecution(execution); err != nil {
		return nil, err
	}

	return execution, nil
}

func (s *processService) GetProcessExecutionByID(id uint) (*models.ProcessExecution, error) {
	return s.repo.GetProcessExecutionByID(id)
}

func (s *processService) GetProcessExecutionsByProcessID(processID uint) ([]models.ProcessExecution, error) {
	return s.repo.GetProcessExecutionsByProcessID(processID)
}

func (s *processService) UpdateProcessExecution(execution *models.ProcessExecution) error {
	return s.repo.UpdateProcessExecution(execution)
}

func (s *processService) GetPendingProcessExecutions() ([]models.ProcessExecution, error) {
	return s.repo.GetPendingProcessExecutions()
}

func (s *processService) AddPendingTask(executionID uint, taskID uint) error {
	execution, err := s.repo.GetProcessExecutionByID(executionID)
	if err != nil {
		return err
	}

	// Check if task is already in pending list
	for _, id := range execution.PendingTaskIDs {
		if id == taskID {
			return nil // Task already in pending list
		}
	}

	execution.PendingTaskIDs = append(execution.PendingTaskIDs, taskID)
	return s.repo.UpdateProcessExecution(execution)
}

func (s *processService) RemovePendingTask(executionID uint, taskID uint) error {
	execution, err := s.repo.GetProcessExecutionByID(executionID)
	if err != nil {
		return err
	}

	// Remove task from pending list
	newPendingTasks := make([]uint, 0)
	for _, id := range execution.PendingTaskIDs {
		if id != taskID {
			newPendingTasks = append(newPendingTasks, id)
		}
	}

	execution.PendingTaskIDs = newPendingTasks
	return s.repo.UpdateProcessExecution(execution)
}
