package blockchain

import (
	"time"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"

	"github.com/vmihailenco/msgpack"
)

type TxIn struct {
	PrevHash []byte
	PrevIdx  int
}

type TxOut struct {
	Value   int
	Address []byte
}

type Stamp struct {
	R         []byte
	S         []byte
	Pub       []byte
	Hash      []byte
	Timestamp int64
}

type Transaction struct {
	Ins       []TxIn
	Outs      []TxOut
	Stamp 	  Stamp
}

func (this *Transaction) Verify(bc *Blockchain) bool {
	r := this.Stamp.R
	s := this.Stamp.S
	txHash := this.Stamp.Hash

	this.Stamp.R = []byte{}
	this.Stamp.S = []byte{}
	this.Stamp.Hash = []byte{}

	hash, err := msgpack.Marshal(this)
	
	if err != nil {
		bc.logger.Error("Tx verify: Cannot marshal the tx")
		
		return false
	}
		
	newHash := NewHash(hash)
	
	this.Stamp.Hash = txHash

	if compare(newHash, this.Stamp.Hash) != 0 {
		bc.logger.Error("Tx verify: Hash dont match", newHash)

		return false
	}

	blockPub, _ := pem.Decode(this.Stamp.Pub)

	if blockPub == nil {
		bc.logger.Error("Tx verify: Cannot decode pub signature from tx", 
                    string(this.Stamp.Pub))

		return false
	}

	// verify timestamp

	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	var r_ big.Int
	// var r2 *big.Int
	r_.SetBytes(r)

	var s_ big.Int
	s_.SetBytes(s)

	if !ecdsa.Verify(publicKey, newHash, &r_, &s_) {
		bc.logger.Error("Tx verify: Signatures does not match")

		return false
	}

	this.Stamp.R = r
	this.Stamp.S = s
	

	// lets assume this will work any time
	if len(this.Ins) == 0 && len(this.Outs) == 1 {
		if this.Outs[0].Value != 100 {
			bc.logger.Error("Tx verify: Bad coinbase amount")

			return false
		}

		return true
	}

	insTotal := 0
	for _, in := range this.Ins {
		prevUnspentOut := bc.getCorrespondingOutTx(this.Stamp.Pub, &in)

		if prevUnspentOut == nil {
			bc.logger.Error("Tx verify: Cannot find corresponding OutTx for given In")

			return false
		}

		insTotal += prevUnspentOut.Out.Value
	}

	outsTotal := 0
	for _, out := range this.Outs {
		outsTotal += out.Value
	}

	if outsTotal > insTotal {
		bc.logger.Error("Tx verify: Outs total amount exceeds in amount")

		return false
	}

	return true
}

func NewTransaction(value int, dest []byte, bc *Blockchain) *Transaction {
	outs := bc.GetEnoughOwnUnspentOut(value)

	insRes, outRes := bc.GetInOutFromUnspent(value, dest, outs)

	if len(outs) == 0 {
		bc.logger.Warning("Cannot create transaction: no outs")

		return nil
	}

	transac := &Transaction{
		Stamp: Stamp{
			Pub: bc.wallets["main.key"].pub,
			Timestamp: time.Now().Unix(),
			Hash: []byte{},
			R: []byte{},
			S: []byte{},
		},
		Ins:  insRes,
		Outs: outRes,
	}
	
	hash, err := msgpack.Marshal(transac)
	
	if err != nil {
		bc.logger.Warning("Cannot marshal the transaction", err)
		
		return nil
	}
	
	newHash := NewHash(hash)
	transac.Stamp.Hash = newHash
	
	r, s, err := ecdsa.Sign(rand.Reader, bc.wallets["main.key"].key, newHash)
	
	if err != nil {
		bc.logger.Warning("Cannot create transaction: Signature error", err)
		
		return nil
	}
	
	transac.Stamp.R = r.Bytes()
	transac.Stamp.S = s.Bytes()
	
	return transac
}

func NewCoinBaseTransaction(bc *Blockchain) *Transaction {
	transac := &Transaction{
		Stamp: Stamp{
			Pub: bc.wallets["main.key"].pub,
			Timestamp: time.Now().Unix(),
			Hash: []byte{},
			R: []byte{},
			S: []byte{},
		},
		Ins: []TxIn{},
		Outs: []TxOut{TxOut{
			Value:   100,
			Address: []byte(SanitizePubKey(bc.wallets["main.key"].pub)),
		}},
	}
	
	hash, err := msgpack.Marshal(transac)
	
	if err != nil {
		bc.logger.Warning("Cannot marshal the transaction", err)
		
		return nil
	}

	newHash := NewHash(hash)
	
	transac.Stamp.Hash = newHash
	
	r, s, err := ecdsa.Sign(rand.Reader, bc.wallets["main.key"].key, newHash)
	
	if err != nil {
		return nil
	}

	transac.Stamp.R = r.Bytes()
	transac.Stamp.S = s.Bytes()

	return transac
}

func (this *Blockchain) RemovePendingTransaction(insTx []Transaction) {
	for _, inTx := range insTx {
		inTxHash := inTx.Stamp.Hash

		idx := -1

		for i, tx := range this.pendingTransactions {
			if compare(inTxHash, tx.Stamp.Hash) == 0 {
				idx = i
				break
			}
		}

		if idx > -1 {
			this.pendingTransactions = append(this.pendingTransactions[:idx], 
																				this.pendingTransactions[idx+1:]...)
		}
	}
}

func (this *Blockchain) AddTransationToWaiting(tx *Transaction) bool {
	if !tx.Verify(this) || this.hasPending(tx){
		this.logger.Warning("Cannot add transaction to waiting")

		return false
	}

	if HasDoubleSpend(append(this.pendingTransactions, *tx)) {
		this.logger.Warning("Cannot add transaction to waiting: Has double spend")

		return false
	}

	outs := []*UnspentTxOut{}
	for _, in := range tx.Ins {
		out := this.getCorrespondingOutTx(tx.Stamp.Pub, &in)
		outs = append(outs, out)

		if out == nil {
			this.logger.Warning("Cannot find corresponding out")

			return false
		}

		if out.IsTargeted {
			this.logger.Warning("Got transaction with double spending")

			return false
		}
	}

	for _, out := range outs {
		out.IsTargeted = true
	}

	this.pendingTransactions = append(this.pendingTransactions, *tx)

	return true
}