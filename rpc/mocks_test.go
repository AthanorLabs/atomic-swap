package rpc

import (
	"time"

	"github.com/MarinX/monerorpc/wallet"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2ptest "github.com/libp2p/go-libp2p/core/test"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/txsender"
)

//
// This file only contains mock definitions used by other test files
//

type mockNet struct {
	peerID peer.ID
}

func (*mockNet) Addresses() []string {
	panic("not implemented")
}

func (m *mockNet) PeerID() peer.ID {
	if m.peerID == "" {
		var err error
		m.peerID, err = libp2ptest.RandPeerID()
		if err != nil {
			panic(err)
		}
	}
	return m.peerID
}

func (*mockNet) ConnectedPeers() []string {
	panic("not implemented")
}

func (*mockNet) Discover(provides types.ProvidesCoin, searchTime time.Duration) ([]peer.ID, error) {
	return nil, nil
}

func (*mockNet) Query(who peer.ID) (*message.QueryResponse, error) {
	return &message.QueryResponse{Offers: []*types.Offer{{ID: testSwapID}}}, nil
}

func (*mockNet) Initiate(who peer.AddrInfo, msg *message.SendKeysMessage, s common.SwapStateNet) error {
	return nil
}

func (*mockNet) CloseProtocolStream(types.Hash) {
	panic("not implemented")
}

type mockSwapManager struct{}

func (*mockSwapManager) WriteSwapToDB(_ *swap.Info) error {
	return nil
}

func (*mockSwapManager) GetPastIDs() ([]types.Hash, error) {
	panic("not implemented")
}

func (*mockSwapManager) GetPastSwap(id types.Hash) (*swap.Info, error) {
	return &swap.Info{}, nil
}

func (*mockSwapManager) GetOngoingSwaps() ([]swap.Info, error) {
	return nil, nil
}

func (*mockSwapManager) GetOngoingSwap(id types.Hash) (swap.Info, error) {
	statusCh := make(chan types.Status, 1)
	statusCh <- types.CompletedSuccess

	return *swap.NewInfo(
		id,
		types.ProvidesETH,
		1,
		1,
		1,
		types.EthAssetETH,
		types.CompletedSuccess,
		1,
		statusCh,
	), nil
}

func (*mockSwapManager) AddSwap(*swap.Info) error {
	panic("not implemented")
}

func (*mockSwapManager) CompleteOngoingSwap(*swap.Info) error {
	panic("not implemented")
}

type mockXMRTaker struct{}

func (*mockXMRTaker) Provides() types.ProvidesCoin {
	panic("not implemented")
}

func (*mockXMRTaker) GetOngoingSwapState(types.Hash) common.SwapState {
	return new(mockSwapState)
}

func (*mockXMRTaker) InitiateProtocol(providesAmount float64, _ *types.Offer) (common.SwapState, error) {
	return new(mockSwapState), nil
}

func (*mockXMRTaker) Refund(types.Hash) (ethcommon.Hash, error) {
	panic("not implemented")
}

func (*mockXMRTaker) SetSwapTimeout(_ time.Duration) {
	panic("not implemented")
}

func (*mockXMRTaker) ExternalSender(_ types.Hash) (*txsender.ExternalSender, error) {
	panic("not implemented")
}

type mockXMRMaker struct{}

func (m *mockXMRMaker) Provides() types.ProvidesCoin {
	panic("not implemented")
}

func (m *mockXMRMaker) GetOngoingSwapState(hash types.Hash) common.SwapState {
	panic("not implemented")
}

func (*mockXMRMaker) MakeOffer(offer *types.Offer, _ string, _ float64) (*types.OfferExtra, error) {
	offerExtra := &types.OfferExtra{
		StatusCh: make(chan types.Status, 1),
	}
	offerExtra.StatusCh <- types.CompletedSuccess
	return offerExtra, nil
}

func (*mockXMRMaker) GetOffers() []*types.Offer {
	panic("not implemented")
}

func (*mockXMRMaker) ClearOffers([]types.Hash) error {
	panic("not implemented")
}

func (*mockXMRMaker) GetMoneroBalance() (string, *wallet.GetBalanceResponse, error) {
	panic("not implemented")
}

type mockSwapState struct{}

func (*mockSwapState) HandleProtocolMessage(msg message.Message) error {
	return nil
}

func (*mockSwapState) Exit() error {
	return nil
}

func (*mockSwapState) SendKeysMessage() *message.SendKeysMessage {
	return &message.SendKeysMessage{}
}

func (*mockSwapState) ID() types.Hash {
	return testSwapID
}

type mockProtocolBackend struct {
	sm *mockSwapManager
}

func newMockProtocolBackend() *mockProtocolBackend {
	return &mockProtocolBackend{
		sm: new(mockSwapManager),
	}
}

func (*mockProtocolBackend) Env() common.Environment {
	return common.Development
}

func (*mockProtocolBackend) SetSwapTimeout(timeout time.Duration) {
	panic("not implemented")
}

func (*mockProtocolBackend) SwapTimeout() time.Duration {
	panic("not implemented")
}

func (b *mockProtocolBackend) SwapManager() swap.Manager {
	return b.sm
}

func (*mockProtocolBackend) SetXMRDepositAddress(mcrypto.Address, types.Hash) {
	panic("not implemented")
}

func (*mockProtocolBackend) ClearXMRDepositAddress(types.Hash) {
	panic("not implemented")
}

func (*mockProtocolBackend) ETHClient() extethclient.EthClient {
	panic("not implemented")
}
