package main

import (
	"encoding/json"
	"fmt"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/champii/crypto-dht/blockchain"
)

var (
	AppName string
	BuiltAt string
	window  *astilectron.Window
	app     *astilectron.Astilectron
	bc      *blockchain.Blockchain
)

type MinerInfo struct {
	Hashrate               int  `json:"hashrate"`
	Running                bool `json:"running"`
	WaitingTransactions    int  `json:"waitingTransactions"`
	ProcessingTransactions int  `json:"processingTransactions"`
}

type BaseInfo struct {
	MinerInfo          MinerInfo              `json:"minerInfo"`
	Wallets            []WalletClient         `json:"wallets"`
	NodesNb            int                    `json:"nodesNb"`
	Synced             bool                   `json:"synced"`
	BlocksHeight       int64                  `json:"blocksHeight"`
	Difficulty         int64                  `json:"difficulty"`
	NextDifficulty     int64                  `json:"nextDifficulty"`
	TimeSinceLastBlock int64                  `json:"timeSinceLastBlock"`
	StoredKeys         int                    `json:"storedKeys"`
	History            []blockchain.HistoryTx `json:"history"`
	OwnWaitingTx       []blockchain.HistoryTx `json:"ownWaitingTx"`
}

type WalletClient struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

func GetBaseInfos() BaseInfo {
	wallets := bc.Wallets()
	var walletsRes []WalletClient

	for _, wallet := range wallets {
		walletsRes = append(walletsRes, WalletClient{
			Name:    wallet.Name(),
			Address: blockchain.SanitizePubKey(wallet.Pub()),
			Amount:  bc.GetAvailableFunds(wallet.Pub()),
		})
	}

	stats := bc.Stats()

	hashRate := 0

	if len(stats.HashesPerSec) > 0 {
		hashRate = stats.HashesPerSec[len(stats.HashesPerSec)-1]
	}

	return BaseInfo{
		Wallets:            walletsRes,
		NodesNb:            bc.GetConnectedNodesNb(),
		Synced:             bc.Synced(),
		BlocksHeight:       bc.BlocksHeight(),
		Difficulty:         bc.Difficulty(),
		NextDifficulty:     bc.NextDifficulty(),
		StoredKeys:         bc.StoredKeys(),
		TimeSinceLastBlock: bc.TimeSinceLastBlock(),
		History:            bc.GetOwnHistory(),
		OwnWaitingTx:       bc.GetOwnWaitingTx(),
		MinerInfo: MinerInfo{
			Hashrate:               hashRate,
			Running:                bc.Running(),
			WaitingTransactions:    bc.WaitingTransactionCount(),
			ProcessingTransactions: bc.ProcessingTransactionCount(),
		},
	}
}

func handleMessages(w *astilectron.Window, m bootstrap.MessageIn) (payload interface{}, err error) {
	switch m.Name {
	case "getInfos":
		payload = GetBaseInfos()

	case "send":
		var r string

		json.Unmarshal(m.Payload, &r)

		err := bc.SendTo(r)

		payload = ""

		if err != nil {
			payload = err.Error()
		}
	}
	return
}

func gui(bc_ *blockchain.Blockchain) {
	err := bootstrap.Run(bootstrap.Options{
		Asset:          Asset,
		RestoreAssets:  RestoreAssets,
		Homepage:       "index.html",
		MessageHandler: handleMessages,
		MenuOptions:    []*astilectron.MenuItemOptions{},
		OnWait: func(a *astilectron.Astilectron, w *astilectron.Window, _ *astilectron.Menu, t *astilectron.Tray, _ *astilectron.Menu) error {
			window = w
			app = a
			bc = bc_

			// w.OpenDevTools()

			return nil
		},
		WindowOptions: &astilectron.WindowOptions{
			BackgroundColor: astilectron.PtrStr("#333"),
			Center:          astilectron.PtrBool(true),
			Height:          astilectron.PtrInt(450),
			Width:           astilectron.PtrInt(1050),
			Resizable:       astilectron.PtrBool(false),
			Frame:           astilectron.PtrBool(false),
			HasShadow:       astilectron.PtrBool(true),
			Transparent:     astilectron.PtrBool(true),
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
