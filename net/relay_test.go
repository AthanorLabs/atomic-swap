package net

import (
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
)

func twoHostRelayerSetup(t *testing.T) (*Host, *Host) {
	// ha is not a relayer
	ha := newHost(t, basicTestConfig(t), true)
	err := ha.Start()
	require.NoError(t, err)

	// hb is a relayer
	hbCfg := basicTestConfig(t)
	hbCfg.Bootnodes = []string{ha.Addresses()[0].String()}
	hb := newHost(t, hbCfg, true)
	require.NoError(t, err)
	err = hb.Start()
	require.NoError(t, err)

	ha.Advertise()                     // hb wasn't around on ha's first advertisement loop
	time.Sleep(500 * time.Millisecond) // give hb time to advertise in DHT

	return ha, hb
}

func TestHost_DiscoverRelayers(t *testing.T) {
	ha, hb := twoHostRelayerSetup(t)

	peerIDs, err := ha.DiscoverRelayers()
	require.NoError(t, err)
	require.Len(t, peerIDs, 1)
	require.Equal(t, hb.PeerID(), peerIDs[0])

	peerIDs, err = hb.DiscoverRelayers()
	require.NoError(t, err)
	require.Len(t, peerIDs, 1)
	require.Equal(t, ha.PeerID(), peerIDs[0])
}

func createTestClaimRequest() *message.RelayClaimRequest {
	secret := [32]byte{0x1}
	sig := [65]byte{0x1}

	req := &message.RelayClaimRequest{
		SFContractAddress: ethcommon.Address{0x1},
		Swap: &contracts.SwapFactorySwap{
			Owner:        ethcommon.Address{0x1},
			Claimer:      ethcommon.Address{0x1},
			PubKeyClaim:  [32]byte{0x1},
			PubKeyRefund: [32]byte{0x1},
			Timeout0:     big.NewInt(time.Now().Add(30 * time.Minute).Unix()),
			Timeout1:     big.NewInt(time.Now().Add(60 * time.Minute).Unix()),
			Asset:        ethcommon.Address(types.EthAssetETH),
			Value:        big.NewInt(1e18),
			Nonce:        big.NewInt(1),
		},
		Secret:    secret[:],
		Signature: sig[:],
	}

	return req
}

func TestHost_SubmitClaimToRelayer(t *testing.T) {
	ha, hb := twoHostRelayerSetup(t)

	resp, err := ha.SubmitClaimToRelayer(hb.PeerID(), createTestClaimRequest())
	require.NoError(t, err)
	require.Equal(t, mockEthTXHash.Hex(), resp.TxHash.Hex())
}

func TestHost_SubmitClaimToRelayer_fail(t *testing.T) {
	ha, hb := twoHostRelayerSetup(t)

	req := createTestClaimRequest()
	req.Secret = []byte{0x1} // wrong size
	_, err := ha.SubmitClaimToRelayer(hb.PeerID(), req)
	require.ErrorContains(t, err, "Field validation for 'Secret' failed on the 'len' tag")

	req = createTestClaimRequest()
	req.Signature = []byte{0x1, 0x2} // wrong size
	_, err = ha.SubmitClaimToRelayer(hb.PeerID(), req)
	require.ErrorContains(t, err, "Field validation for 'Signature' failed on the 'len' tag")
}
