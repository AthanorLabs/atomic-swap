package monero

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"

	"github.com/stretchr/testify/require"
)

func TestCallGenerateFromKeys(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	r, err := rand.Int(rand.Reader, big.NewInt(999))
	require.NoError(t, err)

	c := NewClient(common.DefaultBobMoneroEndpoint)
	err = c.callGenerateFromKeys(kp.SpendKey(), kp.ViewKey(), kp.Address(common.Mainnet),
		fmt.Sprintf("test-wallet-%d", r), "")
	require.NoError(t, err)
}
