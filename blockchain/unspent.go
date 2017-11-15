package blockchain

func (this *Blockchain) getUnspentTxOut() {

}

func (this *Blockchain) getCorrespondingOutTx(wallet []byte, in *TxIn) *UnspentTxOut {
	walletStr := SanitizePubKey(wallet)

	outs, ok := this.unspentTxOut[walletStr]

	if !ok {
		return nil
	}

	for i, out := range outs {
		if out.inIdx == in.PrevIdx && compare(out.txHash, in.PrevHash) == 0 {
			return this.unspentTxOut[walletStr][i]
		}
	}

	return nil
}

func (this *Blockchain) UpdateUnspentTxOuts(block *Block) {
	for _, tx := range block.Transactions {
		hash := tx.GetHash()

		for _, in := range tx.Ins {
			out := this.getCorrespondingOutTx(tx.Stamp.Pub, &in)

			if out == nil {
				this.logger.Critical("WARNING !!!!! IMPOSSIBLE TO FIND UNSPENT TX OUT FROM APPARENTLY VALID BLOCK")

				return
			}

			this.RemoveUnspentOut(tx.Stamp.Pub, out)
		}

		for i, out := range tx.Outs {
			walletStr := SanitizePubKey(out.Address)

			this.unspentTxOut[walletStr] = append(this.unspentTxOut[walletStr], &UnspentTxOut{
				out:    out,
				inIdx:  i,
				txHash: hash,
			})
		}
	}
}

func (this *Blockchain) RemoveUnspentOut(wallet []byte, out *UnspentTxOut) {
	walletStr := SanitizePubKey(wallet)
	idx := -1

	for i := range this.unspentTxOut[walletStr] {
		if this.unspentTxOut[walletStr][i] == out {
			idx = i
		}
	}

	if idx == -1 {
		this.logger.Critical("WARNING !!!!! IMPOSSIBLE REMOVEUNSPENT TX OUT")

		return
	}

	this.unspentTxOut[walletStr] = append(this.unspentTxOut[walletStr][:idx], this.unspentTxOut[walletStr][idx+1:]...)
}

func (this *Blockchain) GetEnoughOwnUnspentOut(value int) []*UnspentTxOut {
	walletStr := SanitizePubKey(this.wallets["main.key"].pub)

	var res []*UnspentTxOut

	total := 0
	for _, unspent := range this.unspentTxOut[walletStr] {
		total += unspent.out.Value

		res = append(res, unspent)

		if total > value {
			break
		}
	}

	if total < value {
		return []*UnspentTxOut{}
	}

	return res
}

func (this *Blockchain) GetAvailableFunds(wallet []byte) float64 {
	walletStr := SanitizePubKey(wallet)
	var total float64

	total = 0

	// fmt.Println("Ouech", hex.EncodeToString(wallet.pub))
	for _, out := range this.unspentTxOut[walletStr] {
		total += float64(out.out.Value)
	}

	return total / 1000000
}

// Used to create a transaction without loss
func (this *Blockchain) GetInOutFromUnspent(value int, destWallet []byte, outs []*UnspentTxOut) ([]TxIn, []TxOut) {
	insRes := []TxIn{}
	outsRes := []TxOut{}

	total := 0
	for _, out := range outs {
		insRes = append(insRes, TxIn{
			PrevHash: out.txHash,
			PrevIdx:  out.inIdx,
		})

		total += out.out.Value
	}

	outsRes = append(outsRes, TxOut{
		Value:   value,
		Address: destWallet,
	})

	if total > value {
		outsRes = append(outsRes, TxOut{
			Value:   total - value,
			Address: this.wallets["main.key"].pub,
		})
	}

	return insRes, outsRes
}
