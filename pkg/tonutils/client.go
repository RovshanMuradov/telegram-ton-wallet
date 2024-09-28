// pkt/tonutils/client.go
package tonutils

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/rovshanmuradov/telegram-ton-wallet/internal/config"
	"github.com/tyler-smith/go-bip39"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type TonClient struct {
	client *liteclient.ConnectionPool
	ctx    context.Context
	api    *ton.APIClient
}

func NewTonClient(cfg *config.Config) (*TonClient, error) {
	ctx := context.Background()
	client := liteclient.NewConnectionPool()

	err := client.AddConnectionsFromConfigUrl(ctx, cfg.TonConfigURL)
	if err != nil {
		return nil, fmt.Errorf("failed to add connection: %w", err)
	}

	api := ton.NewAPIClient(client)

	return &TonClient{
		client: client,
		ctx:    ctx,
		api:    api,
	}, nil
}

func (c *TonClient) CreateWallet(seedPhrase string) (*Wallet, error) {
	// Используем wallet.NewSeed() вместо предоставленной seed-фразы
	seed := wallet.NewSeed()

	w, err := wallet.FromSeed(c.api, seed, wallet.V3R2)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet from seed: %w (seed words: %v)", err, seed)
	}

	address := w.Address()

	// Преобразуем seed в строку для сохранения
	seedPhrase = strings.Join(seed, " ")

	return &Wallet{
		Address:    address.String(),
		PrivateKey: seedPhrase,
	}, nil
}

func GenerateSeedPhrase() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	words := strings.Split(mnemonic, " ")
	if len(words) != 24 {
		return "", fmt.Errorf("generated mnemonic has invalid length: expected 24 words, got %d", len(words))
	}

	return mnemonic, nil
}

func (c *TonClient) GetBalance(addressStr string) (string, error) {
	addr, err := address.ParseAddr(addressStr)
	if err != nil {
		return "", fmt.Errorf("invalid address: %w", err)
	}

	block, err := c.api.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current block: %w", err)
	}

	account, err := c.api.GetAccount(c.ctx, block, addr)
	if err != nil {
		return "", fmt.Errorf("failed to get account: %w", err)
	}

	if account.IsActive {
		return account.State.Balance.String(), nil
	}

	return "0", nil
}

// Более гибкая и потенциально более точная, но сложнее в реализации и поддержке.
// Она лучше подходит для проектов, где важна точность оценки комиссий и где комиссии могут значительно варьироваться в зависимости от типа транзакции.
func (c *TonClient) EstimateFees(fromAddress string, toAddress string, amount *big.Int) (*big.Int, error) {
	// Константы для приблизительной оценки (в наноTON)
	const (
		baseStorageFee = 10000000 // 0.01 TON
		baseComputeFee = 10000000 // 0.01 TON
		gasPerByte     = 1000     // 0.000001 TON per byte
	)

	// Парсинг адресов
	_, err := address.ParseAddr(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}
	_, err = address.ParseAddr(toAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}

	// Оценка размера сообщения (приблизительно)
	messageSize := 100 + (amount.BitLen() / 8)

	// Расчет приблизительной комиссии
	gasFee := new(big.Int).SetUint64(uint64(messageSize) * gasPerByte)
	totalFee := new(big.Int).Add(
		new(big.Int).SetUint64(baseStorageFee+baseComputeFee),
		gasFee,
	)

	return totalFee, nil
}

// Проще, быстрее, но менее точная.
// Она может подойти для проектов, где скорость работы важнее точности оценки комиссий, или где комиссии относительно стабильны и предсказуемы.
/*func (c *TonClient) EstimateFees(fromSeedPhrase string, toAddress string, amount *big.Int) (*big.Int, error) {
	// Парсинг адреса получателя
	toAddr, err := address.ParseAddr(toAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}

	// Разделение seed-фразы на слова
	seedWords := strings.Split(fromSeedPhrase, " ")

	// Создание кошелька из seed-фразы
	w, err := wallet.FromSeed(c.api, seedWords, wallet.V3R2)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Преобразование суммы в tlb.Coins
	coins := tlb.MustFromTON(amount.String()) // Используем функцию для создания Coins

	// Построение сообщения для транзакции
	_, err = w.BuildTransfer(toAddr, coins, true, "")
	if err != nil {
		return nil, fmt.Errorf("failed to build transfer message: %w", err)
	}

	// Использование фиксированной комиссии
	fixedFee := big.NewInt(10000000) // 0.01 TON

	return fixedFee, nil
}
*/

func (c *TonClient) SendTransaction(privateKey string, toAddress string, amount string, comment string) error {
	// Создание дочернего контекста с тайм-аутом
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Minute)
	defer cancel()

	// Парсинг адреса получателя
	to, err := address.ParseAddr(toAddress)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	// Парсинг суммы
	coins, err := tlb.FromTON(amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Создание кошелька из seed-фразы
	seedWords := strings.Split(privateKey, " ")
	w, err := wallet.FromSeed(c.api, seedWords, wallet.V3R2)
	if err != nil {
		return fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Проверка достаточности баланса
	balance, err := c.GetBalance(w.Address().String())
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	balanceCoins, err := tlb.FromTON(balance)
	if err != nil {
		return fmt.Errorf("invalid balance value: %w", err)
	}
	if balanceCoins.Nano().Cmp(coins.Nano()) < 0 {
		return fmt.Errorf("insufficient balance for transaction")
	}

	// Отправка транзакции с контекстом
	err = w.Transfer(ctx, to, coins, comment, true)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil
}

func (c *TonClient) RecoverWalletFromSeed(seedPhrase string) (*Wallet, error) {
	seedWords := strings.Split(seedPhrase, " ")
	w, err := wallet.FromSeed(c.api, seedWords, wallet.V3R2)
	if err != nil {
		return nil, fmt.Errorf("failed to recover wallet from seed: %w", err)
	}

	return &Wallet{
		Address:    w.Address().String(),
		PrivateKey: seedPhrase,
	}, nil
}

type Wallet struct {
	Address    string
	PrivateKey string // В данном случае это seed-фраза
}
