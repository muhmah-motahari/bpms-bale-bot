package service

import (
	"bbb/internal/models"
	"sync"
)

// GroupBuilderService handles the group creation flow
type GroupBuilderService struct {
	builders map[int64]*GroupBuilder
	mu       sync.RWMutex
}

// GroupBuilder represents the state of a group being built
type GroupBuilder struct {
	Name string
}

// NewGroupBuilderService creates a new GroupBuilderService
func NewGroupBuilderService() *GroupBuilderService {
	return &GroupBuilderService{
		builders: make(map[int64]*GroupBuilder),
	}
}

// StartGroup starts a new group creation process for a user
func (s *GroupBuilderService) StartGroup(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.builders[userID] = &GroupBuilder{}
}

// GetBuilder returns the builder for a user if it exists
func (s *GroupBuilderService) GetBuilder(userID int64) (*GroupBuilder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	builder, exists := s.builders[userID]
	return builder, exists
}

// SetName sets the name for the group being built
func (s *GroupBuilderService) SetName(userID int64, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	builder, exists := s.builders[userID]
	if !exists {
		return false
	}
	builder.Name = name
	return true
}

// CompleteGroup completes the group creation and returns the group
func (s *GroupBuilderService) CompleteGroup(userID int64) (*models.Group, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	builder, exists := s.builders[userID]
	if !exists {
		return nil, false
	}

	group := &models.Group{
		Name:    builder.Name,
		OwnerID: userID,
	}

	delete(s.builders, userID)
	return group, true
}
