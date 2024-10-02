# TODO List for Improving Telegram TON Wallet Bot

## 1. Wallet Management
- [x] Limit wallet creation to one per user
  - [x] Modify the logic of /create_wallet command
  - [x] Add existing wallet check
  - [x] Implement display of existing wallet info on repeated requests
- [ ] Develop a wallet backup mechanism
- [ ] Implement wallet recovery function

## 2. Checks and Validation
- [ ] Add wallet existence check before executing commands
- [ ] Improve input error handling
  - [ ] Add validation for address format and amount when sending
  - [ ] Implement informative error messages
- [ ] Implement sufficient balance check before sending

## 3. Improving Transaction Process
- [ ] Add transaction confirmation step
- [ ] Add calculation and display of transaction fees
- [ ] Implement transaction status tracking function
- [ ] Add user notifications about transaction status

## 4. Blockchain Interaction Optimization
- [ ] Implement balance caching
  - [ ] Add periodic updating of cached balance
  - [ ] Optimize blockchain queries

## 5. Security
- [ ] Add additional security measures (e.g., action confirmation)
- [ ] Implement rate limiting for command usage
- [ ] Add system of temporary blocks for suspicious activity

## 6. Improving User Experience
- [ ] Add more detailed command descriptions in /help
- [ ] Implement interactive buttons for frequently used functions
- [ ] Add ability to configure notifications (e.g., for large transactions)

## 7. Expanding Functionality
- [ ] Add support for multiple currencies (if applicable)
- [ ] Implement currency exchange function (if applicable)
- [ ] Add display of TON price history

## 8. Testing and Debugging
- [ ] Develop a comprehensive test suite for all functions
- [ ] Conduct load testing
- [ ] Implement a logging system to track errors and user behavior

## 9. Documentation
- [ ] Update documentation for API and internal project structure
- [ ] Create a user guide describing all bot functions
- [ ] Create documentation on log interpretation for system operators
- [ ] Document all error codes and their meanings

## 10. Logging and Error Handling
- [ ] Implement structured logging (e.g., using logrus or zap)
- [ ] Implement usage of different logging levels (DEBUG, INFO, WARN, ERROR)
- [ ] Add contextual information to logs (operation ID, timestamp)
- [ ] Implement masking of sensitive data in logs
- [ ] Add ability to configure logging verbosity level through configuration
- [ ] Implement a centralized error handler
- [ ] Add stack trace for critical errors
- [ ] Add logging of execution time for critical operations
- [ ] Implement logging of resource usage (memory, CPU)
- [ ] Add logging of SQL queries in debug mode
- [ ] Improve logging of the transaction sending process
- [ ] Add logging of configuration loading during application startup
- [ ] Add clear logs about application start and stop

## 11. Database
- [ ] Fix logging of wallets table indexes

## 12. Application Lifecycle
- [ ] Log the graceful shutdown process