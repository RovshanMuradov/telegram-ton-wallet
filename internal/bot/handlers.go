package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/wallet"
	"gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) registerHandlers() {
	b.telegramBot.Handle("/start", b.handleStart)
	b.telegramBot.Handle("/create_wallet", b.handleCreateWallet)
	b.telegramBot.Handle("/balance", b.handleBalance)
	b.telegramBot.Handle("/send", b.handleSend)
	b.telegramBot.Handle("/receive", b.handleReceive)
	b.telegramBot.Handle("/help", b.handleHelp)
	b.telegramBot.Handle("/history", b.handleHistory) // Новая команда
}

func (b *Bot) handleStart(m *telebot.Message) {
	b.telegramBot.Send(m.Sender, "Добро пожаловать в TON кошелек! Используйте /help для просмотра доступных команд.")
}

func (b *Bot) handleHelp(m *telebot.Message) {
	helpText := `/start - Начало работы с ботом
/create_wallet - Создание нового кошелька
/balance - Проверка баланса
/send - Отправка TON
/receive - Получение адреса для пополнения
/history - История транзакций
/help - Справка по командам`
	b.telegramBot.Send(m.Sender, helpText)
}

func (b *Bot) handleCreateWallet(m *telebot.Message) {
	userID := int(m.Sender.ID) // Преобразование int64 в int
	w, err := wallet.CreateWallet(userID, b.config)
	if err != nil {
		log.Printf("Ошибка при создании кошелька для пользователя %d: %v", userID, err)
		b.telegramBot.Send(m.Sender, fmt.Sprintf("Ошибка при создании кошелька: %v", err))
		return
	}

	b.telegramBot.Send(m.Sender, fmt.Sprintf("Ваш кошелек успешно создан!\nАдрес: %s", w.Address))
}

func (b *Bot) handleBalance(m *telebot.Message) {
	userID := int(m.Sender.ID)
	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		b.telegramBot.Send(m.Sender, "Кошелек не найден. Создайте его с помощью /create_wallet.")
		return
	}

	balance, err := wallet.GetBalance(w.Address, b.config)
	if err != nil {
		b.telegramBot.Send(m.Sender, fmt.Sprintf("Ошибка при получении баланса: %v", err))
		return
	}

	b.telegramBot.Send(m.Sender, fmt.Sprintf("Ваш баланс: %s TON", balance))
}

func (b *Bot) handleSend(m *telebot.Message) {
	b.telegramBot.Send(m.Sender, "Пожалуйста, введите адрес получателя и сумму через пробел (например: EQAbcdefghijklmnopqrstuvwxyz1234567890abcdefghij 1.5):")

	b.telegramBot.Handle(telebot.OnText, func(c *telebot.Message) {
		args := strings.Split(c.Text, " ")
		if len(args) != 2 {
			b.telegramBot.Send(c.Sender, "Неверный формат. Попробуйте снова.")
			return
		}

		recipientAddress := args[0]
		amount := args[1]

		if err := wallet.ValidateAddress(recipientAddress); err != nil {
			b.telegramBot.Send(c.Sender, fmt.Sprintf("Неверный адрес получателя: %v", err))
			return
		}

		if err := wallet.ValidateAmount(amount); err != nil {
			b.telegramBot.Send(c.Sender, fmt.Sprintf("Неверная сумма: %v", err))
			return
		}

		userID := int(c.Sender.ID)
		comment := "" // Здесь вы можете добавить логику для получения комментария, если это необходимо

		err := wallet.SendTON(userID, recipientAddress, amount, comment, b.config)
		if err != nil {
			b.telegramBot.Send(c.Sender, fmt.Sprintf("Ошибка при отправке транзакции: %v", err))
			return
		}

		b.telegramBot.Send(c.Sender, fmt.Sprintf("Транзакция успешно отправлена! Отправлено %s TON на адрес %s", amount, recipientAddress))
		b.registerHandlers() // Сбрасываем обработчики к исходному состоянию
	})
}

func (b *Bot) handleReceive(m *telebot.Message) {
	userID := int(m.Sender.ID)
	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		b.telegramBot.Send(m.Sender, "Кошелек не найден. Создайте его с помощью /create_wallet.")
		return
	}

	b.telegramBot.Send(m.Sender, fmt.Sprintf("Ваш адрес для пополнения:\n%s", w.Address))
}

func (b *Bot) handleHistory(m *telebot.Message) {
	userID := int(m.Sender.ID)
	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		b.telegramBot.Send(m.Sender, "Кошелек не найден. Создайте его с помощью /create_wallet.")
		return
	}

	transactions, err := wallet.GetTransactionHistory(w, b.config)
	if err != nil {
		b.telegramBot.Send(m.Sender, fmt.Sprintf("Ошибка при получении истории транзакций: %v", err))
		return
	}

	if len(transactions) == 0 {
		b.telegramBot.Send(m.Sender, "У вас пока нет транзакций.")
		return
	}

	historyText := "История ваших транзакций:\n\n"
	for _, tx := range transactions {
		historyText += fmt.Sprintf("ID: %d\nСумма: %s TON\nАдрес: %s\nДата: %s\n\n", tx.ID, tx.Amount, tx.ToAddress, tx.CreatedAt.Format("02.01.2006 15:04:05"))
	}

	b.telegramBot.Send(m.Sender, historyText)
}
