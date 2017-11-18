package blockchain

import (
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack"
)

type BlockHeader struct {
	Height     int64
	Hash       []byte
	PrecHash   []byte
	MerkelHash []byte
	Target     []byte
	Timestamp  int64
	Nonce      int64
}

type Block struct {
	Header       BlockHeader
	Transactions []Transaction
}

func NewBlock(bc *Blockchain) *Block {
	block := &Block{
		Header: BlockHeader{
			Height:    bc.lastBlock.Header.Height + 1,
			PrecHash:  bc.lastBlock.Header.Hash,
			Timestamp: time.Now().Unix(),
			Target:    bc.lastTarget,
			Hash:      []byte{},
		},
		Transactions: bc.pendingTransactions,
	}

	block.Transactions = append([]Transaction{*NewCoinBaseTransaction(bc)}, block.Transactions...)

	// todo: merkel tree

	return block
}

var originalBlock *Block

func OriginBlock(bc *Blockchain) *Block {
	if originalBlock == nil {
		originalBlock = &Block{
			Header: BlockHeader{
				Height:    0,
				PrecHash:  []byte{},
				Timestamp: 0,
				Target:    bc.lastTarget,
				Hash:      []byte{},
			},
			Transactions: []Transaction{},
		}

		hash, _ := msgpack.Marshal(&originalBlock.Header)

		originalBlock.Header.Hash = hash
	}

	return originalBlock
}

func (this *Block) Mine(stats *Stats, mustStop *bool) {
	tmp, _ := msgpack.Marshal(this.Header)
	newHash := NewHash(tmp)

	for !*mustStop && compare(newHash, this.Header.Target) >= 0 {
		this.Header.Nonce++
		this.Header.Timestamp = time.Now().Unix()

		var err error

		tmp, err = msgpack.Marshal(&this.Header)

		if err != nil {
			fmt.Println("ERROR", err)

			return
		}

		newHash = NewHash(tmp)

		stats.lastHashes++
	}

	this.Header.Hash = newHash
}

func (this *Block) VerifyOld(bc *Blockchain) bool {
	storedHeader := bc.headers[this.Header.Height]

	hash := this.Header.Hash
	this.Header.Hash = []byte{}

	tmp, _ := msgpack.Marshal(&this.Header)
	newHash := NewHash(tmp)

	this.Header.Hash = hash

	if compare(newHash, storedHeader.Hash) != 0 {
		bc.logger.Error("Block verify old: Hashes does not match with stored one")

		return false
	}

	if compare(newHash, hash) != 0 {
		bc.logger.Error("Block verify old: Hashes does not match")

		return false
	}

	// todo: check merkelTree

	if len(this.Transactions[0].Ins) > 0 || len(this.Transactions[0].Outs) != 1 {
		bc.logger.Error("Block verify: Bad coinbase transaction")
		return false
	}

	for _, tx := range this.Transactions {
		if !tx.Verify(bc) {
			bc.logger.Error("Block verify: Bad transaction")

			return false
		}
	}

	return true
}

func (this *Block) Verify(bc *Blockchain) bool {
	if this.Header.Height != bc.lastBlock.Header.Height+1 {
		bc.logger.Error("Block verify: Bad height")

		return false
	}

	hash := this.Header.Hash
	this.Header.Hash = []byte{}

	tmp, _ := msgpack.Marshal(&this.Header)
	newHash := NewHash(tmp)

	this.Header.Hash = hash

	if compare(newHash, hash) != 0 {
		bc.logger.Error("Block verify: Hashes does not match")

		return false
	}

	if compare(bc.lastBlock.Header.Hash, this.Header.PrecHash) != 0 {
		bc.logger.Error("Block verify: Bad previous hash")

		return false
	}

	if compare(bc.lastTarget, this.Header.Target) != 0 {
		bc.logger.Error("Block verify: Bad target")

		return false
	}

	// todo: check merkelTree

	if len(this.Transactions[0].Ins) > 0 || len(this.Transactions[0].Outs) != 1 {
		bc.logger.Error("Block verify: Bad coinbase transaction")
		return false
	}

	for _, tx := range this.Transactions {
		if !tx.Verify(bc) {
			bc.logger.Error("Block verify: Bad transaction")

			return false
		}
	}

	if HasDoubleSpend(this.Transactions) {
		bc.logger.Error("Block verify: Double spend")

		return false
	}

	return true
}

func HasDoubleSpend(transactions []Transaction) bool {
	seen := make(map[string]int)

	for _, tx := range transactions {
		for _, in := range tx.Ins {
			v, hasSeen := seen[string(in.PrevHash)]

			if hasSeen && v == in.PrevIdx {
				return true
			}

			seen[string(in.PrevHash)] = in.PrevIdx
		}
	}

	return false
}

func compare(b1, b2 []byte) int {
	if len(b1) > len(b2) {
		return 1
	} else if len(b2) > len(b1) {
		return -1
	}

	for i, v := range b1 {
		if v-b2[i] != 0 {
			return int(v) - int(b2[i])
		}
	}

	return 0
}
