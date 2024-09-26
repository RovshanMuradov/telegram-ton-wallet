package tonutils

import (
	"context"
	"fmt"
	"math/big"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type TonClient struct {
	client *liteclient.ConnectionPool
	ctx    context.Context
}

func NewTonClient(apiKey string) (*TonClient, error) {
	ctx := context.Background()
	client := liteclient.NewConnectionPool()

	// Настройка клиента для подключения к TON сети
	err := client.AddConnection(ctx, "https://ton.org/api/v2/jsonRPC", apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to add connection: %w", err)
	}

	return &TonClient{
		client: client,
		ctx:    ctx,
	}, nil
}

func (c *TonClient) CreateWallet() (*Wallet, error) {
	// Создание нового кошелька
	w, err := wallet.NewWallet(wallet.V3R2, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Получение адреса кошелька
	address := w.Address()

	// Получение приватного ключа (в виде seed-фразы)
	words := w.GetSeed()

	return &Wallet{
		Address:    address.String(),
		PrivateKey: words,
	}, nil
}

func (c *TonClient) GetBalance(address string) (string, error) {
	api := ton.NewAPIClient(c.client)
	addr, err := ton.ParseAddress(address)
	if err != nil {
		return "", fmt.Errorf("invalid address: %w", err)
	}

	balance, err := api.GetBalance(c.ctx, addr)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	return balance.String(), nil
}

func (c *TonClient) EstimateFees(fromAddress string, toAddress string, amount *big.Int) (*big.Int, error) {
	from, err := ton.ParseAddress(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}
	to, err := ton.ParseAddress(toAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}

	api := ton.NewAPIClient(c.client)
	fees, err := api.EstimateFee(c.ctx, from, to, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate fee: %w", err)
	}

	return fees.Total, nil
}

func (c *TonClient) SendTransaction(privateKey string, toAddress string, amount string) error {
	// Парсинг адреса и суммы
	to, err := ton.ParseAddress(toAddress)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	amountValue, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("invalid amount")
	}

	// Создание кошелька
	w, err := wallet.FromSeed(c.client, privateKey, wallet.V3R2)
	if err != nil {
		return fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Оценка комиссии
	fee, err := c.EstimateFees(w.Address().String(), toAddress, amountValue)
	if err != nil {
		return fmt.Errorf("failed to estimate fee: %w", err)
	}

	// Проверка достаточности баланса
	balance, err := c.GetBalance(w.Address().String())
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	balanceValue, _ := new(big.Int).SetString(balance, 10)
	totalRequired := new(big.Int).Add(amountValue, fee)
	if balanceValue.Cmp(totalRequired) < 0 {
		return fmt.Errorf("insufficient balance for transaction and fee")
	}

	// Отправка транзакции с учетом комиссии
	err = w.Transfer(c.ctx, to, amountValue, "")
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil
}

func (c *TonClient) RecoverWalletFromSeed(seedPhrase string) (*Wallet, error) {
	w, err := wallet.FromSeed(c.client, seedPhrase, wallet.V3R2)
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
