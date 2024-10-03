// internal/bot/handler.go
package bot

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/wallet"
	"go.uber.org/zap"
	"gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) registerHandlers() {
	b.telegramBot.Handle("/start", b.handleStart)
	b.telegramBot.Handle("/create_wallet", b.handleCreateWallet)
	b.telegramBot.Handle("/balance", b.handleBalance)
	b.telegramBot.Handle("/send", b.handleSend)
	b.telegramBot.Handle("/receive", b.handleReceive)
	b.telegramBot.Handle("/help", b.handleHelp)
	b.telegramBot.Handle("/history", b.handleHistory)
	b.telegramBot.Handle("/backup", b.handleBackup)
	b.telegramBot.Handle("/restore", b.handleRestore)
	b.telegramBot.Handle(telebot.OnText, b.handleMessages)
	b.telegramBot.Handle(telebot.OnDocument, b.handleMessages)
}

func (b *Bot) handleStart(m *telebot.Message) {
	b.sendMessage(m.Sender, "Welcome to TON wallet! Use /help to view available commands.")
}

func (b *Bot) handleHelp(m *telebot.Message) {
	helpText := `/start - Start working with the bot
/create_wallet - Create a new wallet
/balance - Check balance
/send - Send TON
/receive - Get address for top-up
/history - Transaction history
/help - Command reference
/backup - Backup command
/restore - Restore command
`
	b.sendMessage(m.Sender, helpText)
}

func (b *Bot) handleCreateWallet(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	logging.Info("Create wallet request", zap.Int64("userID", userID))

	// Checking the existence of a wallet
	existingWallet, err := wallet.GetWalletByUserID(userID)
	if err == nil && existingWallet != nil {
		// Wallet already exists
		logging.Info("Wallet already exists", zap.Int64("userID", userID), zap.String("walletAddress", maskAddress(existingWallet.Address)))
		b.sendMessage(m.Sender, fmt.Sprintf("You already have a wallet!\nAddress: %s", existingWallet.Address))
		return
	}

	// Создаем новый кошелек
	w, err := wallet.CreateWallet(userID, b.config)
	if err != nil {
		logging.Error("Error creating wallet", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, fmt.Sprintf("Error creating wallet: %v", err))
		return
	}

	logging.Info("Wallet created successfully",
		zap.Int64("userID", userID),
		zap.String("walletAddress", maskAddress(w.Address)),
	)

	b.sendMessage(m.Sender, fmt.Sprintf("Your wallet has been successfully created!\nAddress: %s", w.Address))
}

func maskAddress(address string) string {
	if len(address) < 10 {
		return address
	}
	return address[:5] + "..." + address[len(address)-5:]
}

func (b *Bot) handleBalance(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	logging.Info("Balance request", zap.Int64("userID", userID))

	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		logging.Error("Error getting wallet", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Wallet not found. Create it using /create_wallet.")
		return
	}

	balance, err := wallet.GetBalance(w.Address, b.config)
	if err != nil {
		logging.Error("Error getting balance", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, fmt.Sprintf("Error getting balance: %v", err))
		return
	}

	logging.Info("Balance retrieved", zap.Int64("userID", userID), zap.String("balance", balance))
	b.sendMessage(m.Sender, fmt.Sprintf("Your balance: %s TON", balance))
}

func (b *Bot) handleSend(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	logging.Info("Send TON request", zap.Int64("userID", userID))

	b.stateMutex.Lock()
	b.userStates[userID] = userState{
		state:     "awaiting_send_details",
		timestamp: time.Now(),
	}
	b.stateMutex.Unlock()

	b.sendMessage(m.Sender, "Please enter the recipient's address and amount separated by a space (e.g., EQAbcdef... 1.5):")
}

func (b *Bot) handleReceive(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	logging.Info("Receive address request", zap.Int64("userID", userID))

	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		logging.Error("Error retrieving wallet", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Wallet not found. Create it using /create_wallet.")
		return
	}

	logging.Info("Sending wallet address", zap.Int64("userID", userID), zap.String("walletAddress", w.Address))
	b.sendMessage(m.Sender, fmt.Sprintf("Your address for top-up:\n%s", w.Address))
}

func (b *Bot) handleHistory(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	logging.Info("Transaction history request", zap.Int64("userID", userID))

	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		logging.Error("Error retrieving wallet", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Wallet not found. Create it using /create_wallet.")
		return
	}

	transactions, err := wallet.GetTransactionHistory(w, b.config)
	if err != nil {
		logging.Error("Error retrieving transaction history", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, fmt.Sprintf("Error getting transaction history: %v", err))
		return
	}

	if len(transactions) == 0 {
		logging.Info("No transactions found", zap.Int64("userID", userID))
		b.sendMessage(m.Sender, "You don't have any transactions yet.")
		return
	}

	// Формируем текст истории транзакций
	historyText := "Your transaction history:\n\n"
	for _, tx := range transactions {
		historyText += fmt.Sprintf("ID: %d\nAmount: %s TON\nAddress: %s\nDate: %s\n\n",
			tx.ID, tx.Amount, tx.ToAddress, tx.CreatedAt.Format("02.01.2006 15:04:05"))
	}

	logging.Info("Sending transaction history", zap.Int64("userID", userID), zap.Int("transactionCount", len(transactions)))
	b.sendMessage(m.Sender, historyText)
}

