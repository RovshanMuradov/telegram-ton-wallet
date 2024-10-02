# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Mechanism for wallet backup
- Functionality for wallet restoration

### Changed
- Improved error handling in handlers.go

### Fixed
- Limited wallet creation to one per user
- Added check for existing wallet
- Implemented display of existing wallet info on repeated request

### Planned
- Improve transaction sending process
- Optimize blockchain interactions
- Enhance security and recovery mechanisms
- Implement spam and abuse protection
- Improve user experience
- Expand functionality
- Enhance testing and debugging
- Update documentation

## [1.0.0] - 2024-09-01

### Added
- Basic functionality of Telegram TON Wallet Bot
- /start command to begin interaction with the bot
- /create_wallet command to create a new wallet
- /balance command to check wallet balance
- /send command to send TON
- /receive command to get address for top-up
- /help command to view available commands
- /history command to view transaction history

### Security
- Basic protection and encryption for wallet key storage