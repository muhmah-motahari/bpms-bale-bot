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

	// Save pending task executions
	for _, taskExecutionID := range execution.PendingTaskExecutionIDs {
		pendingTask := models.PendingTask{
			ProcessExecutionID: execution.ID,
			TaskExecutionID:    taskExecutionID,
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

	// Get completed tasks
	var completedTasks []models.CompletedTask
	if err := r.db.Where("process_execution_id = ?", id).Find(&completedTasks).Error; err != nil {
		return nil, err
	}

	// Get in-progress tasks
	var inProgressTasks []models.InProgressTask
	if err := r.db.Where("process_execution_id = ?", id).Find(&inProgressTasks).Error; err != nil {
		return nil, err
	}

	// Convert to task execution IDs
	execution.PendingTaskExecutionIDs = make([]uint, len(pendingTasks))
	for i, pt := range pendingTasks {
		execution.PendingTaskExecutionIDs[i] = pt.TaskExecutionID
	}

	execution.CompletedTaskExecutionIDs = make([]uint, len(completedTasks))
	for i, ct := range completedTasks {
		execution.CompletedTaskExecutionIDs[i] = ct.TaskExecutionID
	}

	execution.InProgressTaskExecutionIDs = make([]uint, len(inProgressTasks))
	for i, it := range inProgressTasks {
		execution.InProgressTaskExecutionIDs[i] = it.TaskExecutionID
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

	// Delete existing task states
	if err := tx.Where("process_execution_id = ?", execution.ID).Delete(&models.PendingTask{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("process_execution_id = ?", execution.ID).Delete(&models.CompletedTask{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("process_execution_id = ?", execution.ID).Delete(&models.InProgressTask{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Save new pending task executions
	for _, taskExecutionID := range execution.PendingTaskExecutionIDs {
		pendingTask := models.PendingTask{
			ProcessExecutionID: execution.ID,
			TaskExecutionID:    taskExecutionID,
		}
		if err := tx.Create(&pendingTask).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Save completed task executions
	for _, taskExecutionID := range execution.CompletedTaskExecutionIDs {
		completedTask := models.CompletedTask{
			ProcessExecutionID: execution.ID,
			TaskExecutionID:    taskExecutionID,
		}
		if err := tx.Create(&completedTask).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Save in-progress task executions
	for _, taskExecutionID := range execution.InProgressTaskExecutionIDs {
		inProgressTask := models.InProgressTask{
			ProcessExecutionID: execution.ID,
			TaskExecutionID:    taskExecutionID,
		}
		if err := tx.Create(&inProgressTask).Error; err != nil {
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
