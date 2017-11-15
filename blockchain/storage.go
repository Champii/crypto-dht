package blockchain

import (
	"errors"
	"os"
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
