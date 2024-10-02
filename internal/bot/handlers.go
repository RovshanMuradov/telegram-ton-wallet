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
	b.telegramBot.Handle("/history", b.handleHistory)
}

func (b *Bot) handleStart(m *telebot.Message) {
	b.telegramBot.Send(m.Sender, "Welcome to TON wallet! Use /help to view available commands.")
}

func (b *Bot) handleHelp(m *telebot.Message) {
	helpText := `/start - Start working with the bot
/create_wallet - Create a new wallet
/balance - Check balance
/send - Send TON
/receive - Get address for top-up
/history - Transaction history
/help - Command reference`
	b.telegramBot.Send(m.Sender, helpText)
}

func (b *Bot) handleCreateWallet(m *telebot.Message) {
	userID := int64(m.Sender.ID)

	// Checking the existence of a wallet
	existingWallet, err := wallet.GetWalletByUserID(userID)
	if err == nil && existingWallet != nil {
		// Wallet already exists
		b.telegramBot.Send(m.Sender, fmt.Sprintf("You already have a wallet!\nAddress: %s", existingWallet.Address))
		return
	}

	w, err := wallet.CreateWallet(userID, b.config)
	if err != nil {
		log.Printf("Error creating wallet for user %d: %v", userID, err)
		b.telegramBot.Send(m.Sender, fmt.Sprintf("Error creating wallet: %v", err))
		return
	}

	b.telegramBot.Send(m.Sender, fmt.Sprintf("Your wallet has been successfully created!\nAddress: %s", w.Address))
}

func (b *Bot) handleBalance(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	log.Printf("Balance request for user %d", userID)

	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		log.Printf("Error getting wallet for user %d: %v", userID, err)
		b.telegramBot.Send(m.Sender, "Wallet not found. Create it using /create_wallet.")
		return
	}

	balance, err := wallet.GetBalance(w.Address, b.config)
	if err != nil {
		b.telegramBot.Send(m.Sender, fmt.Sprintf("Error getting balance: %v", err))
		return
	}

	b.telegramBot.Send(m.Sender, fmt.Sprintf("Your balance: %s TON", balance))
}

func (b *Bot) handleSend(m *telebot.Message) {
	b.telegramBot.Send(m.Sender, "Please enter the recipient's address and amount separated by a space (e.g., EQAbcdefghijklmnopqrstuvwxyz1234567890abcdefghij 1.5):")

	b.telegramBot.Handle(telebot.OnText, func(c *telebot.Message) {
		args := strings.Split(c.Text, " ")
		if len(args) != 2 {
			b.telegramBot.Send(c.Sender, "Invalid format. Please try again.")
			return
		}

		recipientAddress := args[0]
		amount := args[1]

		if err := wallet.ValidateAddress(recipientAddress); err != nil {
			b.telegramBot.Send(c.Sender, fmt.Sprintf("Invalid recipient address: %v", err))
			return
		}

		if err := wallet.ValidateAmount(amount); err != nil {
			b.telegramBot.Send(c.Sender, fmt.Sprintf("Invalid amount: %v", err))
			return
		}

		userID := int64(c.Sender.ID)
		comment := ""

		err := wallet.SendTON(userID, recipientAddress, amount, comment, b.config)
		if err != nil {
			b.telegramBot.Send(c.Sender, fmt.Sprintf("Error sending transaction: %v", err))
			return
		}

		b.telegramBot.Send(c.Sender, fmt.Sprintf("Transaction sent successfully! Sent %s TON to address %s", amount, recipientAddress))
		b.registerHandlers()
	})
}

func (b *Bot) handleReceive(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		b.telegramBot.Send(m.Sender, "Wallet not found. Create it using /create_wallet.")
		return
	}

	b.telegramBot.Send(m.Sender, fmt.Sprintf("Your address for top-up:\n%s", w.Address))
}

func (b *Bot) handleHistory(m *telebot.Message) {
	userID := int64(m.Sender.ID)
	w, err := wallet.GetWalletByUserID(userID)
	if err != nil {
		b.telegramBot.Send(m.Sender, "Wallet not found. Create it using /create_wallet.")
		return
	}

	transactions, err := wallet.GetTransactionHistory(w, b.config)
	if err != nil {
		b.telegramBot.Send(m.Sender, fmt.Sprintf("Error getting transaction history: %v", err))
		return
	}

	if len(transactions) == 0 {
		b.telegramBot.Send(m.Sender, "You don't have any transactions yet.")
		return
	}

	historyText := "Your transaction history:\n\n"
	for _, tx := range transactions {
		historyText += fmt.Sprintf("ID: %d\nAmount: %s TON\nAddress: %s\nDate: %s\n\n", tx.ID, tx.Amount, tx.ToAddress, tx.CreatedAt.Format("02.01.2006 15:04:05"))
	}

	b.telegramBot.Send(m.Sender, historyText)
}
