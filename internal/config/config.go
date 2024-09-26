// internal/config/config.go
package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	TelegramToken string
	TonAPIKey     string
	DatabaseURL   string
	EncryptionKey string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		TelegramToken: viper.GetString("TELEGRAM_TOKEN"),
		TonAPIKey:     viper.GetString("TON_API_KEY"),
		DatabaseURL:   viper.GetString("DATABASE_URL"),
		EncryptionKey: viper.GetString("ENCRYPTION_KEY"),
	}

	return config, nil
}
