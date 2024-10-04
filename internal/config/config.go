// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"go.uber.org/zap"
)

type Config struct {
	TelegramToken string
	TonAPIKey     string
	DatabaseURL   string
	EncryptionKey string
	TonConfigURL  string
}

func LoadConfig() (*Config, error) {
	// Load environment variables from .env file (this is mandatory)
	err := godotenv.Load()
	if err != nil {
		// Log the error and return it if the .env file is missing
		logging.Error(".env file is required but was not found", zap.Error(err))
		return nil, fmt.Errorf(".env file is required but was not found: %v", err)
	}

	// Load configuration from environment variables
	config := &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		TonAPIKey:     os.Getenv("TON_API_KEY"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		TonConfigURL:  os.Getenv("TON_CONFIG_URL"),
	}

	// Check if essential configuration (like TON_CONFIG_URL) is set
	if config.TonConfigURL == "" {
		// Return an error if it's missing, this is critical
		logging.Error("TON_CONFIG_URL is not set")
		return nil, fmt.Errorf("TON_CONFIG_URL is not set")
	}

	// Return the loaded configuration
	return config, nil
}
