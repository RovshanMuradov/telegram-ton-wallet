// internal/db/db.go
package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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
		log.Printf("Database connection attempt %d/30: %v", i+1, err)
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to the database after 30 attempts: %w", err)
	}

	DB = sqlDB

	// Run migrations
	if err := runMigrations(databaseURL); err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	return nil
}

func CheckWalletsTableStructure() {
	var result []struct {
		ColumnName string
		DataType   string
	}
	if err := DB.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'wallets'").Scan(&result).Error; err != nil {
		log.Printf("Error getting wallets table structure: %v", err)
	} else {
		log.Printf("Wallets table structure:")
		for _, col := range result {
			log.Printf("Column: %s, Type: %s", col.ColumnName, col.DataType)
		}
	}
}

func CheckWalletsTableIndexes() {
	var result []struct {
		IndexName string
		IndexDef  string
	}
	if err := DB.Raw("SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'wallets'").Scan(&result).Error; err != nil {
		log.Printf("Error checking wallets table indexes: %v", err)
	} else {
		log.Printf("Wallets table indexes:")
		for _, idx := range result {
			log.Printf("Index name: %s, Definition: %s", idx.IndexName, idx.IndexDef)
		}
	}
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("error getting sql.DB: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

func runMigrations(databaseURL string) error {
	log.Println("Starting migrations")
	migrationsPath := "/app/migrations/migrations"
	log.Println("Migrations path:", migrationsPath)

	// List files in the directory
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		log.Printf("Error reading migrations directory: %v", err)
	} else {
		for _, file := range files {
			log.Println("Migration file:", file.Name())
		}
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("error initializing migrations: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
