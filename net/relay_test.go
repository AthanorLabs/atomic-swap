// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package net

import (
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
)

func twoHostRelayerSetup(t *testing.T) (*Host, *Host) {
	// ha is not a relayer
	haCfg := basicTestConfig(t)
	haCfg.IsRelayer = false
	ha := newHost(t, haCfg)
	err := ha.Start()
	require.NoError(t, err)

	// hb is a relayer
	hbCfg := basicTestConfig(t)
	hbCfg.IsRelayer = true
	hbCfg.Bootnodes = []string{ha.Addresses()[0].String()}
	hb := newHost(t, hbCfg)
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
	require.True(t, hb.isRelayer)
	require.Len(t, peerIDs, 1) // discovers hb
	require.Equal(t, hb.PeerID(), peerIDs[0])

	peerIDs, err = hb.DiscoverRelayers()
	require.NoError(t, err)
	require.False(t, ha.isRelayer)
	require.Len(t, peerIDs, 0) // ha is not a relayer and not discovered
}

func createTestClaimRequest() *message.RelayClaimRequest {
	secret := [32]byte{0x1}
	sig := [65]byte{0x1}

	req := &message.RelayClaimRequest{
		SwapCreatorAddr: ethcommon.Address{0x1},
		Swap: &contracts.SwapCreatorSwap{
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

func TestHost_SubmitClaimToRelayer_dhtRelayer(t *testing.T) {
	ha, hb := twoHostRelayerSetup(t)

	// success path ha->hb, hb is a DHT relayer
	resp, err := ha.SubmitClaimToRelayer(hb.PeerID(), createTestClaimRequest())
	require.NoError(t, err)
	require.Equal(t, mockEthTXHash.Hex(), resp.TxHash.Hex())

	// failure path hb->ha, ha is NOT a DHT relayer. Note that the remote end
	// does not pass back the exact reason for rejecting a claim to avoid
	// possible privacy data leaks, but in this case it is because hb is not
	// a DHT advertising relayer.
	_, err = hb.SubmitClaimToRelayer(ha.PeerID(), createTestClaimRequest())
	require.ErrorContains(t, err, "failed to read RelayClaimResponse")
}

func TestHost_SubmitClaimToRelayer_xmrTakerRelayer(t *testing.T) {
	ha, hb := twoHostRelayerSetup(t)

	request := createTestClaimRequest()
	offerID := types.Hash{0x1}
	request.OfferID = &offerID

	// fail, because there is no ongoing swap between ha and hb
	_, err := hb.SubmitClaimToRelayer(ha.PeerID(), request)
	require.ErrorContains(t, err, "failed to read RelayClaimResponse")

	// create an ongoing swap between ha and hb
	swapState := &mockSwapState{offerID: offerID}
	err = ha.Initiate(peer.AddrInfo{ID: hb.PeerID()}, createSendKeysMessage(t), swapState)
	require.NoError(t, err)
	defer ha.CloseProtocolStream(offerID)

	// same steps will succeed now, because we started a swap first
	response, err := hb.SubmitClaimToRelayer(ha.PeerID(), request)
	require.NoError(t, err)
	require.Equal(t, mockEthTXHash, response.TxHash)
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
