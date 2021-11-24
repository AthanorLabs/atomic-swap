package common

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/noot/atomic-swap/monero"
)

func WriteKeysToFile(basepath string, keys *monero.PrivateKeyPair) error {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	path := fmt.Sprintf("%s-%s.key", basepath, t)

	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}

	bz, err := keys.Marshal()
	if err != nil {
		return err
	}

	_, err = file.Write(bz)
	return err
}

func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
