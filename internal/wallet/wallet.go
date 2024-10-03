// internal/wallet/wallet.go
package wallet

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/config"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/db"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"github.com/rovshanmuradov/telegram-ton-wallet/pkg/tonutils"
	"github.com/rovshanmuradov/telegram-ton-wallet/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateWallet(userID int64, cfg *config.Config) (*db.Wallet, error) {
	logger := logging.With(zap.Int64("userID", userID))
	logger.Info("Starting wallet creation")

	var wallet *db.Wallet
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Check if user exists
		var user db.User
		if err := tx.Where("telegram_id = ?", userID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				user = db.User{TelegramID: userID}
				if err := tx.Create(&user).Error; err != nil {
					logger.Error("Failed to create user", zap.Error(err))
					return fmt.Errorf("failed to create user: %w", err)
				}
				logger.Info("User created", zap.Int64("internalUserID", user.ID))
			} else {
				logger.Error("Error while searching for user", zap.Error(err))
				return fmt.Errorf("error while searching for user: %w", err)
			}
		}

		// Create wallet
		tonClient, err := tonutils.NewTonClient(cfg)
		if err != nil {
			logger.Error("Failed to create TonClient", zap.Error(err))
			return fmt.Errorf("failed to create TonClient: %w", err)
		}

		w, err := tonClient.CreateWallet("")
		if err != nil {
			logger.Error("Failed to create wallet", zap.Error(err))
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		encryptedPrivateKey, err := EncryptPrivateKey(w.PrivateKey, cfg.EncryptionKey)
		if err != nil {
			logger.Error("Failed to encrypt private key", zap.Error(err))
			return fmt.Errorf("failed to encrypt private key: %w", err)
		}

		logger.Debug("User and wallet details", zap.Int64("internalUserID", user.ID), zap.Int64("telegramUserID", userID))
		wallet = &db.Wallet{
			UserID:     user.ID,
			Address:    w.Address,
			PrivateKey: encryptedPrivateKey,
		}

		if err := tx.Create(wallet).Error; err != nil {
			logger.Error("Failed to save wallet to database", zap.Error(err))
			return fmt.Errorf("failed to save wallet to database: %w", err)
		}

		// Check if wallet was actually saved
		var count int64
		if err := tx.Model(&db.Wallet{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
			logger.Error("Error while checking wallet existence", zap.Error(err))
		} else {
			logger.Debug("Wallets count after creation", zap.Int64("count", count), zap.Int64("internalUserID", user.ID))
		}

		return nil
	})

	if err != nil {
		logger.Error("Error while creating wallet", zap.Error(err))
		return nil, fmt.Errorf("error while creating wallet: %w", err)
	}

	// Check after transaction
	var savedWallet db.Wallet
	if err := db.DB.Where("user_id = ?", wallet.UserID).First(&savedWallet).Error; err != nil {
		logger.Error("Error while checking saved wallet", zap.Error(err))
	} else {
		logger.Debug("Saved wallet details", zap.Any("wallet", savedWallet))
	}

	logger.Info("Wallet successfully created", zap.String("address", wallet.Address))
	return wallet, nil
}