func (b *Bot) handleBackup(m *telebot.Message) {
	userID := int64(m.Sender.ID)

	backupData, err := wallet.CreateWalletBackup(userID, b.config)
	if err != nil {
		logging.Error("Error creating backup", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, fmt.Sprintf("Error creating backup: %v", err))
		return
	}

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "wallet_backup_*.bin")
	if err != nil {
		logging.Error("Error creating backup file", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Error creating backup file")
		return
	}
	defer os.Remove(tmpfile.Name())

	// Write the backup data to the file
	if _, err := tmpfile.Write(backupData); err != nil {
		logging.Error("Error writing backup data", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Error writing backup data")
		return
	}

	// Close the file before sending
	if err := tmpfile.Close(); err != nil {
		logging.Error("Error closing backup file", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Error closing backup file")
		return
	}

	// Send the file
	doc := &telebot.Document{File: telebot.FromDisk(tmpfile.Name())}
	if _, err := b.telegramBot.Send(m.Sender, doc, "Here's your wallet backup. Keep it safe!"); err != nil {
		logging.Error("Error sending backup file", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Error sending backup file. Please try again later.")
		return
	}
}

func (b *Bot) hasStateExpired(userState userState) bool {
	return time.Since(userState.timestamp) > 5*time.Minute
}

func (b *Bot) handleRestore(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	logging.Info("Restore wallet request", zap.Int64("userID", userID))

	// Устанавливаем состояние пользователя с блокировкой мьютекса
	b.stateMutex.Lock()
	b.userStates[userID] = userState{
		state:     "awaiting_backup_file",
		timestamp: time.Now(),
	}
	b.stateMutex.Unlock()

	b.sendMessage(m.Sender, "Please send me the backup file.")
}

func (b *Bot) handleMessages(m *telebot.Message) {
	userID := int64(m.Sender.ID)

	b.stateMutex.RLock()
	userState, exists := b.userStates[userID]
	b.stateMutex.RUnlock()

	if !exists {
		// Если состояние не установлено, игнорируем сообщение или обрабатываем как обычную команду
		return
	}

	// Проверяем тайм-аут
	if b.hasStateExpired(userState) {
		// Сбрасываем состояние
		b.stateMutex.Lock()
		delete(b.userStates, userID)
		b.stateMutex.Unlock()
		b.sendMessage(m.Sender, "Session timed out. Please start again.")
		return
	}

	// Обрабатываем сообщение в соответствии с состоянием
	switch userState.state {
	case "awaiting_send_details":
		b.processSendDetails(m)
	case "awaiting_backup_file":
		if m.Document != nil {
			b.processBackupFile(m)
		} else {
			b.sendMessage(m.Sender, "Please send a valid backup file.")
		}
	// Можно добавить другие состояния по мере необходимости
	default:
		b.sendMessage(m.Sender, "I didn't understand that command. Use /help to see available commands.")
	}
}

func (b *Bot) processSendDetails(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	args := strings.Split(m.Text, " ")
	if len(args) != 2 {
		b.sendMessage(m.Sender, "Invalid format. Please try again.")
		return
	}

	recipientAddress := args[0]
	amount := args[1]

	if err := wallet.ValidateAddress(recipientAddress); err != nil {
		logging.Error("Invalid recipient address", zap.String("address", recipientAddress), zap.Error(err))
		b.sendMessage(m.Sender, fmt.Sprintf("Invalid recipient address: %v", err))
		return
	}

	if err := wallet.ValidateAmount(amount); err != nil {
		logging.Error("Invalid amount", zap.String("amount", amount), zap.Error(err))
		b.sendMessage(m.Sender, fmt.Sprintf("Invalid amount: %v", err))
		return
	}

	comment := ""

	err := wallet.SendTON(userID, recipientAddress, amount, comment, b.config)
	if err != nil {
		logging.Error("Error sending transaction", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, fmt.Sprintf("Error sending transaction: %v", err))
		return
	}

	b.sendMessage(m.Sender, fmt.Sprintf("Transaction sent successfully! Sent %s TON to address %s", amount, recipientAddress))

	// Сбрасываем состояние пользователя
	b.stateMutex.Lock()
	delete(b.userStates, userID)
	b.stateMutex.Unlock()
}

func (b *Bot) processBackupFile(m *telebot.Message) {
	userID := int64(m.Sender.ID)

	// Загрузка файла
	fileInfo, err := b.telegramBot.GetFile(m.Document.MediaFile())
	if err != nil {
		logging.Error("Error accessing backup file", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Error accessing backup file")
		return
	}

	var buf bytes.Buffer

	// Копируем содержимое файла в буфер
	if _, err := io.Copy(&buf, fileInfo); err != nil {
		logging.Error("Error reading backup file", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, "Error reading backup file")
		return
	}

	backupData := buf.Bytes()

	// Восстанавливаем кошелек
	err = wallet.RestoreWalletFromBackup(userID, backupData, b.config)
	if err != nil {
		logging.Error("Error restoring wallet", zap.Int64("userID", userID), zap.Error(err))
		b.sendMessage(m.Sender, fmt.Sprintf("Error restoring wallet: %v", err))
		return
	}

	b.sendMessage(m.Sender, "Wallet successfully restored!")

	// Сбрасываем состояние пользователя
	b.stateMutex.Lock()
	delete(b.userStates, userID)
	b.stateMutex.Unlock()
}
