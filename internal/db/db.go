// internal/db/db.go
package db

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(databaseURL string) error {
	var err error
	var sqlDB *gorm.DB

	// Try to connect to the database up to 30 times
	for i := 0; i < 30; i++ {
		sqlDB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err == nil {
			break
		}
		// Log each connection attempt
		logging.Warn("Database connection attempt failed", zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(time.Second * 2)
	}

	// If after 30 attempts connection failed, log the error and return it
	if err != nil {
		logging.Error("Failed to connect to the database after 30 attempts", zap.Error(err))
		return fmt.Errorf("failed to connect to the database after 30 attempts: %w", err)
	}

	DB = sqlDB

	// Run migrations
	if err := runMigrations(databaseURL); err != nil {
		logging.Error("Error running migrations", zap.Error(err))
		return fmt.Errorf("error running migrations: %w", err)
	}

	return nil
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			logging.Error("Error getting sql.DB instance", zap.Error(err))
			return fmt.Errorf("error getting sql.DB: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			logging.Error("Error closing database connection", zap.Error(err))
			return fmt.Errorf("error closing database connection: %w", err)
		}
		logging.Info("Database connection closed successfully")
	}
	return nil
}

func runMigrations(databaseURL string) error {
	logging.Info("Starting migrations")
	migrationsPath := "/app/migrations/migrations"
	logging.Info("Migrations path", zap.String("path", migrationsPath))

	// List migration files in the directory
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		logging.Error("Error reading migrations directory", zap.Error(err))
		return fmt.Errorf("error reading migrations directory: %w", err)
	}

	for _, file := range files {
		logging.Info("Found migration file", zap.String("file", file.Name()))
	}

	// Initialize migrations
	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		logging.Error("Error initializing migrations", zap.Error(err))
		return fmt.Errorf("error initializing migrations: %w", err)
	}

	// Run migrations and handle the "no change" case
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logging.Error("Error running migrations", zap.Error(err))
		return fmt.Errorf("error running migrations: %w", err)
	}

	logging.Info("Migrations completed successfully")
	return nil
}
