# TODO List for improving Telegram TON Wallet Bot

## 1. Wallet Management
  - [ ] Limit wallet creation to one per user
  - [ ] Modify the logic of the /create_wallet command
  - [ ] Add a check for existing wallet
  - [ ] Implement display of existing wallet information on repeated request

## 2. Checks and Validation
- [ ] Add wallet existence check before executing commands
  - [ ] Implement for /balance
  - [ ] Implement for /send
  - [ ] Implement for /receive
  - [ ] Implement for /history
- [ ] Improve input error handling
  - [ ] Add validation for address format and amount when sending
  - [ ] Implement informative error messages

## 3. Improving Transaction Sending Process
- [ ] Add transaction confirmation step
- [ ] Implement balance sufficiency check before sending
- [ ] Add calculation and display of transaction fees
- [ ] Implement transaction status tracking function
- [ ] Add user notifications about transaction status

## 4. Optimizing Blockchain Interaction
- [ ] Implement balance caching
  - [ ] Add periodic updating of cached balance
  - [ ] Optimize blockchain queries

## 5. Security and Recovery
- [ ] Develop wallet backup mechanism
- [ ] Implement wallet recovery function
- [ ] Add additional security measures (e.g., action confirmation)

## 6. Spam and Abuse Protection
- [ ] Implement rate limiting for command usage
- [ ] Add system of temporary blocks for suspicious activity

## 7. Improving User Experience
- [ ] Add more detailed command descriptions in /help
- [ ] Implement interactive buttons for frequently used functions
- [ ] Add ability to customize notifications (e.g., for large transactions)

## 8. Expanding Functionality
- [ ] Add support for multiple currencies (if applicable)
- [ ] Implement currency exchange function (if applicable)
- [ ] Add display of TON price history

## 9. Testing and Debugging
- [ ] Develop comprehensive test suite for all functions
- [ ] Conduct load testing
- [ ] Implement logging system for tracking errors and user behavior

## 10. Documentation
- [ ] Update documentation for API and internal project structure
- [ ] Create user guide with description of all bot functions