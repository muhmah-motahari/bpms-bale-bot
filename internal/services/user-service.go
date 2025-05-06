package service

import (
	"bbb/internal/dto"
	"bbb/internal/models"
	"bbb/internal/repository"
)

type (
	UserService interface {
		SaveOrUpdateUser(message dto.Message) error
		GetUserByID(userID int64) (*models.User, error)
	}

	userService struct {
		userRepo repository.UserRepository
	}
)

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		userRepo: repo,
	}
}

func (s *userService) SaveOrUpdateUser(message dto.Message) error {
	// Check if user exists
	existingUser, err := s.userRepo.GetByID(message.From.ID)
	if err != nil {
		// User doesn't exist, create new one
		user := models.User{
			ID:        message.From.ID,
			FirstName: message.From.First_name,
			LastName:  message.From.Last_name,
			Username:  message.From.Username,
		}
		return s.userRepo.Save(&user)
	}

	// User exists, update if needed
	if existingUser.FirstName != message.From.First_name ||
		existingUser.LastName != message.From.Last_name ||
		existingUser.Username != message.From.Username {
		existingUser.FirstName = message.From.First_name
		existingUser.LastName = message.From.Last_name
		existingUser.Username = message.From.Username
		return s.userRepo.Update(existingUser)
	}

	return nil
}

func (s *userService) GetUserByID(userID int64) (*models.User, error) {
	return s.userRepo.GetByID(userID)
}
