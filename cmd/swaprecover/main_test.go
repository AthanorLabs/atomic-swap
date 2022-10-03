package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker"
	"github.com/athanorlabs/atomic-swap/protocol/xmrtaker"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func newTestContext(t *testing.T, description string, flags map[string]any) *cli.Context {
	set := flag.NewFlagSet(description, 0)
	for flag, value := range flags {
		switch v := value.(type) {
		case bool:
			set.Bool(flag, v, "")
		case string:
			set.String(flag, v, "")
		case uint:
			set.Uint(flag, v, "")
		case int64:
			set.Int64(flag, v, "")
		case []string:
			set.Var(&cli.StringSlice{}, flag, "")
		default:
			t.Fatalf("unexpected cli value type: %T", value)
		}
	}

	ctx := cli.NewContext(app, set, nil)

	for flag, value := range flags {
		switch v := value.(type) {
		case bool, uint, int64, string:
			require.NoError(t, ctx.Set(flag, fmt.Sprintf("%v", v)))
		case []string:
			for _, str := range v {
				require.NoError(t, ctx.Set(flag, str))
			}
		default:
			t.Fatalf("unexpected cli value type: %T", value)
		}
	}

	return ctx
}

type mockRecoverer struct{}

func (r *mockRecoverer) WalletFromSharedSecret(_ *mcrypto.PrivateKeyInfo) (mcrypto.Address, error) {
	return mcrypto.Address(""), nil
}

func (r *mockRecoverer) RecoverFromXMRMakerSecretAndContract(b backend.Backend, _ string, xmrmakerSecret,
	contractAddr string, swapID [32]byte, _ contracts.SwapFactorySwap) (*xmrmaker.RecoveryResult, error) {
	return &xmrmaker.RecoveryResult{
		Claimed: true,
	}, nil
}

func (r *mockRecoverer) RecoverFromXMRTakerSecretAndContract(b backend.Backend, _ string, xmrtakerSecret string,
	swapID [32]byte, _ contracts.SwapFactorySwap) (*xmrtaker.RecoveryResult, error) {
	return &xmrtaker.RecoveryResult{
		Claimed: true,
	}, nil
}

func getMockRecoverer(c *cli.Context, env common.Environment) (Recoverer, error) {
	return &mockRecoverer{}, nil
}

func createInfoFile(t *testing.T, kpA, kpB *mcrypto.PrivateKeyPair, contractAddr string) string {
	if kpA == nil && kpB == nil {
		t.Fatal("must provide a secret key")
	}

	infofile := &pcommon.InfoFileContents{}

	if kpA != nil && kpB != nil {
		sk := mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
		kp, err := sk.AsPrivateKeyPair()
		require.NoError(t, err)
		infofile.SharedSwapPrivateKey = kp.Info(common.Development)
	}

	if kpA != nil {
		infofile.PrivateKeyInfo = kpA.Info(common.Development)
	} else if kpB != nil {
		infofile.PrivateKeyInfo = kpB.Info(common.Development)
	}

	infofile.ContractAddress = contractAddr

	bz, err := json.MarshalIndent(infofile, "", "\t")
	require.NoError(t, err)
	filepath := path.Join(t.TempDir(), "test-infofile.txt")
	err = os.WriteFile(filepath, bz, 0600)
	require.NoError(t, err)
	return filepath
}

func createEthPrivKeyFile(t *testing.T, ethKeyHex string) string {
	fileName := path.Join(t.TempDir(), "eth.key")
	err := os.WriteFile(fileName, []byte(ethKeyHex), 0600)
	require.NoError(t, err)
	return fileName
}

func TestRecover_sharedSwapSecret(t *testing.T) {
	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	infoFilePath := createInfoFile(t, kpA, kpB, "")

	c := newTestContext(t,
		"test --xmrtaker with shared swap secret",
		map[string]any{
			flagEnv:                  "dev",
			flagXMRTaker:             true,
			flagInfoFile:             infoFilePath,
			flagMoneroWalletEndpoint: tests.CreateWalletRPCService(t),
		},
	)

	err = runRecover(c)
	require.NoError(t, err)
}

func TestRecover_withXMRMakerSecretAndContract(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	infoFilePath := createInfoFile(t, nil, kp, "0xabcd")

	c := newTestContext(t,
		"test --xmrmaker with contract address and secret",
		map[string]any{
			flagEnv:                  "dev",
			flagXMRMaker:             true,
			flagInfoFile:             infoFilePath,
			flagEthereumPrivKey:      createEthPrivKeyFile(t, common.DefaultPrivKeyXMRMaker),
			flagMoneroWalletEndpoint: tests.CreateWalletRPCService(t),
		},
	)

	inst := &instance{
		getRecovererFunc: getMockRecoverer,
	}
	err = inst.recover(c)
	require.NoError(t, err)
}

func TestRecover_withXMRTakerSecretAndContract(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	infoFilePath := createInfoFile(t, kp, nil, "0xabcd")

	c := newTestContext(t,
		"test --xmrtaker with contract address and secret",
		map[string]any{
			flagEnv:             "dev",
			flagXMRTaker:        true,
			flagInfoFile:        infoFilePath,
			flagEthereumPrivKey: createEthPrivKeyFile(t, common.DefaultPrivKeyXMRTaker),
		},
	)

	inst := &instance{
		getRecovererFunc: getMockRecoverer,
	}
	err = inst.recover(c)
	require.NoError(t, err)
}
