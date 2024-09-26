package db

import "time"

type User struct {
	ID         int `gorm:"primary_key"`
	TelegramID int
	Wallets    []Wallet
}

type Wallet struct {
	ID         int `gorm:"primary_key"`
	UserID     int
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
