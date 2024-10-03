// internal/bot/handler.go
package bot

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

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
		logging.Info("Wallet already exists", zap.Int64("userID", userID), zap.String("walletAddress", existingWallet.Address))
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

	logging.Info("Wallet created successfully", zap.Int64("userID", userID), zap.String("walletAddress", w.Address))
	b.sendMessage(m.Sender, fmt.Sprintf("Your wallet has been successfully created!\nAddress: %s", w.Address))
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
	b.sendMessage(m.Sender, "Please enter the recipient's address and amount separated by a space (e.g., EQAbcdefghijklmnopqrstuvwxyz1234567890abcdefghij 1.5):")

	b.telegramBot.Handle(telebot.OnText, func(c *telebot.Message) {
		args := strings.Split(c.Text, " ")
		if len(args) != 2 {
			b.sendMessage(c.Sender, "Invalid format. Please try again.")
			return
		}

		recipientAddress := args[0]
		amount := args[1]

		if err := wallet.ValidateAddress(recipientAddress); err != nil {
			logging.Error("Invalid recipient address", zap.String("address", recipientAddress), zap.Error(err))
			b.sendMessage(c.Sender, fmt.Sprintf("Invalid recipient address: %v", err))
			return
		}

		if err := wallet.ValidateAmount(amount); err != nil {
			logging.Error("Invalid amount", zap.String("amount", amount), zap.Error(err))
			b.sendMessage(c.Sender, fmt.Sprintf("Invalid amount: %v", err))
			return
		}

		userID := int64(c.Sender.ID)
		comment := ""

		err := wallet.SendTON(userID, recipientAddress, amount, comment, b.config)
		if err != nil {
			logging.Error("Error sending transaction", zap.Int64("userID", userID), zap.Error(err))
			b.sendMessage(c.Sender, fmt.Sprintf("Error sending transaction: %v", err))
			return
		}

		b.sendMessage(c.Sender, fmt.Sprintf("Transaction sent successfully! Sent %s TON to address %s", amount, recipientAddress))

		b.registerHandlers()
	})
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

func (b *Bot) handleRestore(m *telebot.Message) {
	b.sendMessage(m.Sender, "Please send me the backup file.")

	b.telegramBot.Handle(telebot.OnDocument, func(m *telebot.Message) {
		userID := int64(m.Sender.ID)

		// Download the file
		fileInfo, err := b.telegramBot.GetFile(m.Document.MediaFile())
		if err != nil {
			logging.Error("Error accessing backup file", zap.Int64("userID", userID), zap.Error(err))
			b.sendMessage(m.Sender, "Error accessing backup file")
			return
		}

		// Create a buffer to store the file contents
		var buf bytes.Buffer

		// Download the file content into the buffer
		if _, err := io.Copy(&buf, fileInfo); err != nil {
			logging.Error("Error reading backup file", zap.Int64("userID", userID), zap.Error(err))
			b.sendMessage(m.Sender, "Error reading backup file")
			return
		}

		// Get the backup data from the buffer
		backupData := buf.Bytes()

		// Restore the wallet
		err = wallet.RestoreWalletFromBackup(userID, backupData, b.config)
		if err != nil {
			logging.Error("Error restoring wallet", zap.Int64("userID", userID), zap.Error(err))
			b.sendMessage(m.Sender, fmt.Sprintf("Error restoring wallet: %v", err))
			return
		}

		b.sendMessage(m.Sender, "Wallet successfully restored!")
		b.registerHandlers() // Re-register handlers to stop listening for documents
	})
}
