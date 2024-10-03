// internal/bot/bot.go
package bot

import (
	"log"
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
	stateMutex  sync.RWMutex //need to add mutex
	userStates  map[int64]userState
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
	}, nil
}

func (b *Bot) Start() {
	b.registerHandlers()
	log.Println("The bot has been launched")
	b.telegramBot.Start()

}

// sendMessage отправляет сообщение пользователю и логирует ошибку, если она возникает
func (b *Bot) sendMessage(m *telebot.User, message string) {
	_, err := b.telegramBot.Send(m, message)
	if err != nil {
		// Логируем ошибку отправки сообщения
		logging.Error("Error sending message",
			zap.String("message", message),
			zap.Error(err),
		)
	}
}
