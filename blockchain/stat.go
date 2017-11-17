package blockchain

import (
	"fmt"
	"time"

	"github.com/buger/goterm"
)

type Stats struct {
	lastUpdate      int64
	lastHashes      int
	HashesPerSecAvg int
	HashesPerSec    []int
	foundBlocks     int
}

func (this *Stats) Update() {
	passed := int(time.Now().Unix() - this.lastUpdate)

	if passed == 0 {
		passed = 1
	}

	hashPerSec := this.lastHashes / passed

	this.HashesPerSec = append(this.HashesPerSec, hashPerSec)

	if len(this.HashesPerSec) > 3600 {
		this.HashesPerSec = this.HashesPerSec[1:]
	}

	this.HashesPerSecAvg = 0

	for _, v := range this.HashesPerSec {
		this.HashesPerSecAvg += v
	}

	this.HashesPerSecAvg /= len(this.HashesPerSec)
	this.lastUpdate = time.Now().Unix()
	this.lastHashes = 0
}

func (this *Blockchain) StatsLoop() {

	for {
		goterm.Clear()
		goterm.MoveCursor(1, 1)
		goterm.Println("Crypto DHT v0.0.1          Current Time: ", time.Now().Format(time.RFC1123))
		goterm.Println("")
		goterm.Println("Synced:         ", this.synced)
		goterm.Println("Mining:         ", this.options.Mine)
		goterm.Println("")
		goterm.Println("Funds:          ", this.GetAvailableFunds(this.wallets["main.key"].pub), "ctd")
		goterm.Println("Blocks height:  ", this.BlocksHeight())
		goterm.Println("Address:        ", SanitizePubKey(this.wallets["main.key"].pub))
		goterm.Println("")

		if this.options.Mine {
			goterm.Println("Miner stats:")

			if len(this.stats.HashesPerSec) > 0 {
				goterm.Println("Hash/s:       ", this.stats.HashesPerSec[len(this.stats.HashesPerSec)-1])
			}
			goterm.Println("Hash/s avg:   ", this.stats.HashesPerSecAvg)
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
