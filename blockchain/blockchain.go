package blockchain

import (
	"sync"
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

type HistoryTx struct {
	Address   string `json:"address"`
	Timestamp int64  `json:"timestamp"`
	Amount    int    `json:"amount"`
}

type Blockchain struct {
	sync.RWMutex
	status                 int
	client                 *dht.Dht
	logger                 *logging.Logger
	options                BlockchainOptions
	lastBlockTargetChanged *Block
	lastBlock              *Block
	headers                []*BlockHeader
	baseTarget             []byte
	lastTarget             []byte
	wallets                map[string]*Wallet
	unspentTxOut           map[string][]*UnspentTxOut
	pendingTransactions    []Transaction
	miningBlock            *Block
	synced                 bool
	mustStop               bool
	stats                  *Stats
	running                bool
	history                []HistoryTx
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
	// target, _ := hex.DecodeString("000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

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
			this.Lock()
			defer this.Unlock()
			block := Block{}
			err := msgpack.Unmarshal(cmd.Data.(dht.StoreInst).Data.([]byte), &block)

			if err != nil {
				this.logger.Critical("ONSTORE Unmarshal error", err.Error())

				return false
			}

			if block.Header.Height == this.lastBlock.Header.Height + 1 {
				return block.Verify(this)
			} else if block.Header.Height <= this.lastBlock.Header.Height {
				return block.VerifyOld(this)
			} else {
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
	this.headers = append(this.headers, &originalBlock.Header)
	this.miningBlock = originalBlock
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
			if err := this.SendTo(this.options.Send); err != nil {
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

func (this *Blockchain) SendTo(value string) error {
	splited := strings.Split(value, ":")

	if len(splited) != 2 {
		return errors.New("Bad send format")
	}

	amount, err := strconv.Atoi(splited[0])

	if err != nil || amount <= 0 {
		return errors.New("Invalid amount: " + splited[0])
	}

	pub := UnsanitizePubKey(splited[1])

	tx := NewTransaction(amount, pub, this)

	if tx == nil || !tx.Verify(this) || this.hasPending(tx) {
		return errors.New("Unable to create the transaction")
	}

	if HasDoubleSpend(append(this.pendingTransactions, *tx)) {
		return errors.New("Created a double spending transaction !")
	}

	this.pendingTransactions = append(this.pendingTransactions, *tx)

	this.mustStop = true

	serie, err := msgpack.Marshal(tx)

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


func (this *Blockchain) hasPending(tx *Transaction) bool {
	for _, t := range this.pendingTransactions {
		if compare(t.GetHash(), tx.GetHash()) == 0 {
			return true
		}
	}

	return false
}

func (this *Blockchain) Dispatch(cmd dht.Packet) interface{} {
	switch cmd.Data.(dht.CustomCmd).Command {
	case COMMAND_CUSTOM_NEW_TRANSACTION:
		var tx Transaction

		msgpack.Unmarshal(cmd.Data.(dht.CustomCmd).Data.([]uint8), &tx)

		if !tx.Verify(this) || this.hasPending(&tx){
			return nil
		}

		if HasDoubleSpend(append(this.pendingTransactions, tx)) {
			return nil
		}

		this.pendingTransactions = append(this.pendingTransactions, tx)

		// this.mustStop = true

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
	this.Lock()
	this.headers = append(this.headers, &block.Header)
	this.Unlock()

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
			this.miningBlock = NewBlock(this)

			this.miningBlock.Mine(this.stats, &this.mustStop)

			if this.mustStop {
				this.mustStop = false

				ticker.Stop()
				this.Mine()
				return
			}

			this.logger.Info("Found block !", hex.EncodeToString(this.miningBlock.Header.Hash))

			serie, _ := msgpack.Marshal(this.miningBlock)

			_, nb, err := this.client.StoreAt(NewHash(this.lastBlock.Header.Hash), serie)

			if err != nil || nb == 0 {
				this.logger.Warning("ERROR STORING BLOCK IN THE DHT !", hex.EncodeToString(this.miningBlock.Header.Hash))

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

func (this *Blockchain) WaitingTransactionCount() int {
	return len(this.pendingTransactions)
}

func (this *Blockchain) GetOwnHistory() []HistoryTx {
	return this.history
}

func (this *Blockchain) GetOwnWaitingTx() []HistoryTx {
	res := []HistoryTx{}

	for _, tx := range this.pendingTransactions {
		txValue := 0

		own := false

		addr := tx.Stamp.Pub
		if compare(tx.Stamp.Pub, this.wallets["main.key"].pub) == 0 {
			own = true
		}

		for _, out := range tx.Outs {
			if own && compare(out.Address, this.wallets["main.key"].pub) != 0 {
				txValue -= out.Value
				addr = out.Address
			}

			if !own && compare(out.Address, this.wallets["main.key"].pub) == 0 {
				txValue += out.Value
			}

			if len(tx.Ins) == 0 && len(tx.Outs) == 1 {
				txValue = 0
			}
		}

		if txValue != 0 {
			res = append(res, HistoryTx{
				Address:   SanitizePubKey(addr),
				Timestamp: time.Now().Unix(),
				Amount:    txValue,
			})
		}
	}

	return res
}

func (this *Blockchain) ProcessingTransactionCount() int {
	if !this.running {
		return 0
	}

	return len(this.miningBlock.Transactions) - 1
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
