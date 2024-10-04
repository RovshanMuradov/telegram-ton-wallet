// internal/db/models.go
package db

import (
	"time"
)

type User struct {
	ID         int64 `gorm:"primary_key"`
	TelegramID int64
	Wallets    []Wallet
}

type Wallet struct {
	ID         int64 `gorm:"primary_key"`
	UserID     int64
	Address    string
	PrivateKey string
	Balance    string
	Locked     bool
	LockedAt   time.Time
}

type Transaction struct {
	ID        int   `gorm:"primary_key"`
	WalletID  int64 // Изменено на int64, чтобы соответствовать типу ID в Wallet
	Amount    string
	ToAddress string
	TxHash    string // Добавлено поле для хранения хеша транзакции
	CreatedAt time.Time
}

type TransactionStatus struct {
	Hash        string
	BlockID     string
	Type        string
	Status      string
	LT          uint64
	Time        time.Time
	FromAddress string
	ToAddresses []string
}
