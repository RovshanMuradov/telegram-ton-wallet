# Telegram TON Wallet Bot

A bot for managing your TON wallet in Telegram.

## Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/telegram-ton-wallet.git

2. Go to the project directory:
cd telegram-ton-wallet

3. Install dependencies:
go mod download

## Setup
1. Create a .env file based on .env.example:
cp .env.example .env
2. Fill in the variables in .env:

TELEGRAM_TOKEN – your Telegram bot token.
TON_API_KEY – API key for accessing TON.
DATABASE_URL – connection string to your PostgreSQL database.
ENCRYPTION_KEY – key for encrypting private keys (16, 24 or 32 bytes).