// internal/bot/bot.go
package bot

import (
	"sync"
	"time"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/config"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"go.uber.org/zap"
	"gopkg.in/tucnak/telebot.v2"
)

type userState struct {
	state     string
	timestamp time.Time
}

type Bot struct {
	telegramBot *telebot.Bot
	config      *config.Config
	stateMutex  sync.RWMutex
	userStates  map[int64]userState
	stopChan    chan struct{} // Channel to stop the bot
}

func NewBot(cfg *config.Config) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	return &Bot{
		telegramBot: b,
		config:      cfg,
		userStates:  make(map[int64]userState),
		stopChan:    make(chan struct{}),
	}, nil
}

// Start starts the bot and registers handlers
func (b *Bot) Start() {
	b.registerHandlers()
	logging.Info("The bot has been launched")

	go b.telegramBot.Start()

	// Wait for a signal to stop the bot
	<-b.stopChan
	b.telegramBot.Stop() // Stop the bot
	logging.Info("The bot has been stopped")
}

// Stop signals the end of the bot's work
func (b *Bot) Stop() {
	close(b.stopChan) // Close the channel to complete the work
}

// sendMessage sends a message to the user and logs an error if one occurs
func (b *Bot) sendMessage(m *telebot.User, message string) {
	_, err := b.telegramBot.Send(m, message)
	if err != nil {
		// Log the error of sending the message
		logging.Error("Error sending message",
			zap.String("message", message),
			zap.Error(err),
		)
	}
}
