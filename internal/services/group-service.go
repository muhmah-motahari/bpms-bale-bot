package service

import (
	"bbb/internal/dto"
	"bbb/internal/models"
	"bbb/internal/repository"
	"errors"
	"fmt"
	"math/rand"
)

type TeamService interface {
	CreateTeam(team *models.Team) error
	GetTeamByID(id uint) (*models.Team, error)
	GetTeamMembers(teamID uint) ([]models.User, error)
	AddMember(teamID uint, userID int64) error
	RemoveMember(teamID uint, userID int64) error
	GenerateJoinKey() string
	JoinTeam(userID int64, joinKey string) error
	GetAllTeams() ([]models.Team, error)
	GetTeamsByOwnerID(ownerID int64) ([]*models.Team, error)
}

type teamService struct {
	repo     repository.TeamRepository
	userRepo repository.UserRepository
}

func NewGroupService(repo repository.TeamRepository, userRepo repository.UserRepository) TeamService {
	return &teamService{repo: repo, userRepo: userRepo}
}

func (s *teamService) CreateTeam(team *models.Team) error {
	team.JoinKey = s.GenerateJoinKey()
	return s.repo.Save(*team)
}

func (s *teamService) GetTeamByID(id uint) (*models.Team, error) {
	if id == 0 {
		return nil, errors.New("team ID is zero")
	}
	return s.repo.GetByID(id)
}

func (s *teamService) GetTeamMembers(teamID uint) ([]models.User, error) {
	return s.repo.GetMembers(teamID)
}

func (s *teamService) AddMember(teamID uint, userID int64) error {
	return s.repo.AddMember(teamID, userID)
}

func (s *teamService) RemoveMember(teamID uint, userID int64) error {
	return s.repo.RemoveMember(teamID, userID)
}

func (s *teamService) GenerateJoinKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *teamService) JoinTeam(userID int64, joinKey string) error {
	team, err := s.repo.GetByJoinKey(joinKey)
	if err != nil {
		return err
	}
	if team == nil {
		return errors.New("invalid join key")
	}
	return s.repo.AddMember(team.ID, userID)
}

func (g *teamService) AddToNewTeam(message dto.Message) {
	// Save Team repository
	team := models.Team{
		ID:      uint(message.Chat.ID),
		Name:    message.Chat.Title,
		JoinKey: g.GenerateJoinKey(),
	}
	err := g.repo.Save(team)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	// Save User owner repository
	user := models.User{
		ID:        message.From.ID,
		FirstName: message.From.First_name,
		LastName:  message.From.Last_name,
		Username:  message.From.Username,
	}

	err = g.userRepo.Save(&user)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	// Save UserGroups
	userGroup := models.UserTeams{
		UserID: user.ID,
		TeamID: team.ID,
	}
	err = g.repo.SaveUserGroup(userGroup)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (s *teamService) GetAllTeams() ([]models.Team, error) {
	return s.repo.GetAll()
}

func (s *teamService) GetTeamsByOwnerID(ownerID int64) ([]*models.Team, error) {
	return s.repo.GetGroupsByOwnerID(ownerID)
}
