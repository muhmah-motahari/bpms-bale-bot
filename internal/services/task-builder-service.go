package service

import (
	"bbb/internal/models"
	"sync"
)

type TaskBuilderService struct {
	builders map[int64]*models.TaskBuilder
	mu       sync.RWMutex
}

func NewTaskBuilderService() *TaskBuilderService {
	return &TaskBuilderService{
		builders: make(map[int64]*models.TaskBuilder),
	}
}

func (s *TaskBuilderService) StartTask(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.builders[userID] = &models.TaskBuilder{
		UserID:               userID,
		CurrentStep:          "process",
		Task:                 models.Task{},
		Prerequisites:        make([]uint, 0),
		HasMorePrerequisites: true,
	}
}

func (s *TaskBuilderService) GetBuilder(userID int64) (*models.TaskBuilder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	builder, exists := s.builders[userID]
	return builder, exists
}

func (s *TaskBuilderService) SetProcess(userID int64, processID uint) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists && builder.CurrentStep == "process" {
		builder.ProcessID = processID
		builder.Task.ProcessID = processID
		builder.CurrentStep = "title"
		return true
	}
	return false
}

func (s *TaskBuilderService) SetTitle(userID int64, title string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists && builder.CurrentStep == "title" {
		builder.Task.Title = title
		builder.CurrentStep = "description"
		return true
	}
	return false
}

func (s *TaskBuilderService) SetDescription(userID int64, description string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists && builder.CurrentStep == "description" {
		builder.Task.Description = description
		builder.CurrentStep = "prerequisites"
		return true
	}
	return false
}

func (s *TaskBuilderService) AddPrerequisite(userID int64, prerequisiteID uint) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists && builder.CurrentStep == "prerequisites" {
		builder.Prerequisites = append(builder.Prerequisites, prerequisiteID)
		return true
	}
	return false
}

func (s *TaskBuilderService) SetHasMorePrerequisites(userID int64, hasMore bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists && builder.CurrentStep == "prerequisites" {
		builder.HasMorePrerequisites = hasMore
		if !hasMore {
			builder.CurrentStep = "group"
		}
		return true
	}
	return false
}

func (s *TaskBuilderService) SetGroup(userID int64, groupID uint) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists && builder.CurrentStep == "group" {
		builder.Task.GroupID = &groupID
		return true
	}
	return false
}

func (s *TaskBuilderService) CompleteTask(userID int64) (*models.Task, []uint, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists {
		task := builder.Task
		prerequisites := builder.Prerequisites
		delete(s.builders, userID)
		return &task, prerequisites, true
	}
	return nil, nil, false
}

func (s *TaskBuilderService) CancelTask(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.builders, userID)
}
