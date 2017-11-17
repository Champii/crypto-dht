package blockchain

import (
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/champii/go-dht/dht"
	logging "github.com/op/go-logging"
	"github.com/vmihailenco/msgpack"
)

const (
	COMMAND_CUSTOM_GET_INFO = iota
	COMMAND_CUSTOM_NEW_TRANSACTION
	COMMAND_CUSTOM_NEW_BLOCK
)

type UnspentTxOut struct {
	out    TxOut
	txHash []byte
	inIdx  int
}

type Blockchain struct {
	status              int
	client              *dht.Dht
	logger              *logging.Logger
	options             BlockchainOptions
	target              []byte
	wallets             map[string]*Wallet
	unspentTxOut        map[string][]*UnspentTxOut
	pendingTransactions []Transaction
	lastBlockHash       []byte
	blocksHeight        int
	synced              bool
	mustStop            bool
	stats               *Stats
	running             bool
}

type BlockchainOptions struct {
	BootstrapAddr string
	ListenAddr    string
	Folder        string
	Send          string
	Interactif    bool
	Wallets       bool
	Stats         bool
	Verbose       int
	Mine          bool
	NoGui         bool
}

func New(options BlockchainOptions) *Blockchain {
	target, _ := hex.DecodeString("000009FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	if options.Stats {
		options.Verbose = 2
	}

	bc := &Blockchain{
		status:              0,
		options:             options,
		target:              target,
		wallets:             make(map[string]*Wallet),
		unspentTxOut:        make(map[string][]*UnspentTxOut),
		blocksHeight:        1,
		mustStop:            false,
		stats:               &Stats{},
		pendingTransactions: []Transaction{},
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

	if this.options.Stats {
		go this.StatsLoop()
	}

	go func() {
		this.Sync()

		if !this.synced {
			this.logger.Error("Unable to sync")

			return
		}

		if this.options.Wallets {
			this.ShowWallets()

		}

		if len(this.options.Send) > 0 {
			if err := this.SendTo(); err != nil {
				this.logger.Error("Unable to Send", err)

				return
			}

			time.Sleep(time.Second * 5)
			os.Exit(0)
		}

		if this.options.Mine {
			this.Mine()
		}
	}()

	return nil
}

func (this *Blockchain) SendTo() error {
	splited := strings.Split(this.options.Send, ":")

	if len(splited) != 2 {
		return errors.New("Bad send format")
	}

	amount, err := strconv.Atoi(splited[0])

	if err != nil || amount <= 0 {
		return errors.New("Invalid amount: " + splited[0])
	}

	// _, ok := this.unspentTxOut[splited[1]]

	// if !ok {
	// 	return errors.New("Unknown dest address: " + splited[1])
	// }

	tx := NewTransaction(amount, []byte(splited[1]), this)

	if tx == nil {
		return errors.New("Unable to create the transaction")
	}

	serie, err := msgpack.Marshal(&tx)

	if err != nil {
		return errors.New("Cannot marshal transaction: " + err.Error())
	}

	this.client.Broadcast(dht.CustomCmd{
		Command: COMMAND_CUSTOM_NEW_TRANSACTION,
		Data:    serie,
	})

	return nil
}

func (this *Blockchain) Logger() *logging.Logger {
	return this.logger
}

func (this *Blockchain) Dispatch(cmd dht.Packet) interface{} {
	switch cmd.Data.(dht.CustomCmd).Command {
	case COMMAND_CUSTOM_NEW_TRANSACTION:
		var tx Transaction

		msgpack.Unmarshal(cmd.Data.(dht.CustomCmd).Data.([]uint8), &tx)

		if !tx.Verify(this) {
			return nil
		}

		this.pendingTransactions = append(this.pendingTransactions, tx)

		this.mustStop = true

	case COMMAND_CUSTOM_NEW_BLOCK:
		var block Block

		msgpack.Unmarshal(cmd.Data.(dht.CustomCmd).Data.([]uint8), &block)

		if !this.AddBlock(&block) {
			return nil
		}

		this.mustStop = true
	}

	return nil
}

func (this *Blockchain) Wait() {
	this.client.Wait()
}

func (this *Blockchain) Sync() {
	var lastErr error

	for lastErr == nil {
		block_, err := this.client.Fetch(NewHash(this.lastBlockHash))

		lastErr = err
		if err == nil {
			var block Block

			msgpack.Unmarshal(block_.([]uint8), &block)

			if !this.AddBlock(&block) {
				this.logger.Warning("Sync: Received bad block")

				return
			}
		}
	}
	this.synced = true
}

func (this *Blockchain) AddBlock(block *Block) bool {
	if !block.Verify(this) {
		this.logger.Error("Cannot add block: bad block")

		return false
	}

	this.lastBlockHash = block.Header.Hash

	this.blocksHeight++

	this.UpdateUnspentTxOuts(block)
	this.RemovePendingTransaction(block.Transactions)

	return true
}

func (this *Blockchain) Mine() {
	ticker := time.NewTicker(time.Second)

	go func() {
		for range ticker.C {
			this.stats.Update()
		}
	}()

	this.running = true

	go func() {
		for this.running {
			block := NewBlock(this)

			block.Mine(this.stats, &this.mustStop)

			if this.mustStop {
				this.mustStop = false

				ticker.Stop()
				this.Mine()
				return
			}

			lastBlockHash := NewHash(this.lastBlockHash)

			if !this.AddBlock(block) {
				this.logger.Error("OWN MINED BLOCK IS DEFEICTIVE !", hex.EncodeToString(block.Header.Hash))

				continue
			}

			this.logger.Info("Found block !", hex.EncodeToString(block.Header.Hash))

			this.stats.foundBlocks++

			serie, _ := msgpack.Marshal(&block)

			this.client.StoreAt(lastBlockHash, serie)

			this.client.Broadcast(dht.CustomCmd{
				Command: COMMAND_CUSTOM_NEW_BLOCK,
				Data:    serie,
			})
		}
	}()
}

func (this *Blockchain) Wallets() map[string]*Wallet {
	return this.wallets
}

func (this *Blockchain) Synced() bool {
	return this.synced
}

func (this *Blockchain) Running() bool {
	return this.running
}

func (this *Blockchain) Stats() *Stats {
	return this.stats
}

func (this *Blockchain) GetConnectedNodesNb() int {
	return this.client.GetConnectedNumber()
}

func (this *Blockchain) BlocksHeight() int {
	return this.blocksHeight
}
