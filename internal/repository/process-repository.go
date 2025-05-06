package repository

import (
	"bbb/internal/models"

	"gorm.io/gorm"
)

type (
	ProcessRepository interface {
		Save(req models.Process) error
		GetAll() ([]models.Process, error)
		GetByID(processID uint) (*models.Process, error)
		GetByUserID(userID int64) ([]models.Process, error)
	}

	processRepository struct {
		db *gorm.DB
	}
)

func NewProcessRepository(db *gorm.DB) ProcessRepository {
	return &processRepository{
		db: db,
	}
}

func (r *processRepository) Save(req models.Process) error {
	if req.ID == 0 {
		if err := r.db.Create(&req).Error; err != nil {
			return err
		}
	} else {
		if err := r.db.Save(&req).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *processRepository) GetAll() ([]models.Process, error) {
	var processes []models.Process
	if err := r.db.Find(&processes).Error; err != nil {
		return nil, err
	}
	return processes, nil
}

func (r *processRepository) GetByID(processID uint) (*models.Process, error) {
	var process models.Process
	if err := r.db.First(&process, processID).Error; err != nil {
		return nil, err
	}
	return &process, nil
}

func (r *processRepository) GetByUserID(userID int64) ([]models.Process, error) {
	var processes []models.Process
	err := r.db.Where("user_id = ?", userID).Find(&processes).Error
	return processes, err
}
