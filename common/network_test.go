package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEnv(t *testing.T) {
	expected := map[string]Environment{
		"mainnet":  Mainnet,
		"stagenet": Stagenet,
		"dev":      Development,
	}

	for strVal, expectedResult := range expected {
		env, err := NewEnv(strVal)
		require.NoError(t, err)
		require.Equal(t, expectedResult, env)
		require.Equal(t, strVal, env.String())
		require.NotNil(t, ConfigDefaultsForEnv(env))
	}
}

func TestNewEnv_fail(t *testing.T) {
	_, err := NewEnv("goerli")
	require.ErrorContains(t, err, "unknown")
}
