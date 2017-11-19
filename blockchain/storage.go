package blockchain

import (
	"io/ioutil"
	"net/url"
	"strconv"
	"errors"
	"os"

	"github.com/vmihailenco/msgpack"
)

func SetupStorage(bc *Blockchain) error {
	// ioutil.ReadDir(bc.options.Folder)
	stat, err := os.Stat(bc.options.Folder)

	if err != nil {
		os.Mkdir(bc.options.Folder, 0755)
	} else {
		if !stat.IsDir() {
			return errors.New(bc.options.Folder + " is not a folder")
		}
	}

	stat, err = os.Stat(bc.options.Folder + "/chain")
	if err != nil {
		os.Mkdir(bc.options.Folder+"/chain", 0755)
	} else {
		if !stat.IsDir() {
			return errors.New(bc.options.Folder + "/chain" + " is not a folder")
		}
	}

	stat, err = os.Stat(bc.options.Folder + "/unspent")
	if err != nil {
		os.Mkdir(bc.options.Folder+"/unspent", 0755)
	} else {
		if !stat.IsDir() {
			return errors.New(bc.options.Folder + "/unspent" + " is not a folder")
		}
	}

	stat, err = os.Stat(bc.options.Folder + "/wallets")
	if err != nil {
		os.Mkdir(bc.options.Folder+"/wallets", 0755)
	} else {
		if !stat.IsDir() {
			return errors.New(bc.options.Folder + "/wallets" + " is not a folder")
		}
	}

	err = GetWallets(bc)

	if err != nil {
		return err
	}

	if len(bc.wallets) == 0 {
		wallet, err := CreateWallet("main", bc)

		if err != nil {
			return err
		}

		bc.wallets[wallet.name] = wallet
	}

	return nil
}

// 1000 headers by files
func LoadStoredHeaders(bc *Blockchain) error {
	dir, err := ioutil.ReadDir(bc.options.Folder + "/chain")

	if err != nil {
		return err
	}

	for _, file := range dir {
		headersByte, err := ioutil.ReadFile(bc.options.Folder + "/chain/" +  file.Name())

		if err != nil {
			return err
		}

		var headers []BlockHeader
		err = msgpack.Unmarshal(headersByte, &headers)

		if err != nil {
			return err
		}

		bc.headers = append(bc.headers, headers...)

		if !bc.AreHeadersGood() {
			return errors.New("Load headers: Bad blocks loaded in file " + file.Name())
		}
	}

	bc.logger.Debug("Loaded", len(bc.headers) - 1, "blocks !")

	return nil
}


func StoreLastHeaders(bc *Blockchain) error {
	headersLen := len(bc.headers)

	nb := headersLen % 1000

	toStore, err := msgpack.Marshal(bc.headers[1 + headersLen - nb:])

	if err != nil {
		return err
	}

	fileNumber := strconv.Itoa(headersLen / 1000)

	err = ioutil.WriteFile(bc.options.Folder + "/chain/" +  fileNumber, toStore, 0644)

	if err != nil {
		return err
	}

	bc.logger.Debug("Stored", nb -1, "blocks in file", fileNumber)

	return nil
}

func LoadUnspent(bc *Blockchain) error {
	dir, err := ioutil.ReadDir(bc.options.Folder + "/unspent")

	if err != nil {
		return err
	}

	for _, file := range dir {
		unspentsByte, err := ioutil.ReadFile(bc.options.Folder + "/unspent/" +  file.Name())

		if err != nil {
			return err
		}

		var unspents []UnspentTxOut
		err = msgpack.Unmarshal(unspentsByte, &unspents)

		if err != nil {
			return err
		}

		wallet, _ := url.PathUnescape(file.Name())

		bc.unspentTxOut[wallet] = unspents
	}

	bc.logger.Debug("Loaded", len(bc.unspentTxOut), "wallets unspent out")

	return nil
}

func StoreUnspent(bc *Blockchain) error {
	for walletName, unspent := range bc.unspentTxOut {
		toStore, err := msgpack.Marshal(unspent)

		if err != nil {
			return err
		}

		err = ioutil.WriteFile(bc.options.Folder + "/unspent/" +  url.PathEscape(walletName), toStore, 0644)

		if err != nil {
			return err
		}
	}

	bc.logger.Debug("Stored", len(bc.unspentTxOut), "wallets unspent out")

	return nil
}
