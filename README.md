# Crypto-DHT
Experimental Blockchain over DHT

## Jump To

- [Disclaimer](#disclaimer)
- [Background](#background)
- [Features](#features)
- [Details](#details)
- [Usage](#usage)
- [Build](#build)
  - [Setup](#setup)
  - [Bundle](#bundle)
- [Todo](#todo)

## Disclaimer

This software and the associated DHT are still Proof Of Concept, and still under development.

A lot of work is needed to reach a real-world usable state.

PR are welcomed !

## Background

Bitcoin is a quickly growing crypto-currency, gaining interest from the public
by its capacity to emit some digital money, to transfer that money between wallets,
to assure a certain anonymity in those transfers,
and all of that without intervention of any bank or any other third party. (other than part of the network, ofc)

Bitcoin based blockchains all share the same characteristic: In order to reach
a consensus, each and every nodes participating in the network have to keep a
full copy of the blockchain. Even if this issue has been solved with light wallets
and other trust-based protocol, a majority of the nodes need to keep a full copy
of the blockchain in order for the network to keep working well. At this time,
this blockchain is now 240GB big. (!) New arrivants have to wait at least one week
before starting to actualy use their wallet.

A Decentralized Hash Table (DHT) is a form of network used to store some content
in the form of key/value pairs. It differs from classical hash tables by its
decentralized and distributed nature. In fact, each node participating in a DHT can fetch and store
addressable content by key, as well as storing and serving a fraction of that hash table.

This experimental project try to avoid keeping all the blockchain, but instead
prefers to store it in a DHT. This way every node has to keep a small
portion of the blocks, depending on its own address in the DHT, the number of keys stored
and the number of nodes participating in the network.

Even if the blocks are stored in a decentralized way, every node still have to keep
some records when mining time has come:
- Every block's header, to validate a block when it comes for storage
- Every UTXOs (Unspent Transaction Output, basicaly forming the balances of
every wallet) in order to validate incoming transactions
- By extension every known address (wallet) that hold or have held some coins

Each of that implies that we have to fetch all the chain at least once, build what
we need from it, and throw it away. Then we just have to stay synced with new incoming blocks
to update our data on the fly.

## Features

Based on my own DHT implementation in GO: [go-dht](https://github.com/champii/go-dht)

- One block every minute
- Base block revenue is 1.00 coin (100 cents)
- DHT for block storage.

![Screenshot](https://github.com/champii/crypto-dht/raw/master/screenshot.png "Screenshot")
![Screenshot2](https://github.com/champii/crypto-dht/raw/master/screenshot2.png "Screenshot2")
![Screenshot3](https://github.com/champii/crypto-dht/raw/master/screenshot3.png "Screenshot3")



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

## Details

`What happend when a new node connects to the network ?`

Try to figure the network as a circle, with 2^127 possible points on it from
address 0x0 to 0xFFFF...
When a node connects to the network (via a bootstrap node), it choose a unique 160bit key as its address and to represent itself.

If the node does not have a wallet yet, one is created.

It then starts to ask the bootstrap node for its neighborhood, creating a routing table
with the other nodes it discovers along the way. It then starts to populate its
routing table further by asking for random values, again adding nodes on its way.

At this time, it starts to synchronise from the DHT by polling the next block (or the first if this is a new connection)

Its easy to find a block in the DHT:

Given a block `b1` and a hash function `H`, to find the address `k` of the next block
inside the DHT, we muse obtain the hash `h1` of that first block with

`h1 = H(b1)`.

As this hash is not evenly distributed (must be less than the current target), we
hash it again to obtain the address `k` of the next block:

`k = H(h1)`

We can start from the hash of the genesis block (that is fixed) and keep going
by hashing each block hash we got to get the next one etc, etc.

When the DHT answers a NOT_FOUND or an error, we stop synchronising.


## Build


### Setup

```
$> cd client
$> npm install
$> npm run build // You have to run this everytime the client change
$> cd ..
$> go build
```

### Bundle

```
$> go get -u github.com/asticode/go-astilectron-bundler/...
$> ./scripts/build.sh
```

The output binary will be in `./build/linux-amd64/crypto-dht`

## Todo

- handle forks
- Dont permanently sync but rather use broadcast to spread and listen to new blocks
- Get pending transactions from other nodes
- Fees
- Scrypt
- Better GUI
- Manage wallets
- Recheck blockchain
- Config file
- Daemon ?
- (Make DHT address the hash of the wallet? anonymity may be compromised, don't allow for multiple connexions with same address)
