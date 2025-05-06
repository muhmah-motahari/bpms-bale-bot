package service

import (
	"bbb/internal/models"
	"sync"
)

type ProcessBuilderService struct {
	builders map[int64]*models.ProcessBuilder
	mu       sync.RWMutex
}

func NewProcessBuilderService() *ProcessBuilderService {
	return &ProcessBuilderService{
		builders: make(map[int64]*models.ProcessBuilder),
	}
}

func (s *ProcessBuilderService) StartProcess(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	process := models.Process{}

	s.builders[userID] = &models.ProcessBuilder{
		UserID:      userID,
		CurrentStep: "name",
		Process:     &process,
	}
}

func (s *ProcessBuilderService) GetBuilder(userID int64) (*models.ProcessBuilder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	builder, exists := s.builders[userID]
	return builder, exists
}

func (s *ProcessBuilderService) SetProcessName(userID int64, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists && builder.CurrentStep == "name" {
		builder.Process.Name = name
		builder.CurrentStep = "description"
		return true
	}
	return false
}

func (s *ProcessBuilderService) SetProcessDescription(userID int64, description string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists && builder.CurrentStep == "description" {
		builder.Process.Description = description
		return true
	}
	return false
}

func (s *ProcessBuilderService) CompleteProcess(userID int64) (*models.Process, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if builder, exists := s.builders[userID]; exists {
		process := models.Process{
			Name:        builder.Process.Name,
			Description: builder.Process.Description,
			UserID:      builder.UserID,
		}
		delete(s.builders, userID)
		return &process, true
	}
	return nil, false
}

func (s *ProcessBuilderService) CancelProcess(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.builders, userID)
}
