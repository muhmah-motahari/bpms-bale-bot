package models

import "time"

type (
	Process struct {
		ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
		Name        string    `gorm:"type:varchar(100);not null" json:"name"`
		Description string    `gorm:"type:text" json:"description"`
		UserID      int64     `gorm:"type:bigint;index" json:"user_id"`
		Tasks       []Task    `json:"tasks"`
		CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
		UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	}

	ProcessExecutionStatus string

	ProcessExecution struct {
		ID                         uint                   `gorm:"primaryKey;autoIncrement" json:"id"`
		ProcessID                  uint                   `gorm:"index" json:"process_id"`
		Process                    *Process               `json:"process"`
		Status                     ProcessExecutionStatus `gorm:"type:varchar(50);default:'pending'" json:"status"`
		PendingTaskExecutionIDs    []uint                 `gorm:"-" json:"pending_task_execution_ids"`
		CompletedTaskExecutionIDs  []uint                 `gorm:"-" json:"completed_task_execution_ids"`
		InProgressTaskExecutionIDs []uint                 `gorm:"-" json:"in_progress_task_execution_ids"`
		StartedAt                  time.Time              `gorm:"autoCreateTime" json:"started_at"`
		CompletedAt                *time.Time             `json:"completed_at"`
		CreatedAt                  time.Time              `gorm:"autoCreateTime" json:"created_at"`
		UpdatedAt                  time.Time              `gorm:"autoUpdateTime" json:"updated_at"`
	}

	// PendingTask represents a pending task execution in a process execution
	PendingTask struct {
		ProcessExecutionID uint      `gorm:"primaryKey;index" json:"process_execution_id"`
		TaskExecutionID    uint      `gorm:"primaryKey;index" json:"task_execution_id"`
		CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	}

	// CompletedTask represents a completed task execution in a process execution
	CompletedTask struct {
		ProcessExecutionID uint      `gorm:"primaryKey;index" json:"process_execution_id"`
		TaskExecutionID    uint      `gorm:"primaryKey;index" json:"task_execution_id"`
		CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	}

	// InProgressTask represents a task execution that is assigned but not completed
	InProgressTask struct {
		ProcessExecutionID uint      `gorm:"primaryKey;index" json:"process_execution_id"`
		TaskExecutionID    uint      `gorm:"primaryKey;index" json:"task_execution_id"`
		CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	}

	ProcessBuilder struct {
		ID          uint
		UserID      int64
		User        *User
		CurrentStep string // "name", "description", etc.
		ProcessID   uint
		Process     *Process
	}
)

const (
	ProcessExecutionStatusPending   ProcessExecutionStatus = "pending"
	ProcessExecutionStatusRunning   ProcessExecutionStatus = "running"
	ProcessExecutionStatusCompleted ProcessExecutionStatus = "completed"
	ProcessExecutionStatusFailed    ProcessExecutionStatus = "failed"
)
