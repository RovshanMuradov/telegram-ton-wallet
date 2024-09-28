// cmd/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Добавляем небольшую задержку перед инициализацией базы данных
	time.Sleep(time.Second * 5)

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

	// Ожидание сигнала для завершения
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Завершение работы...")
}
