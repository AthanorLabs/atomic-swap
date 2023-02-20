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
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
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
	msgMu sync.Mutex     // lock needed, as SendSwapMessage is called async from timeout handlers
	msg   common.Message // last value passed to SendSwapMessage
}

func (n *mockNet) LastSentMessage() common.Message {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	return n.msg
}

func (n *mockNet) SendSwapMessage(msg common.Message, _ types.Hash) error {
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
	rdb.EXPECT().PutCounterpartySwapPrivateKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().PutSwapRelayerInfo(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().PutCounterpartySwapKeys(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	rdb.EXPECT().DeleteSwap(gomock.Any()).Return(nil).AnyTimes()

	extendedEC, err := extethclient.NewEthClient(context.Background(), env, ec, pk)
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

	host := NewMockP2pHost(ctrl)

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
	rate := coins.ToExchangeRate(apd.New(1, 0))
	offer := types.NewOffer(coins.ProvidesXMR, one, one, rate, types.EthAssetETH)

	offerDB.EXPECT().PutOffer(offer).Return(nil)
	_, err := inst.offerManager.AddOffer(offer, "", nil)
	require.NoError(t, err)

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
	rdb.EXPECT().GetCounterpartySwapPrivateKey(s.ID).Return(nil, errors.New("some error"))
	rdb.EXPECT().GetContractSwapInfo(s.ID).Return(&db.EthereumSwapInfo{
		StartNumber:     big.NewInt(1),
		ContractAddress: inst.backend.ContractAddr(),
		Swap: &contracts.SwapFactorySwap{
			Timeout0: big.NewInt(1),
			Timeout1: big.NewInt(2),
		},
	}, nil)
	rdb.EXPECT().GetSwapPrivateKey(s.ID).Return(
		sk.SpendKey(), nil,
	)
	offerDB.EXPECT().GetOffer(s.ID).Return(offer, nil)

	err = inst.createOngoingSwap(s)
	require.NoError(t, err)

	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()
	close(inst.swapStates[s.ID].done)
}

func TestInstance_CompleteSwap(t *testing.T) {
	monero.TestBackgroundMineBlocks(t)

	inst, _ := newTestInstanceAndDB(t)
	rdb := inst.backend.RecoveryDB().(*backend.MockRecoveryDB)

	// our keypair
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	id := [32]byte{9, 9, 9}
	rdb.EXPECT().GetSwapPrivateKey(id).Return(kp.SpendKey(), nil)

	// counterparty's keypair
	kpOther, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	rdb.EXPECT().GetCounterpartySwapKeys(id).Return(kpOther.SpendKey().Public(), kpOther.ViewKey(), nil)

	height, err := inst.backend.XMRClient().GetHeight()
	require.NoError(t, err)
	sinfo := &pswap.Info{
		ID:                id,
		MoneroStartHeight: height,
		Status:            types.XMRLocked,
	}
	err = inst.backend.SwapManager().AddSwap(sinfo)
	require.NoError(t, err)

	// the address of the "shared swap wallet"
	address := mcrypto.SumSpendAndViewKeys(
		kp.PublicKeyPair(), kpOther.PublicKeyPair(),
	).Address(common.Development)

	conf := &monero.WalletClientConf{
		Env:                 common.Development,
		WalletFilePath:      path.Join(t.TempDir(), "test-wallet-tcm"),
		MoneroWalletRPCPath: monero.GetWalletRPCDirectory(t),
	}
	err = conf.Fill()
	require.NoError(t, err)

	// mine some xmr to the "shared swap wallet"
	kpAB := pcommon.GetClaimKeypair(
		kp.SpendKey(), kpOther.SpendKey(),
		kp.ViewKey(), kpOther.ViewKey(),
	)
	moneroCli, err := monero.CreateSpendWalletFromKeys(conf, kpAB, 0)
	require.NoError(t, err)
	xmrAmt := coins.StrToDecimal("1")
	pnAmt := coins.MoneroToPiconero(xmrAmt)
	monero.MineMinXMRBalance(t, moneroCli, pnAmt)

	addrRes, err := moneroCli.GetAddress(0)
	require.NoError(t, err)
	require.Equal(t, string(address), addrRes.Address)

	err = inst.completeSwap(sinfo, kpOther.SpendKey())
	require.NoError(t, err)

	balance, err := moneroCli.GetBalance(0)
	require.NoError(t, err)
	require.Equal(t, uint64(0), balance.Balance)
}
