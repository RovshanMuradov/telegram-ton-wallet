package bot

import (
	"log"
	"time"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/config"
	"gopkg.in/tucnak/telebot.v2"
)

type Bot struct {
	telegramBot *telebot.Bot
	config      *config.Config
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
	}, nil
}

func (b *Bot) Start() {
	b.registerHandlers()
	log.Println("Бот запущен")
	b.telegramBot.Start()
}
