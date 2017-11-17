package blockchain

import (
	"encoding/hex"
	"errors"
	"math/big"
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

var EXPECTED_10_BLOCKS_TIME int64 = 600

type UnspentTxOut struct {
	out    TxOut
	txHash []byte
	inIdx  int
}

type Blockchain struct {
	status                 int
	client                 *dht.Dht
	logger                 *logging.Logger
	options                BlockchainOptions
	lastBlockTargetChanged *Block
	lastBlock              *Block
	baseTarget             []byte
	lastTarget             []byte
	wallets                map[string]*Wallet
	unspentTxOut           map[string][]*UnspentTxOut
	pendingTransactions    []Transaction
	synced                 bool
	mustStop               bool
	stats                  *Stats
	running                bool
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
	target, _ := hex.DecodeString("000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	if options.Stats {
		options.Verbose = 2
	}

	bc := &Blockchain{
		status:              0,
		options:             options,
		baseTarget:          target,
		lastTarget:          target,
		wallets:             make(map[string]*Wallet),
		unspentTxOut:        make(map[string][]*UnspentTxOut),
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
		OnStore: func(cmd dht.Packet) bool {
			data := cmd.Data.(dht.StoreInst).Data
			switch data.(type) {
			case Block:
				block := data.(Block)
				return this.AddBlock(&block)
			default:
				return false
			}
		},

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

	this.lastBlockTargetChanged = originalBlock
	this.lastBlock = originalBlock
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
		// var block Block

		// msgpack.Unmarshal(cmd.Data.(dht.CustomCmd).Data.([]uint8), &block)

		// if !this.AddBlock(&block) {
		// 	return nil
		// }

		// this.mustStop = true
	}

	return nil
}

func (this *Blockchain) Wait() {
	this.client.Wait()
}

func (this *Blockchain) Sync() {
	var lastErr error

	for lastErr == nil {
		block_, err := this.client.Fetch(NewHash(this.lastBlock.Header.Hash))

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

	go func() {
		for {
			block_, err := this.client.Fetch(NewHash(this.lastBlock.Header.Hash))

			if err != nil {
				time.Sleep(time.Second * 5)
				continue
			}

			var block Block

			msgpack.Unmarshal(block_.([]uint8), &block)

			if !this.AddBlock(&block) {
				this.logger.Warning("Sync: Received bad block")

				return
			}

			this.mustStop = true

			time.Sleep(time.Second * 5)
		}
	}()
}

func (this *Blockchain) AddBlock(block *Block) bool {
	if !block.Verify(this) {
		this.logger.Error("Cannot add block: bad block")

		return false
	}

	if compare(block.Header.PrecHash, originalBlock.Header.Hash) == 0 {
		this.lastBlockTargetChanged = block
	}

	this.lastBlock = block

	this.UpdateUnspentTxOuts(block)
	this.RemovePendingTransaction(block.Transactions)

	if block.Header.Height%10 == 0 {
		this.adjustDifficulty(block)
	}

	return true
}

func (this *Blockchain) adjustDifficulty(block *Block) {
	base := big.NewInt(0)
	actual := big.NewInt(0)
	base.SetString(hex.EncodeToString(this.baseTarget), 16)
	actual.SetString(hex.EncodeToString(block.Header.Target), 16)

	oldDiff := big.NewInt(0)
	oldDiff = oldDiff.Quo(base, actual)

	timePassed := block.Header.Timestamp - this.lastBlockTargetChanged.Header.Timestamp

	newDiff := big.NewInt(0)
	newDiff = newDiff.Mul(oldDiff, big.NewInt(EXPECTED_10_BLOCKS_TIME/timePassed))

	test := big.NewInt(0)
	if newDiff.Int64() > test.Mul(oldDiff, big.NewInt(4)).Int64() {
		newDiff = test
	}

	test = big.NewInt(0)
	if newDiff.Int64() < test.Quo(oldDiff, big.NewInt(4)).Int64() {
		newDiff = test
	}

	if newDiff.Int64() < 1 {
		newDiff = big.NewInt(1)
	}

	test = big.NewInt(0)
	this.lastTarget = test.Quo(base, newDiff).Bytes()

	for len(this.lastTarget) < len(this.baseTarget) {
		this.lastTarget = append([]byte{0}, this.lastTarget...)
	}

	this.lastBlockTargetChanged = block
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

			this.logger.Info("Found block !", hex.EncodeToString(block.Header.Hash))

			serie, _ := msgpack.Marshal(&block)

			_, nb, err := this.client.StoreAt(NewHash(this.lastBlock.Header.Hash), serie)

			if err != nil || nb == 0 {
				this.logger.Warning("ERROR STORING BLOCK IN THE DHT !", hex.EncodeToString(block.Header.Hash))

				continue

			}

			this.stats.foundBlocks++
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

func (this *Blockchain) BlocksHeight() int64 {
	return this.lastBlock.Header.Height
}

func (this *Blockchain) TimeSinceLastBlock() int64 {
	return time.Now().Unix() - this.lastBlock.Header.Timestamp
}

func (this *Blockchain) StoredKeys() int {
	return this.client.StoredKeys()
}

func (this *Blockchain) Difficulty() int64 {
	base := big.NewInt(0)
	actual := big.NewInt(0)
	base.SetString(hex.EncodeToString(this.baseTarget), 16)
	actual.SetString(hex.EncodeToString(this.lastTarget), 16)

	return base.Quo(base, actual).Int64()
}

func (this *Blockchain) NextDifficulty() int64 {
	base := big.NewInt(0)
	actual := big.NewInt(0)
	base.SetString(hex.EncodeToString(this.baseTarget), 16)
	actual.SetString(hex.EncodeToString(this.lastTarget), 16)

	oldDiff := big.NewInt(0)
	oldDiff = oldDiff.Quo(base, actual)

	timePassed := (time.Now().Unix() - this.lastBlockTargetChanged.Header.Timestamp)

	if timePassed == 0 {
		return oldDiff.Int64()
	}

	nbBlocks := this.lastBlock.Header.Height - this.lastBlockTargetChanged.Header.Height + 1

	if nbBlocks == 0 {
		nbBlocks = 1
	}

	timePassed = (timePassed / nbBlocks) * 10

	if timePassed == 0 {
		timePassed = 1
	}

	newDiff := big.NewInt(0)
	newDiff = newDiff.Mul(oldDiff, big.NewInt((EXPECTED_10_BLOCKS_TIME / timePassed)))

	test := big.NewInt(0)
	if newDiff.Int64() > test.Mul(oldDiff, big.NewInt(4)).Int64() {
		newDiff = test
	}

	test = big.NewInt(0)
	if newDiff.Int64() < test.Quo(oldDiff, big.NewInt(4)).Int64() {
		newDiff = test
	}

	if newDiff.Int64() < 1 {
		newDiff = big.NewInt(1)
	}

	return newDiff.Int64()
}
