# SKATEPAY HUB

## Start
    go mod init
    go mod tidy
    go run main.go

## Test Checklist
### App Invite
    - [ ] Start hub subscriber
    - [ ] Copy npub
    - [ ] Send npub to prorobot via Primal or Damus 
    - [ ] Send direct message to npub via SkatePay

### Publisher ping DM
    - [ ] Obtain npub on Primal
    - [ ] Modify publisher.go and run project
    - [ ] Check messages in Primal
    
## References
- [skatepay](https://github.com/SkatePay/skatepay) - [MIT License, Copyright (c) 2024 SKATEPAY.CHAT](https://github.com/SkatePay/skatepay/blob/main/LICENSE)
- [prorobot](https://prorobot.ai)

## Acknowledgements
- [go-nostr](https://github.com/nbd-wtf/go-nostr) - [MIT License, Copyright (c) 2022 nbd](https://github.com/nbd-wtf/go-nostr/blob/master/LICENSE.md)