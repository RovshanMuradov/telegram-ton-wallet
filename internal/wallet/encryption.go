// internal/wallet/encryption.go
package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"go.uber.org/zap"
)

const (
	keySize = 32 // AES-256
)

func EncryptPrivateKey(privateKey string, encryptionKey string) (string, error) {
	if privateKey == "" {
		logging.Error("Encryption failed: private key is empty")
		return "", errors.New("private key cannot be empty")
	}
	if len(encryptionKey) < keySize {
		err := fmt.Errorf("encryption key must be at least %d bytes long", keySize)
		logging.Error("Encryption failed: encryption key too short", zap.Error(err))
		return "", err
	}

	key := []byte(encryptionKey)[:keySize]
	plaintext := []byte(privateKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		logging.Error("Error creating AES cipher for encryption", zap.Error(err))
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logging.Error("Error creating GCM for encryption", zap.Error(err))
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		logging.Error("Error reading random nonce for encryption", zap.Error(err))
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	logging.Info("Private key encrypted successfully")
	return hex.EncodeToString(ciphertext), nil
}

func DecryptPrivateKey(encryptedPrivateKey string, encryptionKey string) (string, error) {
	if encryptedPrivateKey == "" {
		logging.Error("Decryption failed: encrypted private key is empty")
		return "", errors.New("encrypted private key cannot be empty")
	}
	if len(encryptionKey) < keySize {
		err := fmt.Errorf("encryption key must be at least %d bytes long", keySize)
		logging.Error("Decryption failed: encryption key too short", zap.Error(err))
		return "", err
	}

	key := []byte(encryptionKey)[:keySize]
	ciphertext, err := hex.DecodeString(encryptedPrivateKey)
	if err != nil {
		logging.Error("Error decoding encrypted private key", zap.Error(err))
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		logging.Error("Error creating AES cipher for decryption", zap.Error(err))
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logging.Error("Error creating GCM for decryption", zap.Error(err))
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize() {
		err := errors.New("ciphertext too short")
		logging.Error("Decryption failed: ciphertext too short", zap.Error(err))
		return "", err
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logging.Error("Error decrypting private key", zap.Error(err))
		return "", err
	}

	logging.Info("Private key decrypted successfully")
	return string(plaintext), nil
}
