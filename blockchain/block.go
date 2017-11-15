package blockchain

import (
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack"
)

type BlockHeader struct {
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
			PrecHash:  bc.lastBlockHash,
			Timestamp: time.Now().UnixNano(),
			Target:    bc.target,
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
		originalBlock = NewBlock(bc)

		originalBlock.Header.Timestamp = time.Date(1990, time.January, 0, 0, 0, 0, 0, time.Local).Unix()

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

func (this *Block) Verify(bc *Blockchain) bool {
	hash := this.Header.Hash
	this.Header.Hash = []byte{}

	tmp, _ := msgpack.Marshal(&this.Header)
	newHash := NewHash(tmp)

	this.Header.Hash = hash

	if compare(newHash, hash) != 0 {
		bc.logger.Error("Block verify: Hashes does not match")

		return false
	}

	if compare(bc.lastBlockHash, this.Header.PrecHash) != 0 {
		bc.logger.Error("Block verify: Bad previous hash")

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
