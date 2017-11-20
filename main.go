package main

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/champii/crypto-dht/blockchain"
)

func main() {
	parseArgs(func(options blockchain.BlockchainOptions) {
		if options.Cluster > 0 {
			cluster(options)
		} else {
			node := startOne(options)

			if options.NoGui {
				node.Wait()
			} else {
				gui(node)
			}

			listenExitSignals(node)
		}
	})
}

func listenExitSignals(client *blockchain.Blockchain) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		exitProperly(client)

		os.Exit(0)
	}()
}

func exitProperly(client *blockchain.Blockchain) {
	client.Stop()
}

func startOne(options blockchain.BlockchainOptions) *blockchain.Blockchain {
	client := blockchain.New(options)

	if err := client.Start(); err != nil {
		client.Logger().Critical(err)
	}

	return client
}

func cluster(options blockchain.BlockchainOptions) {
	network := []*blockchain.Blockchain{}
	i := 0

	if len(options.BootstrapAddr) == 0 {
		client := startOne(options)

		network = append(network, client)

		options.BootstrapAddr = options.ListenAddr

		i++
	}

	for ; i < options.Cluster; i++ {
		options2 := options

		addrPort := strings.Split(options.ListenAddr, ":")

		addr := addrPort[0]

		port, _ := strconv.Atoi(addrPort[1])

		options2.ListenAddr = addr + ":" + strconv.Itoa(port+i)
		options2.Folder = options.Folder + strconv.Itoa(i)

		client := startOne(options2)

		network = append(network, client)
	}

	for {
		time.Sleep(time.Second)
	}
}
