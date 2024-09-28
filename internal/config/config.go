// internal/config/config.go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	TonAPIKey     string
	DatabaseURL   string
	EncryptionKey string
	TonConfigURL  string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found in current directory")
	}

	config := &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		TonAPIKey:     os.Getenv("TON_API_KEY"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		TonConfigURL:  os.Getenv("TON_CONFIG_URL"),
	}

	// Проверка обязательных переменных
	if config.TonConfigURL == "" {
		return nil, fmt.Errorf("TON_CONFIG_URL is not set")
	}

	return config, nil
}
