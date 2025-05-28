package models

import "time"

type (
	Group struct {
		ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
		Name        string    `gorm:"type:varchar(100);not null" json:"name"`
		Description string    `gorm:"type:text" json:"description"`
		JoinKey     string    `gorm:"type:varchar(8);unique;not null" json:"join_key"`
		Users       []User    `gorm:"many2many:user_groups;" json:"users"`
		OwnerID     int64     `gorm:"not null" json:"owner_id"`
		CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
		UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	}
)
