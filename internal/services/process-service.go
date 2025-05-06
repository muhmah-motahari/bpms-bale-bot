package service

import (
	"bbb/internal/models"
	"bbb/internal/repository"
	"errors"
)

type ProcessService interface {
	CreateProcess(process *models.Process) error
	GetProcessByID(id uint) (*models.Process, error)
	GetProcessesByUserID(userID int64) ([]models.Process, error)
	GetAllProcesses() ([]models.Process, error)
}

type processService struct {
	repo repository.ProcessRepository
}

func NewProcessService(repo repository.ProcessRepository) ProcessService {
	return &processService{repo: repo}
}

func (s *processService) CreateProcess(process *models.Process) error {
	if process.Name == "" {
		return errors.New("process name is required")
	}

	if process.Description == "" {
		return errors.New("process description is required")
	}

	return s.repo.Save(*process)
}

func (s *processService) GetProcessByID(id uint) (*models.Process, error) {
	return s.repo.GetByID(id)
}

func (s *processService) GetProcessesByUserID(userID int64) ([]models.Process, error) {
	return s.repo.GetByUserID(userID)
}

func (s *processService) GetAllProcesses() ([]models.Process, error) {
	return s.repo.GetAll()
}
