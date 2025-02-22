# SKATEPAY HUB

This hub is used to monitor and dispatch SKATEPAY related [nostr][nostr] and [solana][solana] events. 

## Manual
### Start hub or run utility command
```
go run main.go up
go run main.go dm-bot
go run main.go group-bot
```

### Refactor assist
```
brew install tree
tree --prune -I "$(paste -sd'|' .treeignore)" > project-structure.txt
```

[nostr]: https://github.com/fiatjaf/nostr
[solana]: https://docs.solanalabs.com/cli/install

## References
- [skatepay](https://github.com/SkatePay/skatepay) - [MIT License, Copyright (c) 2025 SKATEPAY.CHAT](https://github.com/SkatePay/skatepay/blob/main/LICENSE)
- [prorobot](https://prorobot.ai)

## Acknowledgements
- [go-nostr](https://github.com/nbd-wtf/go-nostr) - [MIT License, Copyright (c) 2022 nbd](https://github.com/nbd-wtf/go-nostr/blob/master/LICENSE.md)
- [solana-go](https://github.com/gagliardetto/solana-go) - [Apache License 2.0](https://github.com/gagliardetto/solana-go/blob/main/LICENSE)
- [octane](https://github.com/anza-xyz/octane) - [Apache License 2.0](https://github.com/anza-xyz/octane/blob/master/LICENSE)
- [nostream](https://github.com/cameri/nostream) - [The MIT License (MIT)](https://github.com/cameri/nostream/blob/main/LICENSE)
