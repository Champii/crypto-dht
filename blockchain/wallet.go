package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

type Wallet struct {
	name string
	key  *ecdsa.PrivateKey
	pub  []byte
}

func (this *Wallet) Name() string{
	return this.name
}

func (this *Wallet) Pub() []byte {
	return this.pub
}

func GetWallets(bc *Blockchain) error {
	wallets, err := ioutil.ReadDir(bc.options.Folder + "/wallets")

	if err != nil {
		return err
	}

	for _, wallet := range wallets {
		blob, err := ioutil.ReadFile(bc.options.Folder + "/wallets/" + wallet.Name())

		if err != nil {
			bc.logger.Warning("Wallet", wallet.Name(), "is not readable", err)

			return err
		}

		block, _ := pem.Decode([]byte(blob))
		x509Encoded := block.Bytes
		privateKey, err := x509.ParseECPrivateKey(x509Encoded)

		if err != nil {
			bc.logger.Warning("Wallet", wallet.Name(), "is corrupted !", err)

			return err
		}

		x509EncodedPub, err := x509.MarshalPKIXPublicKey(privateKey.Public())

		if err != nil {
			return err
		}

		pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

		bc.logger.Info("Loaded wallet", wallet.Name(), SanitizePubKey(pemEncodedPub))

		bc.wallets[wallet.Name()] = &Wallet{
			name: wallet.Name(),
			key:  privateKey,
			pub:  pemEncodedPub,
		}

	}

	return nil
}

func CreateWallet(name string, bc *Blockchain) (*Wallet, error) {
	_, err := os.Stat(bc.options.Folder + "/wallets/" + name + ".key")

	if err == nil {
		return nil, errors.New("Existing wallet " + name + ".key")
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return nil, err
	}

	priv, err := x509.MarshalECPrivateKey(key)

	if err != nil {
		return nil, err
	}

	x509EncodedPub, err := x509.MarshalPKIXPublicKey(key.Public())

	if err != nil {
		return nil, err
	}
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: priv})

	err = ioutil.WriteFile(bc.options.Folder+"/wallets/"+name+".key", pemEncoded, 0600)

	if err != nil {
		return nil, err
	}

	bc.logger.Info("Created wallet", name+".key", SanitizePubKey(pemEncodedPub))

	return &Wallet{
		name: name + ".key",
		key:  key,
		pub:  pemEncodedPub,
	}, nil
}

func SanitizePubKey(pub []byte) string {
	pemEncodedPubStr := strings.Replace(string(pub), "-----BEGIN PUBLIC KEY-----", "", 1)
	pemEncodedPubStr = strings.Replace(pemEncodedPubStr, "-----END PUBLIC KEY-----", "", 1)
	pemEncodedPubStr = strings.Replace(pemEncodedPubStr, "\n", "", -1)

	return pemEncodedPubStr
}

func UnsanitizePubKey(pub string) []byte {
	pub = pub[:64] + "\n" + pub[64:]
	return []byte("-----BEGIN PUBLIC KEY-----\n" + pub + "\n-----END PUBLIC KEY-----\n")
}
