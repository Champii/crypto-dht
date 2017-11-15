package blockchain

import (
	"fmt"
	"time"

	"github.com/buger/goterm"
)

type Stats struct {
	lastUpdate      int64
	lastHashes      int
	hashesPerSecAvg int
	hashesPerSec    []int
	foundBlocks     int
}

func (this *Stats) Update() {
	hashPerSec := this.lastHashes / int(time.Now().Unix()-this.lastUpdate)

	this.hashesPerSec = append(this.hashesPerSec, hashPerSec)

	if len(this.hashesPerSec) > 3600 {
		this.hashesPerSec = this.hashesPerSec[1:]
	}

	this.hashesPerSecAvg = 0

	for _, v := range this.hashesPerSec {
		this.hashesPerSecAvg += v
	}

	this.hashesPerSecAvg /= len(this.hashesPerSec)
	this.lastUpdate = time.Now().Unix()
	this.lastHashes = 0
}

func (this *Blockchain) Stats() {

	for {
		goterm.Clear()
		goterm.MoveCursor(1, 1)
		goterm.Println("Crypto DHT v0.0.1          Current Time: ", time.Now().Format(time.RFC1123))
		goterm.Println("")
		goterm.Println("Synced:         ", this.synced)
		goterm.Println("Mining:         ", this.options.Mine)
		goterm.Println("")
		goterm.Println("Funds:          ", this.GetAvailableFunds(this.wallets["main.key"].pub), "ctd")
		goterm.Println("Blocks height:  ", this.blocksHeight)
		goterm.Println("Address:        ", SanitizePubKey(this.wallets["main.key"].pub))
		goterm.Println("")

		if this.options.Mine {
			goterm.Println("Miner stats:")

			if len(this.stats.hashesPerSec) > 0 {
				goterm.Println("Hash/s:       ", this.stats.hashesPerSec[len(this.stats.hashesPerSec)-1])
			}
			goterm.Println("Hash/s avg:   ", this.stats.hashesPerSecAvg)
			goterm.Println("Found blocks: ", this.stats.foundBlocks)

		}

		time.Sleep(time.Second)
		goterm.Flush() // Call it every time at the end of rendering
	}
	goterm.Flush() // Call it every time at the end of rendering
}

func (this *Blockchain) ShowWallets() {
	for name, wallet := range this.wallets {
		fmt.Println("Name:    ", name)
		fmt.Println("Address: ", SanitizePubKey(wallet.pub))
		fmt.Println("Amount:  ", this.GetAvailableFunds(wallet.pub))
		fmt.Println("")
	}
}
