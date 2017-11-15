package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/champii/crypto-dht/blockchain"
	"github.com/urfave/cli"
)

func prepareArgs() *cli.App {
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
	app.Version = "0.0.1"
	app.Compiled = time.Now()

	app.Usage = "Experimental Blockchain over DHT"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "b, bootstrap",
			Usage: "Connect to bootstrap node ip:port",
		},
		cli.StringFlag{
			Name:  "p, port",
			Usage: "Listening port",
			Value: "0.0.0.0:3000",
		},
		cli.StringFlag{
			Name:  "f, folder",
			Usage: "Base Folder",
			Value: os.Getenv("HOME") + "/.crypto-dht",
		},
		cli.BoolFlag{
			Name:  "i",
			Usage: "Interactif",
		},
		cli.BoolFlag{
			Name:  "s",
			Usage: "Stat mode",
		},
		cli.StringFlag{
			Name:  "S, send",
			Usage: "Send coins from main.key. Must be of the form 'amount:destAddress'",
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

func manageArgs() {
	app := prepareArgs()

	app.Action = func(c *cli.Context) error {
		options := blockchain.BlockchainOptions{
			ListenAddr:    c.String("p"),
			BootstrapAddr: c.String("b"),
			Folder:        c.String("f"),
			Send:          c.String("S"),
			Verbose:       c.Int("v"),
			Stats:         c.Bool("s"),
			Wallets:       c.Bool("w"),
			Interactif:    c.Bool("i"),
			NoGui:         c.Bool("g"),
			Mine:          c.Bool("m"),
		}

		if c.Int("n") > 0 {
			cluster(c.Int("n"), options)

			return nil
		}

		client := blockchain.New(options)

		if err := client.Start(); err != nil {
			client.Logger().Critical(err)
			return err
		}

		if options.NoGui {
			client.Wait()
		} else {
			gui()
		}

		return nil
	}

	app.Run(os.Args)
}

func main() {
	manageArgs()
}

var (
	AppName string
	BuiltAt string
	window  *astilectron.Window
	app     *astilectron.Astilectron
)

// ListItem represents a list item
type ListItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// handleMessages handles messages
func handleMessages(w *astilectron.Window, m bootstrap.MessageIn) (payload interface{}, err error) {
	switch m.Name {
	case "get.list":
		// Get user
		var u *user.User
		if u, err = user.Current(); err != nil {
			return
		}

		// Read dir
		var files []os.FileInfo
		if files, err = ioutil.ReadDir(u.HomeDir); err != nil {
			return
		}

		// Build list items
		var items []ListItem
		for _, f := range files {
			var item = ListItem{Name: f.Name()}
			if f.IsDir() {
				item.Type = "dir"
			} else {
				item.Type = "file"
			}
			items = append(items, item)
		}
		payload = items
	}
	return
}

func gui() {
	err := bootstrap.Run(bootstrap.Options{
		Asset:          Asset,
		RestoreAssets:  RestoreAssets,
		Homepage:       "index.html",
		MessageHandler: handleMessages,
		MenuOptions:    []*astilectron.MenuItemOptions{},
		OnWait: func(a *astilectron.Astilectron, w *astilectron.Window, _ *astilectron.Menu, t *astilectron.Tray, _ *astilectron.Menu) error {
			window = w
			app = a

			return nil
		},
		WindowOptions: &astilectron.WindowOptions{
			BackgroundColor: astilectron.PtrStr("#333"),
			Center:          astilectron.PtrBool(true),
			Height:          astilectron.PtrInt(700),
			Width:           astilectron.PtrInt(1100),
		},
	})

	if err != nil {
		return
	}

}

func cluster(count int, options blockchain.BlockchainOptions) {
	network := []*blockchain.Blockchain{}
	i := 0

	if len(options.BootstrapAddr) == 0 {
		client := startOne(options)

		network = append(network, client)

		options.BootstrapAddr = options.ListenAddr

		i++
	}

	for ; i < count; i++ {
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

func startOne(options blockchain.BlockchainOptions) *blockchain.Blockchain {
	client := blockchain.New(options)

	if err := client.Start(); err != nil {
		client.Logger().Critical(err)
	}

	return client
}
