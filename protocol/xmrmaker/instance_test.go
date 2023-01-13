package xmrmaker

import (
	"context"
	"errors"
	"math/big"
	"path"
	"sync"
	"testing"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/db"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
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
	msgMu sync.Mutex      // lock needed, as SendSwapMessage is called async from timeout handlers
	msg   message.Message // last value passed to SendSwapMessage
}

func (n *mockNet) LastSentMessage() message.Message {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	return n.msg
}

func (n *mockNet) SendSwapMessage(msg message.Message, _ types.Hash) error {
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

func newBackendAndNet(t *testing.T) (backend.Backend, *mockNet) {
	pk := tests.GetMakerTestKey(t)
	ec, chainID := tests.NewEthClient(t)

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

	extendedEC, err := extethclient.NewEthClient(context.Background(), ec, pk)
	require.NoError(t, err)

	net := new(mockNet)
	bcfg := &backend.Config{
		Ctx:                 context.Background(),
		MoneroClient:        monero.CreateWalletClient(t),
		EthereumClient:      extendedEC,
		Environment:         common.Development,
		SwapContract:        contract,
		SwapContractAddress: addr,
		SwapManager:         newSwapManager(t),
		Net:                 net,
		RecoveryDB:          rdb,
	}

	b, err := backend.NewBackend(bcfg)
	require.NoError(t, err)

	return b, net
}

func newTestInstanceAndDBAndNet(t *testing.T) (*Instance, *offers.MockDatabase, *mockNet) {
	b, net := newBackendAndNet(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := offers.NewMockDatabase(ctrl)
	db.EXPECT().GetAllOffers()
	db.EXPECT().DeleteOffer(gomock.Any()).Return(nil).AnyTimes()

	host := NewMockP2pnetHost(ctrl)

	cfg := &Config{
		Backend:        b,
		DataDir:        path.Join(t.TempDir(), "xmrmaker"),
		WalletFile:     testWallet,
		WalletPassword: "",
		Database:       db,
		Network:        host,
	}

	xmrmaker, err := NewInstance(cfg)
	require.NoError(t, err)

	oneXMR := coins.MoneroToPiconero(apd.New(1, 0))
	monero.MineMinXMRBalance(t, b.XMRClient(), oneXMR)
	err = b.XMRClient().Refresh()
	require.NoError(t, err)
	return xmrmaker, db, net
}

func newTestInstanceAndDB(t *testing.T) (*Instance, *offers.MockDatabase) {
	inst, db, _ := newTestInstanceAndDBAndNet(t)
	return inst, db
}

func newTestInstanceAndNet(t *testing.T) (*Instance, *mockNet) {
	inst, _, net := newTestInstanceAndDBAndNet(t)
	return inst, net
}

func TestInstance_createOngoingSwap(t *testing.T) {
	inst, offerDB := newTestInstanceAndDB(t)
	rdb := inst.backend.RecoveryDB().(*backend.MockRecoveryDB)

	one := apd.New(1, 0)
	rate := coins.ToExchangeRate(apd.New(1, 0)) // 100% relayer commission
	offer := types.NewOffer(coins.ProvidesXMR, one, one, rate, types.EthAssetETH)

	s := &pswap.Info{
		ID:             offer.ID,
		Provides:       coins.ProvidesXMR,
		ProvidedAmount: one,
		ExpectedAmount: one,
		ExchangeRate:   rate,
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
