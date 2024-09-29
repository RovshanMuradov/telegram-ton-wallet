# Telegram TON Wallet Bot

A Telegram bot for managing your TON (The Open Network) wallet. This bot allows users to create wallets, check balances, send and receive TON, and view transaction history directly through Telegram.

## Features

- Create TON wallets
- Check wallet balance
- Send TON to other addresses
- Receive TON (get wallet address for top-up)
- View transaction history
- Secure storage of private keys

## Technologies Used

- Go (Golang)
- Docker and Docker Compose
- PostgreSQL
- Telegram Bot API
- TON SDK

## Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- PostgreSQL
- TON API key
- Telegram Bot Token

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/rovshanmuradov/telegram-ton-wallet.git
   cd telegram-ton-wallet
   ```

2. Copy the example environment file and edit it with your configuration:
   ```bash
   cp .env.example .env
   ```
   Edit the .env file with your specific configuration.

3. Build and run the application using Docker Compose:
   ```bash
   docker-compose up --build
   ```

## Configuration

Ensure the following environment variables are set in your .env file:

- `TELEGRAM_TOKEN`: Your Telegram Bot API token
- `TON_API_KEY`: Your TON API key
- `DATABASE_URL`: PostgreSQL connection string
- `ENCRYPTION_KEY`: Key for encrypting private keys (must be 16, 24, or 32 bytes long)
- `TON_CONFIG_URL`: URL for TON network configuration

## Usage

Once the bot is running, you can interact with it on Telegram using the following commands:

- `/start`: Start the bot and get a welcome message
- `/create_wallet`: Create a new TON wallet
- `/balance`: Check your wallet balance
- `/send`: Send TON to another address
- `/receive`: Get your wallet address for receiving TON
- `/history`: View your transaction history
- `/help`: Get a list of available commands

## Development

To run the project locally for development:

1. Ensure you have Go installed on your machine.
2. Install project dependencies:
   ```bash
   go mod download
   ```
3. Run the project:
   ```bash
   go run cmd/main.go
   ```

## Database Migrations

We use golang-migrate for database migrations.

- To create a new migration:
  ```bash
  make migrate-create name=your_migration_name
  ```
- To apply migrations:
  ```bash
  make migrate-up
  ```
- To rollback the last migration:
  ```bash
  make migrate-down
  ```