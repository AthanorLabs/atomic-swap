package main

import (
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/protocol/alice"
	"github.com/noot/atomic-swap/protocol/bob"

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

func (r *mockRecoverer) WalletFromSecrets(aliceSecret, bobSecret string) (mcrypto.Address, error) {
	return mcrypto.Address(""), nil
}

func (r *mockRecoverer) RecoverFromBobSecretAndContract(b *bob.Instance, bobSecret, contractAddr string, swapID *big.Int) (*bob.RecoveryResult, error) {
	return &bob.RecoveryResult{
		Claimed: true,
	}, nil
}

func (r *mockRecoverer) RecoverFromAliceSecretAndContract(a *alice.Instance, aliceSecret, contractAddr string, swapID *big.Int) (*alice.RecoveryResult, error) {
	return &alice.RecoveryResult{
		Claimed: true,
	}, nil
}

func getMockRecoverer(c *cli.Context, env common.Environment) (Recoverer, error) {
	return &mockRecoverer{}, nil
}

func TestRecover_withBothSecrets(t *testing.T) {
	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	c := newTestContext(t,
		"test --alice-secret and --bob-secret",
		[]string{flagAliceSecret, flagBobSecret},
		[]interface{}{
			kpA.SpendKey().Hex(),
			kpB.SpendKey().Hex(),
		},
	)

	err = runRecover(c)
	require.NoError(t, err)
}

func TestRecover_withBobSecretAndContract(t *testing.T) {
	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	c := newTestContext(t,
		"test --contract-addr and --bob-secret",
		[]string{flagContractAddr, flagBobSecret},
		[]interface{}{
			"0xabcd",
			kp.SpendKey().Hex(),
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

	c := newTestContext(t,
		"test --contract-addr and --alice-secret",
		[]string{flagContractAddr, flagAliceSecret},
		[]interface{}{
			"0xabcd",
			kp.SpendKey().Hex(),
		},
	)

	inst := &instance{
		getRecovererFunc: getMockRecoverer,
	}
	err = inst.recover(c)
	require.NoError(t, err)
}
