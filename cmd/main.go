// cmd/main.go
package main

import (
	"log"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/bot"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/config"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/db"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
)

func main() {
	// Инициализация логирования
	logging.Init()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация базы данных
	err = db.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Создание и запуск бота
	b, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	b.Start()
}
