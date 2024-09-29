// internal/db/db.go
package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(databaseURL string) error {
	var err error
	var sqlDB *gorm.DB

	for i := 0; i < 30; i++ {
		sqlDB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Попытка подключения к базе данных %d/30: %v", i+1, err)
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных после 30 попыток: %w", err)
	}

	DB = sqlDB

	// Выполнение миграций
	if err := runMigrations(databaseURL); err != nil {
		return fmt.Errorf("ошибка при выполнении миграций: %w", err)
	}

	return nil
}

func CheckWalletsTableStructure() {
	var result []struct {
		ColumnName string
		DataType   string
	}
	if err := DB.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'wallets'").Scan(&result).Error; err != nil {
		log.Printf("Ошибка при получении структуры таблицы wallets: %v", err)
	} else {
		log.Printf("Структура таблицы wallets:")
		for _, col := range result {
			log.Printf("Колонка: %s, Тип: %s", col.ColumnName, col.DataType)
		}
	}
}

func CheckWalletsTableIndexes() {
	var result []struct {
		IndexName string
		IndexDef  string
	}
	if err := DB.Raw("SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'wallets'").Scan(&result).Error; err != nil {
		log.Printf("Ошибка при проверке индексов таблицы wallets: %v", err)
	} else {
		log.Printf("Индексы таблицы wallets:")
		for _, idx := range result {
			log.Printf("Имя индекса: %s, Определение: %s", idx.IndexName, idx.IndexDef)
		}
	}
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("ошибка при получении sql.DB: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

func runMigrations(databaseURL string) error {
	log.Println("Начало выполнения миграций")
	migrationsPath := "/app/migrations/migrations" // Обновленный путь
	log.Println("Путь к миграциям:", migrationsPath)

	// Выведем список файлов в директории
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		log.Printf("Ошибка при чтении директории миграций: %v", err)
	} else {
		for _, file := range files {
			log.Println("Файл миграции:", file.Name())
		}
	}

	m, err := migrate.New(
		"file://"+migrationsPath, // Используем обновленный путь
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("ошибка при инициализации миграций: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка при выполнении миграций: %w", err)
	}

	log.Println("Миграции успешно выполнены")
	return nil
}
