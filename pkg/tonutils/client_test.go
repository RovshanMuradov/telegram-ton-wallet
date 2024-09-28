package tonutils

import (
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/tyler-smith/go-bip39"
)

func TestGenerateSeedPhrase(t *testing.T) {
	// Тест генерации seed-фразы
	seedPhrase, err := GenerateSeedPhrase()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Проверяем, что seed-фраза не пустая
	if seedPhrase == "" {
		t.Fatal("expected non-empty seed phrase")
	}

	// Проверяем, что seed-фраза состоит из 24 слов
	words := strings.Split(seedPhrase, " ")
	if len(words) != 24 {
		t.Fatalf("expected 24 words, got %d", len(words))
	}

	// Проверяем, что seed-фраза может быть преобразована в энтропию
	_, err = bip39.MnemonicToByteArray(seedPhrase)
	if err != nil {
		t.Fatalf("expected valid mnemonic, got error: %v", err)
	}
}

func TestGenerateSeedPhraseMultiple(t *testing.T) {
	// Генерируем несколько seed-фраз
	seedPhrase1, err := GenerateSeedPhrase()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	seedPhrase2, err := GenerateSeedPhrase()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Проверяем, что разные вызовы возвращают разные seed-фразы
	if seedPhrase1 == seedPhrase2 {
		t.Fatal("expected different seed phrases, got identical ones")
	}
}

func TestGenerateSeedPhraseNotPredictable(t *testing.T) {
	// Генерируем seed-фразу
	seedPhrase, err := GenerateSeedPhrase()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Проверяем, что seed-фраза не является каким-либо известным фиксированным значением
	// (этот тест может быть адаптирован под конкретные известные случаи, если такие существуют)
	if seedPhrase == "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about" {
		t.Fatal("generated seed phrase is too predictable")
	}
}

func TestNewTonClient(t *testing.T) {
	// Загрузка переменных из .env файла
	err := godotenv.Load()
	if err != nil {
		t.Fatalf("Ошибка при загрузке .env файла: %v", err)
	}

	t.Run("Успешное подключение", func(t *testing.T) {
		apiKey := os.Getenv("TON_API_KEY")

		if apiKey == "" {
			t.Fatal("TON_API_KEY не установлен в .env файле")
		}

		client, err := NewTonClient(apiKey)
		if err != nil {
			t.Fatalf("Ошибка при создании TonClient: %v", err)
		}

		if client == nil {
			t.Fatal("TonClient не должен быть nil")
		}

		// Проверка, что клиент действительно подключен к сети TON
		block, err := client.api.CurrentMasterchainInfo(client.ctx)
		if err != nil {
			t.Fatalf("Ошибка при получении информации о текущем блоке: %v", err)
		}

		if block.SeqNo == 0 {
			t.Fatal("Номер блока не должен быть равен 0")
		}
	})

	t.Run("Неверный API ключ", func(t *testing.T) {
		_, err := NewTonClient("неверный_ключ")
		if err == nil {
			t.Fatal("Ожидалась ошибка при использовании неверного API ключа")
		}
	})

	t.Run("Пустой API ключ", func(t *testing.T) {
		_, err := NewTonClient("")
		if err == nil {
			t.Fatal("Ожидалась ошибка при использовании пустого API ключа")
		}
	})
}
