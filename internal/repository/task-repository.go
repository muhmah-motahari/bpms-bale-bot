package repository

import (
	"bbb/internal/models"

	"gorm.io/gorm"
)

type (
	TaskRepository interface {
		Save(req *models.Task) error
		GetByProcessID(processID uint) ([]models.Task, error)
		GetByID(taskID uint) (*models.Task, error)
		GetPrerequisites(taskID uint) ([]uint, error)
		AddPrerequisite(taskID uint, prerequisiteID uint) error
		StartTaskExecution(taskID uint) error
		GetTaskExecutionsByTaskID(taskID uint) ([]models.TaskExecution, error)
		GetTaskExecutionByID(taskExecutionID uint) (*models.TaskExecution, error)
		GetTaskExecutionsByUserID(userID int64) ([]models.TaskExecution, error)
		UpdateTaskExecution(taskExecution *models.TaskExecution) error
	}

	taskRepository struct {
		db *gorm.DB
	}
)

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) Save(req *models.Task) error {
	if req.ID == 0 {
		if err := r.db.Create(req).Error; err != nil {
			return err
		}
	} else {
		if err := r.db.Save(req).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *taskRepository) GetByProcessID(processID uint) ([]models.Task, error) {
	var tasks []models.Task
	if err := r.db.Where("process_id = ?", processID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) GetByID(taskID uint) (*models.Task, error) {
	var task models.Task
	if err := r.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) GetPrerequisites(taskID uint) ([]uint, error) {
	var prerequisites []models.TaskPrerequisite
	if err := r.db.Where("task_id = ?", taskID).Find(&prerequisites).Error; err != nil {
		return nil, err
	}

	prerequisiteIDs := make([]uint, len(prerequisites))
	for i, prerequisite := range prerequisites {
		prerequisiteIDs[i] = prerequisite.PrerequisiteID
	}
	return prerequisiteIDs, nil
}

func (r *taskRepository) AddPrerequisite(taskID uint, prerequisiteID uint) error {
	prerequisite := models.TaskPrerequisite{
		TaskID:         taskID,
		PrerequisiteID: prerequisiteID,
	}
	return r.db.Create(&prerequisite).Error
}

func (r *taskRepository) StartTaskExecution(taskID uint) error {
	return r.db.Create(&models.TaskExecution{
		TaskID: taskID,
		Status: models.TaskStatusPending,
	}).Error
}

func (r *taskRepository) GetTaskExecutionsByTaskID(taskID uint) ([]models.TaskExecution, error) {
	var taskExecutions []models.TaskExecution
	if err := r.db.Where("task_id = ?", taskID).Find(&taskExecutions).Error; err != nil {
		return nil, err
	}
	return taskExecutions, nil
}

func (r *taskRepository) GetTaskExecutionByID(taskExecutionID uint) (*models.TaskExecution, error) {
	var taskExecution models.TaskExecution
	if err := r.db.First(&taskExecution, taskExecutionID).Error; err != nil {
		return nil, err
	}
	return &taskExecution, nil
}

func (r *taskRepository) GetTaskExecutionsByUserID(userID int64) ([]models.TaskExecution, error) {
	var taskExecutions []models.TaskExecution
	if err := r.db.Where("user_id = ?", userID).Find(&taskExecutions).Error; err != nil {
		return nil, err
	}
	return taskExecutions, nil
}

func (r *taskRepository) UpdateTaskExecution(taskExecution *models.TaskExecution) error {
	return r.db.Model(&models.TaskExecution{}).Where("id = ?", taskExecution.ID).Updates(taskExecution).Error
}
