package tonutils

// import (
// 	"context"
// 	"log"
// 	"os"
// 	"testing"

// 	"github.com/joho/godotenv"
// )

// func init() {
// 	err := godotenv.Load("../../.env")
// 	if err != nil {
// 		log.Printf("Warning: .env file not found, environment variables must be set manually for tests")
// 	}
// }

// func TestNewTonClient(t *testing.T) {
// 	err := godotenv.Load()
// 	if err != nil {
// 		t.Fatalf("Ошибка при загрузке .env файла: %v", err)
// 	}

// 	configURL := os.Getenv("TON_CONFIG_URL")
// 	if configURL == "" {
// 		t.Fatal("TON_CONFIG_URL не установлен в .env файле")
// 	}

// 	t.Run("Успешное подключение", func(t *testing.T) {
// 		client, err := NewTonClient(configURL)
// 		if err != nil {
// 			t.Fatalf("Ошибка при создании TonClient: %v", err)
// 		}

// 		if client == nil {
// 			t.Fatal("TonClient не должен быть nil")
// 		}

// 		// Проверка, что клиент действительно подключен к сети TON
// 		block, err := client.api.CurrentMasterchainInfo(context.Background())
// 		if err != nil {
// 			t.Fatalf("Ошибка при получении информации о текущем блоке: %v", err)
// 		}

// 		if block.SeqNo == 0 {
// 			t.Fatal("Номер блока не должен быть равен 0")
// 		}
// 	})

// 	t.Run("Неверный URL конфигурации", func(t *testing.T) {
// 		_, err := NewTonClient("https://invalid-url.com/config.json")
// 		if err == nil {
// 			t.Fatal("Ожидалась ошибка при использовании неверного URL конфигурации")
// 		}
// 	})

// 	t.Run("Пустой URL конфигурации", func(t *testing.T) {
// 		_, err := NewTonClient("")
// 		if err == nil {
// 			t.Fatal("Ожидалась ошибка при использовании пустого URL конфигурации")
// 		}
// 	})
// }
