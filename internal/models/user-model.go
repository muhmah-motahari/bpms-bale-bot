package models

import (
	"time"

	"gorm.io/gorm"
)

type (
	User struct {
		ID        int64     `gorm:"primaryKey;type:bigint" json:"id"`
		Username  string    `gorm:"type:varchar(100)" json:"username"`
		FirstName string    `gorm:"type:varchar(100);column:firstname" json:"firstname"`
		LastName  string    `gorm:"type:varchar(100);column:lastname" json:"lastname"`
		Groups    []Group   `gorm:"many2many:user_groups;" json:"groups"`
		CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
		UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	}

	UserGroups struct {
		UserID    int64 `gorm:"primaryKey"`
		GroupID   uint  `gorm:"primaryKey"`
		CreatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
)
