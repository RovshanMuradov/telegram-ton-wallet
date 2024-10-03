// pkt/tonutils/client.go
package tonutils

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/config"
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"github.com/tyler-smith/go-bip39"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"go.uber.org/zap"
)

type TonClient struct {
	client *liteclient.ConnectionPool
	ctx    context.Context
	api    *ton.APIClient
}

func NewTonClient(cfg *config.Config) (*TonClient, error) {
	logger := logging.GetLogger()
	logger.Info("Creating new TON client")

	ctx := context.Background()
	client := liteclient.NewConnectionPool()

	err := client.AddConnectionsFromConfigUrl(ctx, cfg.TonConfigURL)
	if err != nil {
		logger.Error("Failed to add connection", zap.Error(err))
		return nil, fmt.Errorf("failed to add connection: %w", err)
	}

	api := ton.NewAPIClient(client)

	logger.Info("TON client created successfully")
	return &TonClient{
		client: client,
		ctx:    ctx,
		api:    api,
	}, nil
}

func (c *TonClient) CreateWallet(seedPhrase string) (*Wallet, error) {
	logger := logging.GetLogger()
	logger.Info("Creating new wallet")

	var seed []string
	if seedPhrase == "" {
		logger.Debug("Generating new seed phrase")
		seed = wallet.NewSeed()
	} else {
		logger.Debug("Using provided seed phrase")
		seed = strings.Split(seedPhrase, " ")
	}

	w, err := wallet.FromSeed(c.api, seed, wallet.V3R2)
	if err != nil {
		logger.Error("Failed to create wallet from seed", zap.Error(err))
		return nil, fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	address := w.Address()
	finalSeedPhrase := strings.Join(seed, " ")

	logger.Info("Wallet created successfully", zap.String("address", address.String()))
	return &Wallet{
		Address:    address.String(),
		PrivateKey: finalSeedPhrase,
	}, nil
}

func GenerateSeedPhrase() (string, error) {
	logger := logging.GetLogger()
	logger.Info("Generating new seed phrase")

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		logger.Error("Failed to generate entropy", zap.Error(err))
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		logger.Error("Failed to generate mnemonic", zap.Error(err))
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	words := strings.Split(mnemonic, " ")
	if len(words) != 24 {
		logger.Error("Generated mnemonic has invalid length", zap.Int("wordCount", len(words)))
		return "", fmt.Errorf("generated mnemonic has invalid length: expected 24 words, got %d", len(words))
	}

	logger.Info("Seed phrase generated successfully")
	return mnemonic, nil
}

func (c *TonClient) GetBalance(addressStr string) (string, error) {
	logger := logging.With(zap.String("address", addressStr))
	logger.Info("Getting balance")

	var balance string

	// Определяем функцию для повторения
	getBalanceFunc := func() error {
		addr, err := address.ParseAddr(addressStr)
		if err != nil {
			logger.Error("Invalid address", zap.Error(err))
			return fmt.Errorf("invalid address: %w", err)
		}

		block, err := c.api.CurrentMasterchainInfo(c.ctx)
		if err != nil {
			logger.Error("Failed to get current block", zap.Error(err))
			return fmt.Errorf("failed to get current block: %w", err)
		}

		account, err := c.api.GetAccount(c.ctx, block, addr)
		if err != nil {
			logger.Error("Failed to get account", zap.Error(err))
			return fmt.Errorf("failed to get account: %w", err)
		}

		if account.IsActive {
			balance = account.State.Balance.String()
		} else {
			balance = "0"
		}

		return nil
	}

	// Используем функцию retry
	err := retry.Do(
		getBalanceFunc,
		retry.Attempts(3),
		retry.Delay(1*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			logger.Warn("Retry attempt", zap.Uint("attempt", n), zap.Error(err))
		}),
	)

	if err != nil {
		logger.Error("Failed to get balance after retries", zap.Error(err))
		return "", fmt.Errorf("failed to get balance after retries: %w", err)
	}

	logger.Info("Balance retrieved successfully", zap.String("balance", balance))
	return balance, nil
}

// More flexible and potentially more accurate, but more complex to implement and maintain.
// It's better suited for projects where fee estimation accuracy is important and where fees can vary significantly depending on the transaction type.
func (c *TonClient) EstimateFees(fromAddress string, toAddress string, amount *big.Int) (*big.Int, error) {
	logger := logging.With(zap.String("fromAddress", fromAddress), zap.String("toAddress", toAddress), zap.String("amount", amount.String()))
	logger.Info("Estimating transaction fees")

	// Constants for approximate estimation (in nanoTON)
	const (
		baseStorageFee = 10000000 // 0.01 TON
		baseComputeFee = 10000000 // 0.01 TON
		gasPerByte     = 1000     // 0.000001 TON per byte
	)

	// Parsing addresses
	_, err := address.ParseAddr(fromAddress)
	if err != nil {
		logger.Error("Invalid sender address", zap.Error(err))
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}
	_, err = address.ParseAddr(toAddress)
	if err != nil {
		logger.Error("Invalid recipient address", zap.Error(err))
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}

	// Estimating message size (approximately)
	messageSize := 100 + (amount.BitLen() / 8)

	// Calculating approximate fee
	gasFee := new(big.Int).SetUint64(uint64(messageSize) * gasPerByte)
	totalFee := new(big.Int).Add(
		new(big.Int).SetUint64(baseStorageFee+baseComputeFee),
		gasFee,
	)

	logger.Info("Fee estimation completed", zap.String("estimatedFee", totalFee.String()))
	return totalFee, nil
}