func CreateWalletBackup(userID int64, cfg *config.Config) ([]byte, error) {
	logger := logging.With(zap.Int64("userID", userID))
	logger.Info("Starting wallet backup creation")

	wallet, err := GetWalletByUserID(userID)
	if err != nil {
		logger.Error("Failed to get wallet", zap.Error(err))
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Decrypt the private key
	privateKey, err := DecryptPrivateKey(wallet.PrivateKey, cfg.EncryptionKey)
	if err != nil {
		logger.Error("Failed to decrypt private key", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// Create a backup structure
	backup := struct {
		Address    string `json:"address"`
		PrivateKey string `json:"private_key"`
		Timestamp  int64  `json:"timestamp"`
	}{
		Address:    wallet.Address,
		PrivateKey: privateKey,
		Timestamp:  time.Now().Unix(),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(backup)
	if err != nil {
		logger.Error("Failed to marshal backup data", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal backup data: %w", err)
	}

	// Encrypt the JSON data
	encryptedData, err := utils.EncryptData(jsonData, cfg.EncryptionKey)
	if err != nil {
		logger.Error("Failed to encrypt backup data", zap.Error(err))
		return nil, fmt.Errorf("failed to encrypt backup data: %w", err)
	}

	logger.Info("Wallet backup created successfully")
	return encryptedData, nil
}

func RestoreWalletFromBackup(userID int64, encryptedBackup []byte, cfg *config.Config) error {
	logger := logging.With(zap.Int64("userID", userID))
	logger.Info("Starting wallet restoration from backup")

	// Decrypt the backup data
	jsonData, err := utils.DecryptData(encryptedBackup, cfg.EncryptionKey)
	if err != nil {
		logger.Error("Failed to decrypt backup data", zap.Error(err))
		return fmt.Errorf("failed to decrypt backup data: %w", err)
	}

	// Parse the JSON data
	var backup struct {
		Address    string `json:"address"`
		PrivateKey string `json:"private_key"`
		Timestamp  int64  `json:"timestamp"`
	}
	if err := json.Unmarshal(jsonData, &backup); err != nil {
		logger.Error("Failed to unmarshal backup data", zap.Error(err))
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}

	// Validate the backup data
	if err := ValidateAddress(backup.Address); err != nil {
		logger.Error("Invalid address in backup", zap.String("address", backup.Address), zap.Error(err))
		return fmt.Errorf("invalid address in backup: %w", err)
	}

	// Check if the wallet already exists
	existingWallet, err := GetWalletByUserID(userID)
	if err == nil && existingWallet != nil {
		logger.Warn("User already has a wallet", zap.String("existingAddress", existingWallet.Address))
		return fmt.Errorf("user already has a wallet")
	}

	// Encrypt the private key
	encryptedPrivateKey, err := EncryptPrivateKey(backup.PrivateKey, cfg.EncryptionKey)
	if err != nil {
		logger.Error("Failed to encrypt private key", zap.Error(err))
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create a new wallet entry
	wallet := &db.Wallet{
		UserID:     userID,
		Address:    backup.Address,
		PrivateKey: encryptedPrivateKey,
	}

	if err := db.DB.Create(wallet).Error; err != nil {
		logger.Error("Failed to create wallet in database", zap.Error(err))
		return fmt.Errorf("failed to create wallet in database: %w", err)
	}

	logger.Info("Wallet restored successfully", zap.String("address", wallet.Address))
	return nil
}

func GetWalletByUserID(userID int64) (*db.Wallet, error) {
	logger := logging.With(zap.Int64("userID", userID))
	logger.Info("Attempting to get wallet")

	var user db.User
	if err := db.DB.Where("telegram_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info("User not found")
			return nil, nil
		}
		logger.Error("Error while searching for user", zap.Error(err))
		return nil, err
	}

	var wallet db.Wallet
	if err := db.DB.Where("user_id = ?", user.ID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info("Wallet not found for user")
			return nil, nil
		}
		logger.Error("Error while getting wallet", zap.Error(err))
		return nil, err
	}

	logger.Info("Wallet successfully retrieved", zap.String("address", wallet.Address))
	return &wallet, nil
}

func GetBalance(address string, cfg *config.Config) (string, error) {
	logger := logging.With(zap.String("address", address))
	logger.Info("Attempting to get balance")

	tonClient, err := tonutils.NewTonClient(cfg)
	if err != nil {
		logger.Error("Error while creating TonClient", zap.Error(err))
		return "", fmt.Errorf("failed to create TonClient: %w", err)
	}

	balance, err := tonClient.GetBalance(address)
	if err != nil {
		logger.Error("Error while getting balance", zap.Error(err))
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	logger.Info("Balance retrieved successfully", zap.String("balance", balance))
	return balance, nil
}

func ValidateAddress(address string) error {
	logger := logging.With(zap.String("address", address))
	logger.Debug("Validating TON address")

	// Basic TON address validation (may need refinement)
	match, _ := regexp.MatchString("^[0-9a-fA-F]{48}$", address)
	if !match {
		logger.Warn("Invalid TON address format")
		return fmt.Errorf("invalid TON address format")
	}

	logger.Debug("TON address validated successfully")
	return nil
}

func ValidateAmount(amount string) error {
	logger := logging.With(zap.String("amount", amount))
	logger.Debug("Validating amount")

	_, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		logger.Warn("Invalid amount format", zap.Error(err))
		return fmt.Errorf("invalid amount format")
	}

	// Additional checks can be added here, e.g., minimum amount
	logger.Debug("Amount validated successfully")
	return nil
}

func UpdateWalletBalance(wallet *db.Wallet, cfg *config.Config) error {
	logger := logging.With(zap.String("address", wallet.Address))
	logger.Info("Updating wallet balance")

	balance, err := GetBalance(wallet.Address, cfg)
	if err != nil {
		logger.Error("Failed to get balance", zap.Error(err))
		return err
	}

	wallet.Balance = balance
	if err := db.DB.Save(wallet).Error; err != nil {
		logger.Error("Failed to save updated balance", zap.Error(err))
		return err
	}

	logger.Info("Wallet balance updated successfully", zap.String("balance", balance))
	return nil
}

func LockWallet(wallet *db.Wallet) error {
	logger := logging.With(zap.String("address", wallet.Address))
	logger.Info("Locking wallet")

	wallet.Locked = true
	wallet.LockedAt = time.Now()
	if err := db.DB.Save(wallet).Error; err != nil {
		logger.Error("Failed to lock wallet", zap.Error(err))
		return err
	}

	logger.Info("Wallet locked successfully")
	return nil
}

func UnlockWallet(wallet *db.Wallet) error {
	logger := logging.With(zap.String("address", wallet.Address))
	logger.Info("Unlocking wallet")

	wallet.Locked = false
	wallet.LockedAt = time.Time{}
	if err := db.DB.Save(wallet).Error; err != nil {
		logger.Error("Failed to unlock wallet", zap.Error(err))
		return err
	}

	logger.Info("Wallet unlocked successfully")
	return nil
}

func CheckSuspiciousActivity(wallet *db.Wallet, amount string) bool {
	logger := logging.With(zap.String("address", wallet.Address), zap.String("amount", amount))
	logger.Debug("Checking for suspicious activity")

	threshold, _ := strconv.ParseFloat("1000", 64) // 1000 TON
	sendAmount, _ := strconv.ParseFloat(amount, 64)
	isSuspicious := sendAmount > threshold

	if isSuspicious {
		logger.Warn("Suspicious activity detected", zap.Float64("threshold", threshold))
	} else {
		logger.Debug("No suspicious activity detected")
	}

	return isSuspicious
}

func SendTON(userID int64, toAddress string, amount string, comment string, cfg *config.Config) error {
	logger := logging.With(zap.Int64("userID", userID), zap.String("toAddress", toAddress), zap.String("amount", amount))
	logger.Info("Initiating TON transfer")

	if err := ValidateAddress(toAddress); err != nil {
		logger.Error("Invalid address", zap.Error(err))
		return err
	}
	if err := ValidateAmount(amount); err != nil {
		logger.Error("Invalid amount", zap.Error(err))
		return err
	}

	wallet, err := GetWalletByUserID(userID)
	if err != nil {
		logger.Error("Failed to get user's wallet", zap.Error(err))
		return fmt.Errorf("failed to get user's wallet: %w", err)
	}

	if wallet.Locked {
		logger.Warn("Wallet is locked")
		return fmt.Errorf("wallet is locked")
	}

	if CheckSuspiciousActivity(wallet, amount) {
		logger.Warn("Suspicious activity detected")
		if err := LockWallet(wallet); err != nil {
			logger.Error("Failed to lock wallet", zap.Error(err))
			return err
		}
		return fmt.Errorf("transaction blocked due to suspicious activity")
	}

	privateKey, err := DecryptPrivateKey(wallet.PrivateKey, cfg.EncryptionKey)
	if err != nil {
		logger.Error("Failed to decrypt private key", zap.Error(err))
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	tonClient, err := tonutils.NewTonClient(cfg)
	if err != nil {
		logger.Error("Failed to create TonClient", zap.Error(err))
		return fmt.Errorf("failed to create TonClient: %w", err)
	}

	err = utils.Retry(3, time.Second, func() error {
		return tonClient.SendTransaction(privateKey, toAddress, amount, comment)
	})

	if err != nil {
		logger.Error("Failed to send transaction", zap.Error(err))
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	if err := UpdateWalletBalance(wallet, cfg); err != nil {
		logger.Error("Failed to update wallet balance", zap.Error(err))
		// We don't return an error here as the transaction has already been sent
	}

	logger.Info("Transaction sent successfully")
	return nil
}

func GetTransactionHistory(wallet *db.Wallet, cfg *config.Config) ([]db.Transaction, error) {
	logger := logging.With(zap.String("address", wallet.Address))
	logger.Info("Fetching transaction history")

	var transactions []db.Transaction
	err := db.DB.Where("wallet_id = ?", wallet.ID).Order("created_at desc").Find(&transactions).Error
	if err != nil {
		logger.Error("Failed to fetch transaction history", zap.Error(err))
		return nil, err
	}

	logger.Info("Transaction history fetched successfully", zap.Int("count", len(transactions)))
	return transactions, nil
}

func RecoverWallet(userID int64, seedPhrase string, cfg *config.Config) (*db.Wallet, error) {
	logger := logging.With(zap.Int64("userID", userID))
	logger.Info("Attempting to recover wallet")

	tonClient, err := tonutils.NewTonClient(cfg)
	if err != nil {
		logger.Error("Failed to create TonClient", zap.Error(err))
		return nil, fmt.Errorf("failed to create TonClient: %w", err)
	}

	w, err := tonClient.RecoverWalletFromSeed(seedPhrase)
	if err != nil {
		logger.Error("Failed to recover wallet from seed", zap.Error(err))
		return nil, err
	}

	encryptedPrivateKey, err := EncryptPrivateKey(w.PrivateKey, cfg.EncryptionKey)
	if err != nil {
		logger.Error("Failed to encrypt private key", zap.Error(err))
		return nil, err
	}

	wallet := &db.Wallet{
		UserID:     userID,
		Address:    w.Address,
		PrivateKey: encryptedPrivateKey,
	}

	if err := db.DB.Create(wallet).Error; err != nil {
		logger.Error("Failed to save recovered wallet to database", zap.Error(err))
		return nil, err
	}

	logger.Info("Wallet recovered successfully", zap.String("address", wallet.Address))
	return wallet, nil
}
