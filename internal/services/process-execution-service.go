package service

import (
	"bbb/internal/models"
	"bbb/internal/repository"
	"errors"
	"sync"
)

type (
	ProcessExecutionService interface {
		StartProcess(processID uint) error
		GetCurrentTask(processID uint) (*models.Task, error)
		CompleteCurrentTask(processID uint) error
		IsProcessRunning(processID uint) bool
		GetRunningProcesses() map[uint]*processExecution
		GetRunningProcessesByUserID(userID int64) map[uint]*processExecution
	}

	processExecutionService struct {
		processRepo repository.ProcessRepository
		taskRepo    repository.TaskRepository
		groupRepo   repository.GroupRepository
		// Map to store running processes and their current task
		runningProcesses map[uint]*processExecution
		mu               sync.RWMutex
	}

	processExecution struct {
		ProcessID     uint
		CurrentTaskID uint
		TaskOrder     []uint // Ordered list of task IDs
	}
)

func NewProcessExecutionService(
	processRepo repository.ProcessRepository,
	taskRepo repository.TaskRepository,
	groupRepo repository.GroupRepository,
) ProcessExecutionService {
	return &processExecutionService{
		processRepo:      processRepo,
		taskRepo:         taskRepo,
		groupRepo:        groupRepo,
		runningProcesses: make(map[uint]*processExecution),
	}
}

func (s *processExecutionService) StartProcess(processID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if process is already running
	if _, exists := s.runningProcesses[processID]; exists {
		return errors.New("process is already running")
	}

	// Get all tasks for the process
	tasks, err := s.taskRepo.GetByProcessID(processID)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return errors.New("process has no tasks")
	}

	// Create ordered list of tasks based on prerequisites
	taskOrder, err := s.createTaskOrder(tasks)
	if err != nil {
		return err
	}

	// Start with the first task
	s.runningProcesses[processID] = &processExecution{
		ProcessID:     processID,
		CurrentTaskID: taskOrder[0],
		TaskOrder:     taskOrder,
	}

	return nil
}

func (s *processExecutionService) GetCurrentTask(processID uint) (*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	execution, exists := s.runningProcesses[processID]
	if !exists {
		return nil, errors.New("process is not running")
	}

	return s.taskRepo.GetByID(execution.CurrentTaskID)
}

func (s *processExecutionService) CompleteCurrentTask(processID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	execution, exists := s.runningProcesses[processID]
	if !exists {
		return errors.New("process is not running")
	}

	// Find current task index
	currentIndex := -1
	for i, taskID := range execution.TaskOrder {
		if taskID == execution.CurrentTaskID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return errors.New("current task not found in task order")
	}

	// If this was the last task, remove the process from running processes
	if currentIndex == len(execution.TaskOrder)-1 {
		delete(s.runningProcesses, processID)
		return nil
	}

	// Move to next task
	execution.CurrentTaskID = execution.TaskOrder[currentIndex+1]
	return nil
}

func (s *processExecutionService) IsProcessRunning(processID uint) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.runningProcesses[processID]
	return exists
}

// Helper function to create ordered list of tasks based on prerequisites
func (s *processExecutionService) createTaskOrder(tasks []models.Task) ([]uint, error) {
	// Create a map of task IDs to their prerequisites
	taskPrerequisites := make(map[uint][]uint)
	for _, task := range tasks {
		prerequisites, err := s.taskRepo.GetPrerequisites(task.ID)
		if err != nil {
			return nil, err
		}
		taskPrerequisites[task.ID] = prerequisites
	}

	// Create ordered list using topological sort
	var order []uint
	visited := make(map[uint]bool)
	temp := make(map[uint]bool)

	var visit func(uint) error
	visit = func(taskID uint) error {
		if temp[taskID] {
			return errors.New("circular dependency detected")
		}
		if visited[taskID] {
			return nil
		}
		temp[taskID] = true

		for _, prereq := range taskPrerequisites[taskID] {
			if err := visit(prereq); err != nil {
				return err
			}
		}

		temp[taskID] = false
		visited[taskID] = true
		order = append(order, taskID)
		return nil
	}

	for _, task := range tasks {
		if !visited[task.ID] {
			if err := visit(task.ID); err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}

func (s *processExecutionService) GetRunningProcesses() map[uint]*processExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy of the map to avoid concurrent access issues
	result := make(map[uint]*processExecution)
	for k, v := range s.runningProcesses {
		result[k] = v
	}
	return result
}

func (s *processExecutionService) GetRunningProcessesByUserID(userID int64) map[uint]*processExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy of the map to avoid concurrent access issues
	result := make(map[uint]*processExecution)

	// Get all processes for the user
	processes, err := s.processRepo.GetByUserID(userID)
	if err != nil {
		return result
	}

	// Filter running processes that belong to the user
	for processID, execution := range s.runningProcesses {
		for _, process := range processes {
			if process.ID == processID {
				result[processID] = execution
				break
			}
		}
	}

	return result
}
