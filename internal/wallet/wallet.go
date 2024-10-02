// internal/wallet/wallet.go
package wallet

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/config"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/db"
	"github.com/rovshanmuradov/telegram-ton-wallet/pkg/tonutils"
	"github.com/rovshanmuradov/telegram-ton-wallet/pkg/utils"
	"gorm.io/gorm"
)

func CreateWallet(userID int64, cfg *config.Config) (*db.Wallet, error) {
	log.Printf("Starting wallet creation for user %d", userID)

	var wallet *db.Wallet
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Check if user exists
		var user db.User
		if err := tx.Where("telegram_id = ?", userID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				user = db.User{TelegramID: userID}
				if err := tx.Create(&user).Error; err != nil {
					return fmt.Errorf("failed to create user: %w", err)
				}
				log.Printf("User %d successfully created", userID)
			} else {
				return fmt.Errorf("error while searching for user: %w", err)
			}
		}

		// Create wallet
		tonClient, err := tonutils.NewTonClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create TonClient: %w", err)
		}

		w, err := tonClient.CreateWallet("")
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		encryptedPrivateKey, err := EncryptPrivateKey(w.PrivateKey, cfg.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt private key: %w", err)
		}

		log.Printf("user.ID: %d, userID: %d", user.ID, userID)
		wallet = &db.Wallet{
			UserID:     user.ID,
			Address:    w.Address,
			PrivateKey: encryptedPrivateKey,
		}

		if err := tx.Create(wallet).Error; err != nil {
			log.Printf("Error while saving wallet to DB: %v", err)
			return fmt.Errorf("failed to save wallet to database: %w", err)
		}

		// Check if wallet was actually saved
		var count int64
		if err := tx.Model(&db.Wallet{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
			log.Printf("Error while checking wallet existence: %v", err)
		} else {
			log.Printf("Number of wallets for user %d after creation: %d", user.ID, count)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while creating wallet: %w", err)
	}

	// Check after transaction
	var savedWallet db.Wallet
	if err := db.DB.Where("user_id = ?", wallet.UserID).First(&savedWallet).Error; err != nil {
		log.Printf("Error while checking saved wallet: %v", err)
	} else {
		log.Printf("Saved wallet: %+v", savedWallet)
	}

	log.Printf("Wallet successfully created for user %d with address %s", userID, wallet.Address)
	return wallet, nil
}

func GetWalletByUserID(userID int64) (*db.Wallet, error) {
	log.Printf("Attempting to get wallet for user %d", userID)

	var user db.User
	if err := db.DB.Where("telegram_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Printf("Error while searching for user: %v", err)
		return nil, err
	}

	var wallet db.Wallet
	if err := db.DB.Where("user_id = ?", user.ID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Printf("Error while getting wallet for user %d: %v", userID, err)
		return nil, err
	}

	log.Printf("Wallet successfully retrieved for user %d: %+v", userID, wallet)
	return &wallet, nil
}

func GetBalance(address string, cfg *config.Config) (string, error) {
	tonClient, err := tonutils.NewTonClient(cfg)
	if err != nil {
		log.Printf("Error while creating TonClient: %v", err)
		return "", fmt.Errorf("failed to create TonClient: %w", err)
	}

	balance, err := tonClient.GetBalance(address)
	if err != nil {
		log.Printf("Error while getting balance for address %s: %v", address, err)
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	log.Printf("Balance retrieved for address %s: %s", address, balance)
	return balance, nil
}

func ValidateAddress(address string) error {
	// Basic TON address validation (may need refinement)
	match, _ := regexp.MatchString("^[0-9a-fA-F]{48}$", address)
	if !match {
		return fmt.Errorf("invalid TON address format")
	}
	return nil
}

func ValidateAmount(amount string) error {
	_, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount format")
	}
	// Additional checks can be added here, e.g., minimum amount
	return nil
}

func UpdateWalletBalance(wallet *db.Wallet, cfg *config.Config) error {
	balance, err := GetBalance(wallet.Address, cfg)
	if err != nil {
		return err
	}

	wallet.Balance = balance
	return db.DB.Save(wallet).Error
}

func LockWallet(wallet *db.Wallet) error {
	wallet.Locked = true
	wallet.LockedAt = time.Now()
	return db.DB.Save(wallet).Error
}

func UnlockWallet(wallet *db.Wallet) error {
	wallet.Locked = false
	wallet.LockedAt = time.Time{}
	return db.DB.Save(wallet).Error
}

func CheckSuspiciousActivity(wallet *db.Wallet, amount string) bool {
	// Example check: block large transactions
	threshold, _ := strconv.ParseFloat("1000", 64) // 1000 TON
	sendAmount, _ := strconv.ParseFloat(amount, 64)
	return sendAmount > threshold
}

func SendTON(userID int64, toAddress string, amount string, comment string, cfg *config.Config) error {
	if err := ValidateAddress(toAddress); err != nil {
		return err
	}
	if err := ValidateAmount(amount); err != nil {
		return err
	}

	wallet, err := GetWalletByUserID(userID)
	if err != nil {
		log.Printf("Error while getting wallet for user %d: %v", userID, err)
		return fmt.Errorf("failed to get user's wallet: %w", err)
	}

	if wallet.Locked {
		return fmt.Errorf("wallet is locked")
	}

	if CheckSuspiciousActivity(wallet, amount) {
		if err := LockWallet(wallet); err != nil {
			return err
		}
		return fmt.Errorf("transaction blocked due to suspicious activity")
	}

	privateKey, err := DecryptPrivateKey(wallet.PrivateKey, cfg.EncryptionKey)
	if err != nil {
		log.Printf("Error while decrypting private key for user %d: %v", userID, err)
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	tonClient, err := tonutils.NewTonClient(cfg)
	if err != nil {
		log.Printf("Error while creating TonClient: %v", err)
		return fmt.Errorf("failed to create TonClient: %w", err)
	}

	err = utils.Retry(3, time.Second, func() error {
		return tonClient.SendTransaction(privateKey, toAddress, amount, comment)
	})

	if err != nil {
		log.Printf("Error while sending transaction from user %d to address %s: %v", userID, toAddress, err)
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	if err := UpdateWalletBalance(wallet, cfg); err != nil {
		log.Printf("Error while updating wallet balance for user %d: %v", userID, err)
		// We don't return an error here as the transaction has already been sent
	}

	log.Printf("Successfully sent %s TON from user %d to address %s", amount, userID, toAddress)
	return nil
}

func GetTransactionHistory(wallet *db.Wallet, cfg *config.Config) ([]db.Transaction, error) {
	var transactions []db.Transaction
	err := db.DB.Where("wallet_id = ?", wallet.ID).Order("created_at desc").Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	// Additional transaction information from TON API can be added here
	return transactions, nil
}

func RecoverWallet(userID int64, seedPhrase string, cfg *config.Config) (*db.Wallet, error) {
	tonClient, err := tonutils.NewTonClient(cfg)
	if err != nil {
		log.Printf("Error while creating TonClient: %v", err)
		return nil, fmt.Errorf("failed to create TonClient: %w", err)
	}

	w, err := tonClient.RecoverWalletFromSeed(seedPhrase)
	if err != nil {
		return nil, err
	}

	encryptedPrivateKey, err := EncryptPrivateKey(w.PrivateKey, cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}

	wallet := &db.Wallet{
		UserID:     userID,
		Address:    w.Address,
		PrivateKey: encryptedPrivateKey,
	}

	if err := db.DB.Create(wallet).Error; err != nil {
		return nil, err
	}

	return wallet, nil
}
