package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"type:varchar(100);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	ProcessID   uint      `gorm:"index" json:"process_id"`
	Process     *Process  `json:"process"`
	TeamID      *uint     `gorm:"index" json:"team_id"`
	Team        *Team     `json:"team"`
	IsFinal     bool      `gorm:"default:false" json:"is_final"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type TaskExecution struct {
	ID                 uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID             uint       `gorm:"index" json:"task_id"`
	Task               *Task      `json:"task"`
	ProcessExecutionID uint       `gorm:"index" json:"process_execution_id"`
	Status             TaskStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
	UserID             *int64     `gorm:"type:bigint;index" json:"user_id"`
	User               *User      `json:"user"`
	AssignedAt         *time.Time `json:"assigned_at"`
	UserDescription    string     `json:"user_description"`
	CompletedAt        *time.Time `json:"completed_at"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TaskPrerequisite represents prerequisite tasks that must be completed before a task can start
type TaskPrerequisite struct {
	TaskID         uint      `gorm:"primaryKey;index" json:"task_id"`         // References Task
	PrerequisiteID uint      `gorm:"primaryKey;index" json:"prerequisite_id"` // References Task
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TaskDependency represents the next tasks that can be started after completing a task
type TaskDependency struct {
	TaskID     uint      `gorm:"primaryKey;index" json:"task_id"`      // References Task
	NextTaskID uint      `gorm:"primaryKey;index" json:"next_task_id"` // References Task
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusAssigned  TaskStatus = "assigned"
	TaskStatusCompleted TaskStatus = "completed"
)

// TaskBuilder manages the state of task creation
type TaskBuilder struct {
	UserID               int64
	CurrentStep          string // "process", "title", "description", "prerequisites", "team"
	ProcessID            uint
	Task                 Task   `gorm:"-"` // GORM will ignore this field
	Prerequisites        []uint // List of prerequisite task IDs
	HasMorePrerequisites bool
}

// BeforeCreate hook for TaskPrerequisite to prevent self-referencing prerequisites
func (td *TaskPrerequisite) BeforeCreate(tx *gorm.DB) (err error) {
	if td.TaskID == td.PrerequisiteID {
		return errors.New("TaskPrerequisite: a task cannot be its own prerequisite")
	}
	return nil
}

// BeforeCreate hook for TaskDependency to prevent self-referencing dependencies
func (td *TaskDependency) BeforeCreate(tx *gorm.DB) (err error) {
	if td.TaskID == td.NextTaskID {
		return errors.New("TaskDependency: a task cannot be its own dependency")
	}
	return nil
}
