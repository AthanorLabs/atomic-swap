package monero

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/noot/atomic-swap/common"
)

// WriteKeysToFile writes the given private key pair to a file within the given path.
func WriteKeysToFile(basepath string, keys *PrivateKeyPair, env common.Environment) error {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	path := fmt.Sprintf("%s-%s.key", basepath, t)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}

	bz, err := keys.Marshal(env)
	if err != nil {
		return err
	}

	_, err = file.Write(bz)
	return err
}
