package common

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"os"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Reverse returns a copy of the slice with the bytes in reverse order
func Reverse(s []byte) []byte {
	l := len(s)
	rs := make([]byte, l)
	for i := 0; i < l; i++ {
		rs[i] = s[l-i-1]
	}
	return rs
}

// EthereumPrivateKeyToAddress returns the address associated with a private key
func EthereumPrivateKeyToAddress(privkey *ecdsa.PrivateKey) ethcommon.Address {
	pub := privkey.Public().(*ecdsa.PublicKey)
	return ethcrypto.PubkeyToAddress(*pub)
}

// GetTopic returns the Ethereum topic (ie. keccak256 hash) of the given event or function signature string.
func GetTopic(sig string) ethcommon.Hash {
	h := ethcrypto.Keccak256([]byte(sig))
	var b [32]byte
	copy(b[:], h)
	return b
}

// MakeDir creates a directory, including leading directories, if they don't already exist.
// File permissions of created directories are only granted to the current user.
func MakeDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}

// FileExists returns whether the given file exists. If a directory exists
// with the name of the passed file, an error is returned.
func FileExists(path string) (bool, error) {
	st, err := os.Stat(path)
	if err == nil {
		if st.IsDir() {
			return false, fmt.Errorf("%q is occupied by a directory", path)
		}
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// SleepWithContext is the same as time.Sleep(...) but with preemption if the context is
// complete. Returns nil if the sleep completed, otherwise the context's error.
func SleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// GetFreeTCPPort returns an OS allocated and immediately freed port. There is
// nothing preventing something else on the system from using the port before
// the caller has a chance, but OS allocated ports are randomised to make the
// risk negligible.
func GetFreeTCPPort() (uint, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = ln.Close() }()

	return uint(ln.Addr().(*net.TCPAddr).Port), nil
}