// Simpler, faster, but less accurate.
// It may be suitable for projects where speed is more important than fee estimation accuracy, or where fees are relatively stable and predictable.
/*func (c *TonClient) EstimateFees(fromSeedPhrase string, toAddress string, amount *big.Int) (*big.Int, error) {
	// Parsing recipient address
	toAddr, err := address.ParseAddr(toAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}

	// Splitting seed phrase into words
	seedWords := strings.Split(fromSeedPhrase, " ")

	// Creating wallet from seed phrase
	w, err := wallet.FromSeed(c.api, seedWords, wallet.V3R2)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Converting amount to tlb.Coins
	coins := tlb.MustFromTON(amount.String()) // Using function to create Coins

	// Building transaction message
	_, err = w.BuildTransfer(toAddr, coins, true, "")
	if err != nil {
		return nil, fmt.Errorf("failed to build transfer message: %w", err)
	}

	// Using fixed fee
	fixedFee := big.NewInt(10000000) // 0.01 TON

	return fixedFee, nil
}
*/

func (c *TonClient) SendTransaction(privateKey string, toAddress string, amount string, comment string) (string, error) {
	logger := logging.With(zap.String("toAddress", toAddress), zap.String("amount", amount))
	logger.Info("Initiating transaction")

	// Creating child context with timeout
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Minute)
	defer cancel()

	// Parsing recipient address
	to, err := address.ParseAddr(toAddress)
	if err != nil {
		logger.Error("Invalid recipient address", zap.Error(err))
		return "", fmt.Errorf("invalid recipient address: %w", err)
	}

	// Parsing amount
	coins, err := tlb.FromTON(amount)
	if err != nil {
		logger.Error("Invalid amount", zap.Error(err))
		return "", fmt.Errorf("invalid amount: %w", err)
	}

	// Creating wallet from seed phrase
	seedWords := strings.Split(privateKey, " ")
	w, err := wallet.FromSeed(c.api, seedWords, wallet.V3R2)
	if err != nil {
		logger.Error("Failed to create wallet from seed", zap.Error(err))
		return "", fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Checking balance sufficiency
	balance, err := c.GetBalance(w.Address().String())
	if err != nil {
		logger.Error("Failed to get balance", zap.Error(err))
		return "", fmt.Errorf("failed to get balance: %w", err)
	}
	balanceCoins, err := tlb.FromTON(balance)
	if err != nil {
		logger.Error("Invalid balance value", zap.Error(err))
		return "", fmt.Errorf("invalid balance value: %w", err)
	}
	if balanceCoins.Nano().Cmp(coins.Nano()) < 0 {
		logger.Warn("Insufficient balance for transaction",
			zap.String("balance", balance),
			zap.String("requiredAmount", amount))
		return "", fmt.Errorf("insufficient balance for transaction")
	}

	// Convert comment to cell.Cell
	commentCell := cell.BeginCell().MustStoreUInt(0, 32).MustStoreStringSnake(comment).EndCell()

	// Using SendWaitTransaction
	tx, block, err := w.SendWaitTransaction(ctx, &wallet.Message{
		Mode: 1, // 1 means pay fees separately
		InternalMessage: &tlb.InternalMessage{
			Bounce:  true,
			DstAddr: to,
			Amount:  coins,
			Body:    commentCell,
		},
	})
	if err != nil {
		logger.Error("Failed to send transaction", zap.Error(err))
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := hex.EncodeToString(tx.Hash)
	logger.Info("Transaction sent successfully",
		zap.String("txHash", txHash),
		zap.Uint32("blockSeqNo", block.SeqNo))

	return txHash, nil
}

func (c *TonClient) RecoverWalletFromSeed(seedPhrase string) (*Wallet, error) {
	logger := logging.GetLogger()
	logger.Info("Recovering wallet from seed phrase")

	seedWords := strings.Split(seedPhrase, " ")
	w, err := wallet.FromSeed(c.api, seedWords, wallet.V3R2)
	if err != nil {
		logger.Error("Failed to recover wallet from seed", zap.Error(err))
		return nil, fmt.Errorf("failed to recover wallet from seed: %w", err)
	}

	recoveredWallet := &Wallet{
		Address:    w.Address().String(),
		PrivateKey: seedPhrase,
	}

	logger.Info("Wallet recovered successfully", zap.String("address", recoveredWallet.Address))
	return recoveredWallet, nil
}

type Wallet struct {
	Address    string
	PrivateKey string // In this case, it's the seed phrase
}
