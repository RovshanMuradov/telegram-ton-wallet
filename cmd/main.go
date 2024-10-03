// cmd/main.go
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/bot"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/config"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/db"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"go.uber.org/zap"
)

func main() {
	// Initialize logging
	logging.InitLogger()
	logging.Info("Zap logger initialized and ready to use")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logging.Fatal("Error loading configuration", zap.Error(err))
	}

	// Add a small delay before initializing the database
	time.Sleep(time.Second * 5)

	// Initialize the database
	err = db.Init(cfg.DatabaseURL)
	if err != nil {
		logging.Fatal("Error connecting to the database", zap.Error(err))
	}
	defer db.Close()

	db.CheckWalletsTableStructure()
	db.CheckWalletsTableIndexes()

	// Create and start the bot
	b, err := bot.NewBot(cfg)
	if err != nil {
		logging.Fatal("Error creating bot", zap.Error(err))
	}

	// Launch the bot in a separate goroutine
	go b.Start()

	// Wait for termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// When receiving a signal, call Stop to terminate gracefully
	logging.Info("Shutting down...")
	b.Stop()
}
