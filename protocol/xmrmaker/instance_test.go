package xmrmaker

import (
	"context"
	"errors"
	"math/big"
	"path"
	"sync"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/db"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	pswap "github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	testWallet = "test-wallet"
)

type mockNet struct {
	msgMu sync.Mutex  // lock needed, as SendSwapMessage is called async from timeout handlers
	msg   net.Message // last value passed to SendSwapMessage
}

func (n *mockNet) LastSentMessage() net.Message {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	return n.msg
}

func (n *mockNet) SendSwapMessage(msg net.Message, _ types.Hash) error {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	n.msg = msg
	return nil
}

func (n *mockNet) CloseProtocolStream(_ types.Hash) {}

func newSwapManager(t *testing.T) pswap.Manager {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := pswap.NewMockDatabase(ctrl)
	db.EXPECT().GetAllSwaps()
	db.EXPECT().PutSwap(gomock.Any()).AnyTimes()

	sm, err := pswap.NewManager(db)
	require.NoError(t, err)
	return sm
}

func newTestBackend(t *testing.T) backend.Backend {
	pk := tests.GetMakerTestKey(t)
	ec, chainID := tests.NewEthClient(t)
	env := common.Development

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	var forwarderAddress ethcommon.Address
	_, tx, contract, err := contracts.DeploySwapFactory(txOpts, ec, forwarderAddress)
	require.NoError(t, err)

	addr, err := bind.WaitDeployed(context.Background(), ec, tx)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rdb := backend.NewMockRecoveryDB(ctrl)
	rdb.EXPECT().PutContractSwapInfo(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().PutSwapPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().PutSharedSwapPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().PutSwapRelayerInfo(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().DeleteSwap(gomock.Any()).Return(nil).AnyTimes()

	extendedEC, err := extethclient.NewEthClient(context.Background(), env, ec, pk)
	require.NoError(t, err)

	bcfg := &backend.Config{
		Ctx:                 context.Background(),
		MoneroClient:        monero.CreateWalletClient(t),
		EthereumClient:      extendedEC,
		Environment:         common.Development,
		SwapContract:        contract,
		SwapContractAddress: addr,
		SwapManager:         newSwapManager(t),
		Net:                 new(mockNet),
		RecoveryDB:          rdb,
	}

	b, err := backend.NewBackend(bcfg)
	require.NoError(t, err)

	return b
}

func newTestInstanceAndDB(t *testing.T) (*Instance, *offers.MockDatabase) {
	b := newTestBackend(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := offers.NewMockDatabase(ctrl)
	db.EXPECT().GetAllOffers()
	db.EXPECT().DeleteOffer(gomock.Any()).Return(nil).AnyTimes()

	net := NewMockHost(ctrl)

	cfg := &Config{
		Backend:        b,
		DataDir:        path.Join(t.TempDir(), "xmrmaker"),
		WalletFile:     testWallet,
		WalletPassword: "",
		Database:       db,
		Network:        net,
	}

	xmrmaker, err := NewInstance(cfg)
	require.NoError(t, err)

	monero.MineMinXMRBalance(t, b.XMRClient(), 1.0)
	err = b.XMRClient().Refresh()
	require.NoError(t, err)
	return xmrmaker, db
}

func TestInstance_createOngoingSwap(t *testing.T) {
	inst, offerDB := newTestInstanceAndDB(t)
	rdb := inst.backend.RecoveryDB().(*backend.MockRecoveryDB)

	offer := types.NewOffer(
		types.ProvidesXMR,
		1,
		1,
		1,
		types.EthAssetETH,
	)

	s := &pswap.Info{
		ID:             offer.ID,
		Provides:       types.ProvidesXMR,
		ProvidedAmount: 1,
		ReceivedAmount: 1,
		ExchangeRate:   types.ExchangeRate(1),
		EthAsset:       types.EthAssetETH,
		Status:         types.XMRLocked,
	}

	sk, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	rdb.EXPECT().GetSwapRelayerInfo(s.ID).Return(nil, errors.New("some error"))
	rdb.EXPECT().GetSharedSwapPrivateKey(s.ID).Return(nil, errors.New("some error"))
	rdb.EXPECT().GetContractSwapInfo(s.ID).Return(&db.EthereumSwapInfo{
		StartNumber:     big.NewInt(1),
		ContractAddress: inst.backend.ContractAddr(),
		Swap: contracts.SwapFactorySwap{
			Timeout0: big.NewInt(1),
			Timeout1: big.NewInt(2),
		},
	}, nil)
	rdb.EXPECT().GetSwapPrivateKey(s.ID).Return(
		sk.SpendKey(), nil,
	)
	offerDB.EXPECT().GetOffer(s.ID).Return(offer, nil)

	err = inst.createOngoingSwap(*s)
	require.NoError(t, err)

	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()
	close(inst.swapStates[s.ID].done)
}
