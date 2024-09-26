package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

func Init(databaseURL string) error {
	db, err := gorm.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	DB = db
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
