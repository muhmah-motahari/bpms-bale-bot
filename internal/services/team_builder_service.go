package service

import (
	"bbb/internal/models"
	"sync"
)

// TeamBuilderService handles the team creation flow
type TeamBuilderService struct {
	builders map[int64]*TeamBuilder
	mu       sync.RWMutex
}

// TeamBuilder represents the state of a team being built
type TeamBuilder struct {
	Name      string
	IsJoining bool // true if user is in the process of joining a team
}

// NewTeamBuilderService creates a new TeamBuilderService
func NewTeamBuilderService() *TeamBuilderService {
	return &TeamBuilderService{
		builders: make(map[int64]*TeamBuilder),
	}
}

// StartTeam starts a new team creation process for a user
func (s *TeamBuilderService) StartTeam(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.builders[userID] = &TeamBuilder{}
}

// StartJoinTeam starts the process of joining a team
func (s *TeamBuilderService) StartJoinTeam(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.builders[userID] = &TeamBuilder{
		IsJoining: true,
	}
}

// GetBuilder returns the builder for a user if it exists
func (s *TeamBuilderService) GetBuilder(userID int64) (*TeamBuilder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	builder, exists := s.builders[userID]
	return builder, exists
}

// SetName sets the name for the team being built
func (s *TeamBuilderService) SetName(userID int64, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	builder, exists := s.builders[userID]
	if !exists {
		return false
	}
	builder.Name = name
	return true
}

// CompleteTeam completes the team creation and returns the team
func (s *TeamBuilderService) CompleteTeam(userID int64) (*models.Team, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	builder, exists := s.builders[userID]
	if !exists {
		return nil, false
	}

	team := &models.Team{
		Name:    builder.Name,
		OwnerID: userID,
	}

	delete(s.builders, userID)
	return team, true
}

// CompleteJoinTeam completes the join team process and returns the join key
func (s *TeamBuilderService) CompleteJoinTeam(userID int64) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	builder, exists := s.builders[userID]
	if !exists || !builder.IsJoining {
		return "", false
	}

	joinKey := builder.Name
	delete(s.builders, userID)
	return joinKey, true
}
