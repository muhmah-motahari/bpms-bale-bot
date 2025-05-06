package service

import (
	"bbb/internal/dto"
	"bbb/internal/models"
	"bbb/internal/repository"
	"errors"
	"fmt"
	"math/rand"
)

type GroupService interface {
	CreateGroup(group *models.Group) error
	GetGroupByID(id uint) (*models.Group, error)
	GetGroupMembers(groupID uint) ([]models.User, error)
	AddMember(groupID uint, userID int64) error
	RemoveMember(groupID uint, userID int64) error
	GenerateJoinKey() string
	JoinGroup(userID int64, joinKey string) error
	GetAllGroups() ([]models.Group, error)
}

type groupService struct {
	repo     repository.GroupRepository
	userRepo repository.UserRepository
}

func NewGroupService(repo repository.GroupRepository, userRepo repository.UserRepository) GroupService {
	return &groupService{repo: repo, userRepo: userRepo}
}

func (s *groupService) CreateGroup(group *models.Group) error {
	group.JoinKey = s.GenerateJoinKey()
	return s.repo.Save(*group)
}

func (s *groupService) GetGroupByID(id uint) (*models.Group, error) {
	if id == 0 {
		return nil, errors.New("group ID is zero")
	}
	return s.repo.GetByID(id)
}

func (s *groupService) GetGroupMembers(groupID uint) ([]models.User, error) {
	return s.repo.GetMembers(groupID)
}

func (s *groupService) AddMember(groupID uint, userID int64) error {
	return s.repo.AddMember(groupID, userID)
}

func (s *groupService) RemoveMember(groupID uint, userID int64) error {
	return s.repo.RemoveMember(groupID, userID)
}

func (s *groupService) GenerateJoinKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *groupService) JoinGroup(userID int64, joinKey string) error {
	group, err := s.repo.GetByJoinKey(joinKey)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("invalid join key")
	}
	return s.repo.AddMember(group.ID, userID)
}

func (g *groupService) AddToNewGroup(message dto.Message) {
	// Save Group repository
	group := models.Group{
		ID:      uint(message.Chat.ID),
		Name:    message.Chat.Title,
		JoinKey: g.GenerateJoinKey(),
	}
	err := g.repo.Save(group)
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
	userGroup := models.UserGroups{
		UserID:  user.ID,
		GroupID: group.ID,
	}
	err = g.repo.SaveUserGroup(userGroup)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (s *groupService) GetAllGroups() ([]models.Group, error) {
	return s.repo.GetAll()
}
