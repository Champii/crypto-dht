package main

import (
	"encoding/hex"
	"fmt"

	"github.com/champii/go-dht/dht"
	logging "github.com/op/go-logging"
	"github.com/vmihailenco/msgpack"
)

const (
	COMMAND_CUSTOM_GET_INFO = iota
	COMMAND_CUSTOM_NEW_TRANSACTION
	COMMAND_CUSTOM_NEW_BLOCK
)

type Blockchain struct {
	status              int
	client              *dht.Dht
	logger              *logging.Logger
	options             BlockchainOptions
	target              []byte
	wallets             map[string]*Wallet
	pendingTransactions []Transaction
	lastBlockHash       []byte
	synced              bool
}

type BlockchainOptions struct {
	BootstrapAddr string
	ListenAddr    string
	Folder        string
	Interactif    bool
	Stats         bool
	Verbose       int
	Mine          bool
}

func New(options BlockchainOptions) *Blockchain {
	target, _ := hex.DecodeString("00000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	bc := &Blockchain{
		status:  0,
		options: options,
		target:  target,
		wallets: make(map[string]*Wallet),
	}

	bc.Init()

	return bc
}

func (this *Blockchain) Init() {
	client := dht.New(dht.DhtOptions{
		ListenAddr:    this.options.ListenAddr,
		BootstrapAddr: this.options.BootstrapAddr,
		Verbose:       this.options.Verbose,
		OnCustomCmd: func(cmd dht.Packet) interface{} {
			return this.Dispatch(cmd)
		},

		OnBroadcast: func(packet dht.Packet) interface{} {
			if this.synced {
				return this.Dispatch(packet)
			}

			return nil
		},
	})

	this.client = client
	this.logger = client.Logger()

	if err := SetupStorage(this); err != nil {
		this.logger.Critical(err)

		return
	}

	OriginBlock(this)

	this.lastBlockHash = originalBlock.Header.Hash
}

func (this *Blockchain) Start() error {
	if err := this.client.Start(); err != nil {
		return err
	}

	// this.client.Broadcast(dht.CustomCmd{
	// 	Command: COMMAND_CUSTOM_NEW_TRANSACTION,
	// 	Data:    nil,
	// })

	this.Sync()
	if this.options.Mine {
		this.Mine()
	}

	return nil
}

func (this *Blockchain) Logger() *logging.Logger {
	return this.logger
}

func (this *Blockchain) Dispatch(cmd dht.Packet) interface{} {
	switch cmd.Data.(dht.CustomCmd).Command {
	case COMMAND_CUSTOM_NEW_TRANSACTION:
		this.logger.Info("New Transaction", cmd)
	case COMMAND_CUSTOM_NEW_BLOCK:
		var block Block

		msgpack.Unmarshal(cmd.Data.(dht.CustomCmd).Data.([]uint8), &block)
		this.logger.Info("New Block", block.Header.Hash)

		this.lastBlockHash = block.Header.Hash
	}

	return nil
}

func (this *Blockchain) Wait() {
	this.client.Wait()
}

func (this *Blockchain) Sync() {
	var lastErr error

	for lastErr == nil {
		block_, err := this.client.Fetch(hex.EncodeToString(NewHash(this.lastBlockHash)))

		lastErr = err
		if err == nil {
			var block Block

			msgpack.Unmarshal(block_.([]uint8), &block)
			fmt.Println("Sync block", block.Header.Hash)
			this.lastBlockHash = block.Header.Hash
		}
	}
	this.synced = true
}

func (this *Blockchain) Mine() {
	go func() {
		for {
			block := NewBlock(this)

			block.Mine()

			fmt.Println("NEW BLOCK")

			serie, _ := msgpack.Marshal(block)

			this.client.StoreAt(hex.EncodeToString(NewHash(this.lastBlockHash)), serie)

			this.lastBlockHash = block.Header.Hash

			this.client.Broadcast(dht.CustomCmd{
				Command: COMMAND_CUSTOM_NEW_BLOCK,
				Data:    serie,
			})

		}
	}()
}
