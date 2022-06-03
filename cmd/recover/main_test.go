package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/protocol/alice"
	"github.com/noot/atomic-swap/protocol/bob"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func newTestContext(t *testing.T, description string, flags []string, values []interface{}) *cli.Context {
	require.Equal(t, len(flags), len(values))

	set := flag.NewFlagSet(description, 0)
	for i := range values {
		switch v := values[i].(type) {
		case bool:
			set.Bool(flags[i], v, "")
		case string:
			set.String(flags[i], v, "")
		case uint:
			set.Uint(flags[i], v, "")
		case int64:
			set.Int64(flags[i], v, "")
		case []string:
			set.Var(&cli.StringSlice{}, flags[i], "")
		default:
			t.Fatalf("unexpected cli value type: %T", values[i])
		}
	}

	ctx := cli.NewContext(app, set, nil)
	var (
		err error
		i   int
	)

	for i = range values {
		switch v := values[i].(type) {
		case bool:
			err = ctx.Set(flags[i], strconv.FormatBool(v))
		case string:
			err = ctx.Set(flags[i], values[i].(string))
		case uint:
			err = ctx.Set(flags[i], strconv.Itoa(int(values[i].(uint))))
		case int64:
			err = ctx.Set(flags[i], strconv.Itoa(int(values[i].(int64))))
		case []string:
			for _, str := range values[i].([]string) {
				err = ctx.Set(flags[i], str)
				require.NoError(t, err)
			}
		default:
			t.Fatalf("unexpected cli value type: %T", values[i])
		}
	}

	require.NoError(t, err, fmt.Sprintf("failed to set cli flag: %T, err: %s", flags[i], err))
	return ctx
}

type mockRecoverer struct{}

func (r *mockRecoverer) WalletFromSharedSecret(_ *mcrypto.PrivateKeyInfo) (mcrypto.Address, error) {
	return mcrypto.Address(""), nil
}

func (r *mockRecoverer) RecoverFromBobSecretAndContract(b *bob.Instance, bobSecret, contractAddr string,
	swapID [32]byte, _ swapfactory.SwapFactorySwap) (*bob.RecoveryResult, error) {
	return &bob.RecoveryResult{
		Claimed: true,
	}, nil
}

func (r *mockRecoverer) RecoverFromAliceSecretAndContract(a *alice.Instance, aliceSecret string,
	swapID [32]byte, _ swapfactory.SwapFactorySwap) (*alice.RecoveryResult, error) {
	return &alice.RecoveryResult{
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
	filepath := os.TempDir() + "/test-infofile.txt"
	err = ioutil.WriteFile(filepath, bz, os.ModePerm)
	require.NoError(t, err)
	return filepath
}

func TestRecover_sharedSwapSecret(t *testing.T) {
	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	infoFilePath := createInfoFile(t, kpA, kpB, "")

	c := newTestContext(t,
		"test --xmrtaker with shared swap secret",
		[]string{flagXMRTaker, flagInfoFile},
		[]interface{}{
			true,
			infoFilePath,
		},
	)

	err = runRecover(c)
	require.NoError(t, err)
}

func TestRecover_withBobSecretAndContract(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	infoFilePath := createInfoFile(t, nil, kp, "0xabcd")

	c := newTestContext(t,
		"test --xmrmaker with contract address and secret",
		[]string{flagXMRMaker, flagInfoFile},
		[]interface{}{
			true,
			infoFilePath,
		},
	)

	inst := &instance{
		getRecovererFunc: getMockRecoverer,
	}
	err = inst.recover(c)
	require.NoError(t, err)
}

func TestRecover_withAliceSecretAndContract(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	infoFilePath := createInfoFile(t, kp, nil, "0xabcd")

	c := newTestContext(t,
		"test --xmrtaker with contract address and secret",
		[]string{flagXMRTaker, flagInfoFile},
		[]interface{}{
			true,
			infoFilePath,
		},
	)

	inst := &instance{
		getRecovererFunc: getMockRecoverer,
	}
	err = inst.recover(c)
	require.NoError(t, err)
}
