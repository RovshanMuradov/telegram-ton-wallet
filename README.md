# Telegram TON Wallet Bot

Бот для управления кошельком TON в Telegram.

## Установка

1. Клонируйте репозиторий:

   ```bash
   git clone https://github.com/yourusername/telegram-ton-wallet.git

2. Перейдите в директорию проекта:
cd telegram-ton-wallet

3. Установите зависимости:
go mod download

## Настройка
1. Создайте файл .env на основе .env.example:
cp .env.example .env
2. Заполните переменные в .env:

TELEGRAM_TOKEN – токен вашего Telegram бота.
TON_API_KEY – API ключ для доступа к TON.
DATABASE_URL – строка подключения к вашей базе данных PostgreSQL.
ENCRYPTION_KEY – ключ для шифрования приватных ключей (16, 24 или 32 байта).
