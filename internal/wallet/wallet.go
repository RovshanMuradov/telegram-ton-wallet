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
)

func CreateWallet(userID int, cfg *config.Config) (*db.Wallet, error) {
	log.Printf("Начало создания кошелька для пользователя %d", userID)

	tonClient, err := tonutils.NewTonClient(cfg.TonAPIKey)
	if err != nil {
		log.Printf("Ошибка при создании TonClient: %v", err)
		return nil, fmt.Errorf("не удалось создать TonClient: %w", err)
	}
	log.Printf("TonClient успешно создан")

	// Генерируем новую seed-фразу
	seedPhrase, err := tonutils.GenerateSeedPhrase()
	if err != nil {
		log.Printf("Ошибка при генерации seed-фразы: %v", err)
		return nil, fmt.Errorf("не удалось сгенерировать seed-фразу: %w", err)
	}
	log.Printf("Seed-фраза успешно сгенерирована")

	w, err := tonClient.CreateWallet(seedPhrase)
	if err != nil {
		log.Printf("Ошибка при создании кошелька: %v", err)
		return nil, fmt.Errorf("не удалось создать кошелек: %w", err)
	}
	log.Printf("Кошелек успешно создан в TON")

	encryptedPrivateKey, err := EncryptPrivateKey(w.PrivateKey, cfg.EncryptionKey)
	if err != nil {
		log.Printf("Ошибка при шифровании приватного ключа: %v", err)
		return nil, fmt.Errorf("не удалось зашифровать приватный ключ: %w", err)
	}
	log.Printf("Приватный ключ успешно зашифрован")

	wallet := &db.Wallet{
		UserID:     userID,
		Address:    w.Address,
		PrivateKey: encryptedPrivateKey,
	}

	if err := db.DB.Create(wallet).Error; err != nil {
		log.Printf("Ошибка при сохранении кошелька в базу данных: %v", err)
		return nil, fmt.Errorf("не удалось сохранить кошелек в базу данных: %w", err)
	}

	log.Printf("Создан новый кошелек для пользователя %d: %s", userID, w.Address)
	return wallet, nil
}

func GetWalletByUserID(userID int) (*db.Wallet, error) {
	var wallet db.Wallet
	if err := db.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func GetBalance(address string, cfg *config.Config) (string, error) {
	tonClient, err := tonutils.NewTonClient(cfg.TonAPIKey)
	if err != nil {
		log.Printf("Ошибка при создании TonClient: %v", err)
		return "", fmt.Errorf("не удалось создать TonClient: %w", err)
	}

	balance, err := tonClient.GetBalance(address)
	if err != nil {
		log.Printf("Ошибка при получении баланса для адреса %s: %v", address, err)
		return "", fmt.Errorf("не удалось получить баланс: %w", err)
	}

	log.Printf("Получен баланс для адреса %s: %s", address, balance)
	return balance, nil
}

func ValidateAddress(address string) error {
	// Примерная валидация адреса TON (может потребоваться уточнение)
	match, _ := regexp.MatchString("^[0-9a-fA-F]{48}$", address)
	if !match {
		return fmt.Errorf("неверный формат адреса TON")
	}
	return nil
}

func ValidateAmount(amount string) error {
	_, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("неверный формат суммы")
	}
	// Здесь можно добавить дополнительные проверки, например, на минимальную сумму
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
	// Пример проверки: блокировка при отправке большой суммы
	threshold, _ := strconv.ParseFloat("1000", 64) // 1000 TON
	sendAmount, _ := strconv.ParseFloat(amount, 64)
	return sendAmount > threshold
}

func SendTON(userID int, toAddress string, amount string, comment string, cfg *config.Config) error {
	if err := ValidateAddress(toAddress); err != nil {
		return err
	}
	if err := ValidateAmount(amount); err != nil {
		return err
	}

	wallet, err := GetWalletByUserID(userID)
	if err != nil {
		log.Printf("Ошибка при получении кошелька пользователя %d: %v", userID, err)
		return fmt.Errorf("не удалось получить кошелек пользователя: %w", err)
	}

	if wallet.Locked {
		return fmt.Errorf("кошелек заблокирован")
	}

	if CheckSuspiciousActivity(wallet, amount) {
		if err := LockWallet(wallet); err != nil {
			return err
		}
		return fmt.Errorf("транзакция заблокирована из-за подозрительной активности")
	}

	privateKey, err := DecryptPrivateKey(wallet.PrivateKey, cfg.EncryptionKey)
	if err != nil {
		log.Printf("Ошибка при расшифровке приватного ключа для пользователя %d: %v", userID, err)
		return fmt.Errorf("не удалось расшифровать приватный ключ: %w", err)
	}

	tonClient, err := tonutils.NewTonClient(cfg.TonAPIKey)
	if err != nil {
		log.Printf("Ошибка при создании TonClient: %v", err)
		return fmt.Errorf("не удалось создать TonClient: %w", err)
	}

	err = utils.Retry(3, time.Second, func() error {
		return tonClient.SendTransaction(privateKey, toAddress, amount, comment)
	})

	if err != nil {
		log.Printf("Ошибка при отправке транзакции от пользователя %d на адрес %s: %v", userID, toAddress, err)
		return fmt.Errorf("не удалось отправить транзакцию: %w", err)
	}

	if err := UpdateWalletBalance(wallet, cfg); err != nil {
		log.Printf("Ошибка при обновлении баланса кошелька пользователя %d: %v", userID, err)
		// Не возвращаем ошибку, так как транзакция уже отправлена
	}

	log.Printf("Успешно отправлено %s TON от пользователя %d на адрес %s", amount, userID, toAddress)
	return nil
}

func GetTransactionHistory(wallet *db.Wallet, cfg *config.Config) ([]db.Transaction, error) {
	var transactions []db.Transaction
	err := db.DB.Where("wallet_id = ?", wallet.ID).Order("created_at desc").Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	// Здесь можно добавить получение дополнительной информации о транзакциях из TON API
	return transactions, nil
}

func RecoverWallet(userID int, seedPhrase string, cfg *config.Config) (*db.Wallet, error) {
	tonClient, err := tonutils.NewTonClient(cfg.TonAPIKey)
	if err != nil {
		log.Printf("Ошибка при создании TonClient: %v", err)
		return nil, fmt.Errorf("не удалось создать TonClient: %w", err)
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
