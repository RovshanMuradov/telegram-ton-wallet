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
	// Initialize logging
	logging.Init()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Add a small delay before initializing the database
	time.Sleep(time.Second * 5)

	// Initialize the database
	err = db.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	db.CheckWalletsTableStructure()
	db.CheckWalletsTableIndexes()

	// Create and start the bot
	b, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	b.Start()

	// Wait for termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")
}
