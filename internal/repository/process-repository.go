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
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Save the process execution
	if err := tx.Create(execution).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Save pending tasks
	for _, taskID := range execution.PendingTaskIDs {
		pendingTask := models.PendingTask{
			ProcessExecutionID: execution.ID,
			TaskID:             taskID,
		}
		if err := tx.Create(&pendingTask).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *processRepository) GetProcessExecutionByID(id uint) (*models.ProcessExecution, error) {
	var execution models.ProcessExecution
	if err := r.db.First(&execution, id).Error; err != nil {
		return nil, err
	}

	// Get pending tasks
	var pendingTasks []models.PendingTask
	if err := r.db.Where("process_execution_id = ?", id).Find(&pendingTasks).Error; err != nil {
		return nil, err
	}

	// Convert to task IDs
	execution.PendingTaskIDs = make([]uint, len(pendingTasks))
	for i, pt := range pendingTasks {
		execution.PendingTaskIDs[i] = pt.TaskID
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
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Update the process execution
	if err := tx.Save(execution).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete existing pending tasks
	if err := tx.Where("process_execution_id = ?", execution.ID).Delete(&models.PendingTask{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Save new pending tasks
	for _, taskID := range execution.PendingTaskIDs {
		pendingTask := models.PendingTask{
			ProcessExecutionID: execution.ID,
			TaskID:             taskID,
		}
		if err := tx.Create(&pendingTask).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *processRepository) GetPendingProcessExecutions() ([]models.ProcessExecution, error) {
	var executions []models.ProcessExecution
	if err := r.db.Where("status = ?", models.ProcessExecutionStatusPending).Find(&executions).Error; err != nil {
		return nil, err
	}
	return executions, nil
}
