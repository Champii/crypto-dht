# crypto-dht
Experimental Blockchain over DHT

## Info

Based on my own DHT implementation in GO: [go-dht](https://github.com/champii/go-dht)

- One block every minute
- DHT for block storage.

![Screenshot](https://github.com/champii/crypto-dht/raw/master/screenshot.png "Screenshot")
![Screenshot2](https://github.com/champii/crypto-dht/raw/master/screenshot2.png "Screenshot2")


## Usage

```
NAME:
  Crypto-Dht - Experimental Blockchain over DHT

USAGE:
  crypto-dht [options]

VERSION:
  0.1.0

OPTIONS:
  -c value, --connect value  Connect to node ip:port. If not set, startup a bootstrap node.
  -l value, --listen value   Listening address and port (default: "0.0.0.0:3000")
  -f value, --folder value   Config Folder (default: "/home/champii/.crypto-dht")
  -s                         Stat mode
  -m                         Mine
  -w                         Show wallets and amount
  -g                         Deactivate GUI
  -S value, --send value     Send coins from main.key. Must be of the form 'amount:destAddress'
  -n nodes, --network nodes  Spawn X new nodes network. If -b is not specified, a new network is created. (default: 0)
  -v level, --verbose level  Verbose level, 0 for CRITICAL and 5 for DEBUG (default: 3)
  -h, --help                 Print help
  -V, --version              Print version
```

## Build

```
$> go get -u github.com/asticode/go-astilectron-bundler/...
$> ./scripts/build.sh
```

The output binary will be in `./build/linux-amd64/crypto-dht`

## Todo

- handle forks
- Dont permanently sync but rather use broadcast to spread and listen to new blocks
- Hash pub key to make addresses shorter
- (Make DHT address the hash of the wallet? anonymity may be compromised, don't allow for multiple connexions with same address)
- Get pending transactions from other nodes
- Fees
- Scrypt
- Better GUI
- Manage wallets
- Recheck blockchain
- Config file
- Daemon ?
