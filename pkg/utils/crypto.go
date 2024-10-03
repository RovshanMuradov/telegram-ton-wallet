package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"go.uber.org/zap"
)

func EncryptData(data []byte, key string) ([]byte, error) {
	logger := logging.GetLogger()
	logger.Debug("Starting data encryption")

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		logger.Error("Failed to create AES cipher", zap.Error(err))
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("Failed to create GCM", zap.Error(err))
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Error("Failed to generate nonce", zap.Error(err))
		return nil, err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	logger.Debug("Data encrypted successfully", zap.Int("encryptedSize", len(encrypted)))
	return encrypted, nil
}

func DecryptData(data []byte, key string) ([]byte, error) {
	logger := logging.GetLogger()
	logger.Debug("Starting data decryption", zap.Int("dataSize", len(data)))

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		logger.Error("Failed to create AES cipher", zap.Error(err))
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("Failed to create GCM", zap.Error(err))
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		logger.Error("Ciphertext too short", zap.Int("dataSize", len(data)), zap.Int("nonceSize", gcm.NonceSize()))
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.Error("Failed to decrypt data", zap.Error(err))
		return nil, err
	}

	logger.Debug("Data decrypted successfully", zap.Int("plaintextSize", len(plaintext)))
	return plaintext, nil
}
