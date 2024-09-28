// internal/db/models.go
package db

import "time"

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
	ID        int `gorm:"primary_key"`
	WalletID  int
	Amount    string
	ToAddress string
	CreatedAt time.Time
}
