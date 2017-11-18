package blockchain

import (
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
	R   []byte
	S   []byte
	Pub []byte
}

type Transaction struct {
	Ins   []TxIn
	Outs  []TxOut
	Stamp Stamp
}

func (this *Transaction) GetHash() []byte {
	r := this.Stamp.R
	s := this.Stamp.S

	this.Stamp.R = []byte{}
	this.Stamp.S = []byte{}

	hash, _ := msgpack.Marshal(this)

	this.Stamp.R = r
	this.Stamp.S = s

	return hash
}

func (this *Transaction) Verify(bc *Blockchain) bool {
	r := this.Stamp.R
	s := this.Stamp.S

	this.Stamp.R = []byte{}
	this.Stamp.S = []byte{}

	hash, _ := msgpack.Marshal(this)

	blockPub, _ := pem.Decode(this.Stamp.Pub)

	if blockPub == nil {
		bc.logger.Error("Tx verify: Cannot decode pub signature from tx", string(this.Stamp.Pub))
		return false
	}

	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	var r_ big.Int
	// var r2 *big.Int
	r_.SetBytes(r)

	var s_ big.Int
	s_.SetBytes(s)

	if !ecdsa.Verify(publicKey, hash, &r_, &s_) {
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

		insTotal += prevUnspentOut.out.Value
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
		return nil
	}

	transac := &Transaction{
		Stamp: Stamp{
			Pub: bc.wallets["main.key"].pub,
		},
		Ins:  insRes,
		Outs: outRes,
	}

	hash, _ := msgpack.Marshal(transac)

	r, s, err := ecdsa.Sign(rand.Reader, bc.wallets["main.key"].key, hash)

	if err != nil {
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
		},
		Ins: []TxIn{},
		Outs: []TxOut{TxOut{
			Value:   100,
			Address: bc.wallets["main.key"].pub,
		}},
	}

	hash, _ := msgpack.Marshal(transac)

	r, s, err := ecdsa.Sign(rand.Reader, bc.wallets["main.key"].key, hash)

	if err != nil {
		return nil
	}

	transac.Stamp.R = r.Bytes()
	transac.Stamp.S = s.Bytes()

	return transac
}

func (this *Blockchain) RemovePendingTransaction(insTx []Transaction) {
	for _, inTx := range insTx {
		inTxHash := inTx.GetHash()

		idx := -1

		for i, tx := range this.pendingTransactions {
			if compare(inTxHash, tx.GetHash()) == 0 {
				idx = i
				break
			}
		}

		if idx > -1 {
			this.pendingTransactions = append(this.pendingTransactions[:idx], this.pendingTransactions[idx+1:]...)
		}
	}
}
