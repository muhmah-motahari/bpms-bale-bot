package repository

import (
	"bbb/internal/models"

	"gorm.io/gorm"
)

type (
	UserRepository interface {
		Save(req *models.User) error
		SaveUserTeam(req models.UserTeams) error
		GetByID(userID int64) (*models.User, error)
		Update(req *models.User) error
	}

	userRepository struct {
		db *gorm.DB
	}
)

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Save(req *models.User) error {
	return r.db.Create(req).Error
}

func (r *userRepository) Update(req *models.User) error {
	return r.db.Save(req).Error
}

func (r *userRepository) SaveUserTeam(req models.UserTeams) error {
	if err := r.db.Save(&req).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetByID(userID int64) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
