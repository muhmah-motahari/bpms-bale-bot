package models

import "time"

// Notification represents a message sent to a user
type Notification struct {
	ID      uint      `gorm:"primaryKey;autoIncrement" json:"notification_id"`
	TaskID  uint      `gorm:"index" json:"task_id"`             // References Task
	UserID  int64     `gorm:"type:bigint;index" json:"user_id"` // References User
	Message string    `gorm:"type:text;not null" json:"message"`
	Status  string    `gorm:"type:varchar(50);default:'sent'" json:"status"` // sent, read
	SentAt  time.Time `gorm:"autoCreateTime" json:"sent_at"`
}
