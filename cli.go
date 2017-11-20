package main

import (
	"os"
	"time"

	"github.com/champii/crypto-dht/blockchain"
	"github.com/urfave/cli"
)

func parseArgs(done func(blockchain.BlockchainOptions)) {
	app := setupCli()

	app.Action = func(c *cli.Context) error {
		options := blockchain.BlockchainOptions{
			ListenAddr:    c.String("l"),
			BootstrapAddr: c.String("c"),
			Folder:        c.String("f"),
			Send:          c.String("S"),
			Verbose:       c.Int("v"),
			Stats:         c.Bool("s"),
			Wallets:       c.Bool("w"),
			NoGui:         c.Bool("g"),
			Mine:          c.Bool("m"),
			Cluster:       c.Int("n"),
		}

		if options.Cluster > 0 {
			options.Send = ""
			options.Stats = false
			options.NoGui = true
			options.Wallets = false
		}

		if options.Stats || len(options.Send) > 0 {
			options.NoGui = true
			options.Wallets = false
		}

		if len(options.Send) > 0 {
			options.Stats = false
		}

		done(options)

		return nil
	}

	app.Run(os.Args)
}

func setupCli() *cli.App {
	cli.AppHelpTemplate = `NAME:
	{{.Name}} - {{.Usage}}

USAGE:
	{{if .VisibleFlags}}{{.HelpName}} [options]{{end}}
	{{if len .Authors}}
AUTHOR:
	{{range .Authors}}{{ . }}{{end}}
	{{end}}{{if .Commands}}
VERSION:
	{{.Version}}

OPTIONS:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}{{if .Copyright }}

COPYRIGHT:
	{{.Copyright}}
	{{end}}{{if .Version}}
	{{end}}`

	cli.VersionFlag = cli.BoolFlag{
		Name:  "V, version",
		Usage: "Print version",
	}

	cli.HelpFlag = cli.BoolFlag{
		Name:  "h, help",
		Usage: "Print help",
	}

	app := cli.NewApp()

	app.Name = "Crypto-Dht"
	app.Version = "0.1.0"
	app.Compiled = time.Now()

	app.Usage = "Experimental Blockchain over DHT"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c, connect",
			Usage: "Connect to node ip:port. If not set, startup a bootstrap node.",
		},
		cli.StringFlag{
			Name:  "l, listen",
			Usage: "Listening address and port",
			Value: "0.0.0.0:3000",
		},
		cli.StringFlag{
			Name:  "f, folder",
			Usage: "Config Folder",
			Value: os.Getenv("HOME") + "/.crypto-dht",
		},
		cli.BoolFlag{
			Name:  "s",
			Usage: "Stat mode",
		},
		cli.BoolFlag{
			Name:  "m",
			Usage: "Mine",
		},
		cli.BoolFlag{
			Name:  "w",
			Usage: "Show wallets and amount",
		},
		cli.BoolFlag{
			Name:  "g",
			Usage: "Deactivate GUI",
		},
		cli.StringFlag{
			Name:  "S, send",
			Usage: "Send coins from main.key. Must be of the form 'amount:destAddress'",
		},
		cli.IntFlag{
			Name:  "n, network",
			Value: 0,
			Usage: "Spawn X new `nodes` network. If -b is not specified, a new network is created.",
		},
		cli.IntFlag{
			Name:  "v, verbose",
			Value: 3,
			Usage: "Verbose `level`, 0 for CRITICAL and 5 for DEBUG",
		},
	}

	app.UsageText = "dht [options]"

	return app
}
