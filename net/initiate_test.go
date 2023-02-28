package net

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/stretchr/testify/require"
)

func createSendKeysMessage(t *testing.T) *message.SendKeysMessage {
	keysAndProof, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)
	return &message.SendKeysMessage{
		OfferID:            types.Hash{},
		ProvidedAmount:     new(apd.Decimal),
		PublicSpendKey:     keysAndProof.PublicKeyPair.SpendKey(),
		PrivateViewKey:     keysAndProof.PrivateKeyPair.ViewKey(),
		DLEqProof:          hex.EncodeToString(keysAndProof.DLEqProof.Proof()),
		Secp256k1PublicKey: keysAndProof.Secp256k1PublicKey,
		EthAddress:         ethcommon.Address{},
	}
}

func TestHost_Initiate(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	err := ha.Start()
	require.NoError(t, err)
	hb := newHost(t, basicTestConfig(t))
	err = hb.Start()
	require.NoError(t, err)

	err = ha.h.Connect(ha.ctx, hb.h.AddrInfo())
	require.NoError(t, err)

	err = ha.Initiate(hb.h.AddrInfo(), createSendKeysMessage(t), new(mockSwapState))
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	ha.swapMu.Lock()
	require.NotNil(t, ha.swaps[testID])
	ha.swapMu.Unlock()

	hb.swapMu.Lock()
	require.NotNil(t, hb.swaps[testID])
	hb.swapMu.Unlock()
}

func TestHost_ConcurrentSwaps(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	err := ha.Start()
	require.NoError(t, err)

	hbCfg := basicTestConfig(t)
	hbCfg.Bootnodes = ha.h.Addresses() // get some test coverage on our bootnode code
	hb := newHost(t, hbCfg)
	err = hb.Start()
	require.NoError(t, err)

	testID2 := types.Hash{98}

	err = ha.h.Connect(ha.ctx, hb.h.AddrInfo())
	require.NoError(t, err)

	err = ha.Initiate(hb.h.AddrInfo(), createSendKeysMessage(t), new(mockSwapState))
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	ha.swapMu.Lock()
	require.NotNil(t, ha.swaps[testID])
	ha.swapMu.Unlock()

	hb.swapMu.Lock()
	require.NotNil(t, hb.swaps[testID])
	hb.swapMu.Unlock()

	hb.handler.(*mockHandler).id = testID2

	err = ha.Initiate(hb.h.AddrInfo(), createSendKeysMessage(t), &mockSwapState{testID2})
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 1500)

	ha.swapMu.Lock()
	require.NotNil(t, ha.swaps[testID])
	ha.swapMu.Unlock()

	hb.swapMu.Lock()
	require.NotNil(t, hb.swaps[testID])
	hb.swapMu.Unlock()
}
