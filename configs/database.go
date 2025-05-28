package configs

import (
	"bbb/internal/models"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetUpDatabaseConnection(env Env) *gorm.DB {

	db, err := gorm.Open(postgres.Open(env.DSN), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		panic("Failed to connect database")
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.UserGroups{},
		&models.Process{},
		&models.ProcessBuilder{},
		&models.Notification{},
		&models.Task{},
		&models.TaskPrerequisite{},
		&models.TaskDependency{},
		&models.ProcessExecution{},
		&models.TaskExecution{},
		&models.PendingTask{},
		&models.CompletedTask{},
		&models.InProgressTask{},
	)
	if err != nil {
		fmt.Println(err)
		panic("AutoMigrate failed")
	}

	return db
}
