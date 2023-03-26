//go:build !prod

package extethclient

import (
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
)

// This file is only for test support. Use the build tag "prod" to prevent
// symbols in this file from consuming space in production binaries.

// CreateTestClient creates and extended eth client using the passed ethereum
// wallet key. Cleanup on test completion is handled automatically.
func CreateTestClient(t *testing.T, ethKey *ecdsa.PrivateKey) EthClient {
	ctx := context.Background()
	ec, err := NewEthClient(ctx, common.Development, common.DefaultEthEndpoint, ethKey)
	require.NoError(t, err)
	t.Cleanup(func() {
		ec.Close()
	})
	return ec
}
