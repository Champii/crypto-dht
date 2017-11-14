package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/vmihailenco/msgpack"
)

type Tx struct {
}

type TxIn struct {
	Tx
	PrevHash []byte
	PrevIdx  int
}

type TxOut struct {
	Tx
	Value  int
	Wallet []byte
}

type Stamp struct {
	R   *big.Int
	S   *big.Int
	Pub []byte
}

type Transaction struct {
	Ins   []TxIn
	Outs  []TxOut
	Stamp Stamp
}

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

func NewCoinBaseTransaction(value int, bc *Blockchain) *Transaction {
	transac := &Transaction{
		Outs: []TxOut{TxOut{
			Value:  value,
			Wallet: bc.wallets["main.key"].pub,
		}},
	}

	transac.Stamp.Pub = bc.wallets["main.key"].pub
	hash, _ := msgpack.Marshal(transac)

	r, s, err := ecdsa.Sign(rand.Reader, bc.wallets["main.key"].key, hash)

	if err != nil {
		return nil
	}

	transac.Stamp.R = r
	transac.Stamp.S = s

	return transac
}

func NewBlock(bc *Blockchain) *Block {
	block := &Block{
		Header: BlockHeader{
			PrecHash:  bc.lastBlockHash,
			Timestamp: time.Now().Unix(),
			Target:    bc.target,
		},
		Transactions: bc.pendingTransactions,
	}

	block.Transactions = append([]Transaction{*NewCoinBaseTransaction(1000000, bc)}, block.Transactions...)

	return block
}

var originalBlock *Block

func OriginBlock(bc *Blockchain) *Block {
	if originalBlock == nil {
		originalBlock = NewBlock(bc)
		originalBlock.Header.Timestamp = time.Date(1990, time.January, 0, 0, 0, 0, 0, time.Local).Unix()
		tmp, _ := msgpack.Marshal(&originalBlock.Header)

		originalBlock.Header.Hash = tmp
	}

	return originalBlock
}

func (this *Block) Mine() {
	tmp, _ := msgpack.Marshal(this.Header)
	newHash := NewHash(tmp)

	for compare(newHash, this.Header.Target) >= 0 {
		this.Header.Nonce++
		tmp, _ = msgpack.Marshal(this.Header)
		newHash = NewHash(tmp)
	}

	this.Header.Hash = newHash

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
