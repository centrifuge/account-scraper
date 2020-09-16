# account-scraper
Centrifuge Chain Account Scraper

## Description
This tool processes events in blocks by range. Looking for `Balances.Endowed` ones.

## How to run
Install deps and build binary in GOPATH
```
go install ./...
```

```
scraper --url wss://fullnode-archive.centrifuge.io
```


