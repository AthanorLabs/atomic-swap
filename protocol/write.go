package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
)

type infoFileContents struct {
	ContractAddress      string
	SwapID               uint64
	PrivateKeyInfo       *mcrypto.PrivateKeyInfo
	SharedSwapPrivateKey *mcrypto.PrivateKeyInfo
}

// WriteContractAddressToFile writes the contract address to the given file
func WriteContractAddressToFile(infofile, addr string) error {
	file, contents, err := setupFile(infofile)
	if err != nil {
		return err
	}

	contents.ContractAddress = addr

	bz, err := json.Marshal(contents)
	if err != nil {
		return err
	}

	_, err = file.Write(bz)
	return err
}

// WriteSwapIDToFile writes the swap ID to the given file
func WriteSwapIDToFile(infofile string, id uint64) error {
	file, contents, err := setupFile(infofile)
	if err != nil {
		return err
	}

	contents.SwapID = id

	bz, err := json.Marshal(contents)
	if err != nil {
		return err
	}

	_, err = file.Write(bz)
	return err
}

// WriteKeysToFile writes the given private key pair to the given file
func WriteKeysToFile(infofile string, keys *mcrypto.PrivateKeyPair, env common.Environment) error {
	file, contents, err := setupFile(infofile)
	if err != nil {
		return err
	}

	contents.PrivateKeyInfo = keys.Info(env)

	bz, err := json.Marshal(contents)
	if err != nil {
		return err
	}

	_, err = file.Write(bz)
	return err
}

// WriteSharedSwapKeyPairToFile writes the given private key pair to the given file
func WriteSharedSwapKeyPairToFile(infofile string, keys *mcrypto.PrivateKeyPair, env common.Environment) error {
	file, contents, err := setupFile(infofile)
	if err != nil {
		return err
	}

	contents.SharedSwapPrivateKey = keys.Info(env)

	bz, err := json.Marshal(contents)
	if err != nil {
		return err
	}

	_, err = file.Write(bz)
	return err
}

func setupFile(infofile string) (*os.File, *infoFileContents, error) {
	exists, err := exists(infofile)
	if err != nil {
		return nil, nil, err
	}

	var (
		file     *os.File
		contents *infoFileContents
	)
	if !exists {
		err = makeDir(filepath.Dir(infofile))
		if err != nil {
			return nil, nil, err
		}

		file, err = os.Create(filepath.Clean(infofile))
		if err != nil {
			return nil, nil, err
		}
	} else {
		file, err = os.OpenFile(filepath.Clean(infofile), os.O_RDWR, 0600)
		if err != nil {
			return nil, nil, err
		}

		bz, err := os.ReadFile(filepath.Clean(infofile))
		if err != nil {
			return nil, nil, err
		}

		if err = json.Unmarshal(bz, &contents); err != nil {
			return nil, nil, err
		}

		if err = file.Truncate(0); err != nil {
			return nil, nil, err
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			return nil, nil, err
		}
	}

	if contents == nil {
		contents = &infoFileContents{}
	}

	return file, contents, nil
}

func makeDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return nil
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
