package db

import (
	"github.com/rovshanmuradov/telegram-ton-wallet/internal/logging"
	"go.uber.org/zap"
)

func CheckWalletsTableStructure() {
	var result []struct {
		ColumnName string
		DataType   string
	}

	// Execute a query in the database
	if err := DB.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'wallets'").Scan(&result).Error; err != nil {
		// Log the error using zap
		logging.Error("Error getting wallets table structure", zap.Error(err))
	} else {
		// Log the table structure using zap
		logging.Info("Wallets table structure:")
		for _, col := range result {
			logging.Info("Column info", zap.String("Column", col.ColumnName), zap.String("Type", col.DataType))
		}
	}
}

func CheckWalletsTableIndexes() {
	var result []struct {
		IndexName string
		IndexDef  string
	}

	// Execute a query to the database
	if err := DB.Raw("SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'wallets'").Scan(&result).Error; err != nil {
		// Log the error using zap
		logging.Error("Error checking wallets table indexes", zap.Error(err))
	} else {
		// Log table indexes using zap
		logging.Info("Wallets table indexes:")
		for _, idx := range result {
			logging.Info("Index info", zap.String("Index name", idx.IndexName), zap.String("Definition", idx.IndexDef))
		}
	}
}
