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

	ProcessBuilder struct {
		ID          uint
		UserID      int64
		User        *User
		CurrentStep string // "name", "description", etc.
		ProcessID   uint
		Process     *Process
	}
)
