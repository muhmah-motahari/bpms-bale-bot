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
		SaveProcessExecution(execution *models.ProcessExecution) error
		GetProcessExecutionByID(id uint) (*models.ProcessExecution, error)
		GetProcessExecutionsByProcessID(processID uint) ([]models.ProcessExecution, error)
		UpdateProcessExecution(execution *models.ProcessExecution) error
		GetPendingProcessExecutions() ([]models.ProcessExecution, error)
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

func (r *processRepository) SaveProcessExecution(execution *models.ProcessExecution) error {
	return r.db.Create(execution).Error
}

func (r *processRepository) GetProcessExecutionByID(id uint) (*models.ProcessExecution, error) {
	var execution models.ProcessExecution
	if err := r.db.First(&execution, id).Error; err != nil {
		return nil, err
	}
	return &execution, nil
}

func (r *processRepository) GetProcessExecutionsByProcessID(processID uint) ([]models.ProcessExecution, error) {
	var executions []models.ProcessExecution
	if err := r.db.Where("process_id = ?", processID).Find(&executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}

func (r *processRepository) UpdateProcessExecution(execution *models.ProcessExecution) error {
	return r.db.Save(execution).Error
}

func (r *processRepository) GetPendingProcessExecutions() ([]models.ProcessExecution, error) {
	var executions []models.ProcessExecution
	if err := r.db.Where("status = ?", models.ProcessExecutionStatusPending).Find(&executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}
