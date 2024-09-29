package db

import "log"

func CheckWalletsTableStructure() {
	var result []struct {
		ColumnName string
		DataType   string
	}
	if err := DB.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'wallets'").Scan(&result).Error; err != nil {
		log.Printf("Error getting wallets table structure: %v", err)
	} else {
		log.Printf("Wallets table structure:")
		for _, col := range result {
			log.Printf("Column: %s, Type: %s", col.ColumnName, col.DataType)
		}
	}
}

func CheckWalletsTableIndexes() {
	var result []struct {
		IndexName string
		IndexDef  string
	}
	if err := DB.Raw("SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'wallets'").Scan(&result).Error; err != nil {
		log.Printf("Error checking wallets table indexes: %v", err)
	} else {
		log.Printf("Wallets table indexes:")
		for _, idx := range result {
			log.Printf("Index name: %s, Definition: %s", idx.IndexName, idx.IndexDef)
		}
	}
}
