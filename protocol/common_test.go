// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package protocol

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common"

	"github.com/stretchr/testify/require"
)

func TestKeysAndProof(t *testing.T) {
	kp, err := GenerateKeysAndProof()
	require.NoError(t, err)

	res, err := VerifyKeysAndProof(
		kp.DLEqProof.Proof(),
		kp.Secp256k1PublicKey,
		kp.PublicKeyPair.SpendKey(),
	)
	require.NoError(t, err)
	require.Equal(t, kp.Secp256k1PublicKey.String(), res.Secp256k1PublicKey.String())
	require.Equal(t, kp.PublicKeyPair.SpendKey().String(), res.Ed25519PublicKey.String())
	require.Equal(t, [32]byte(common.Reverse(kp.PrivateKeyPair.SpendKey().Bytes())), kp.DLEqProof.Secret())
}
