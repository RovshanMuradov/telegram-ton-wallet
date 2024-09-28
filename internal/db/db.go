// internal/db/db.go
package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(databaseURL string) error {
	var err error
	var sqlDB *gorm.DB

	for i := 0; i < 30; i++ {
		sqlDB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Попытка подключения к базе данных %d/30: %v", i+1, err)
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных после 30 попыток: %w", err)
	}

	DB = sqlDB
	return nil
}

func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		sqlDB.Close()
	}
}
