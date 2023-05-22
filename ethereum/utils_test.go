// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"testing"

	"github.com/athanorlabs/atomic-swap/common"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestStage_StageToString(t *testing.T) {
	expectedValues := []string{
		"Invalid",
		"Pending",
		"Ready",
		"Completed",
		"UnknownStageValue(4)",
	}
	for s := byte(0); s < byte(5); s++ {
		require.Equal(t, expectedValues[s], StageToString(s))
	}
}

func TestGetTopic(t *testing.T) {
	refundedTopic := ethcommon.HexToHash("0x007c875846b687732a7579c19bb1dade66cd14e9f4f809565e2b2b5e76c72b4f")
	require.Equal(t, common.GetTopic(RefundedEventSignature), refundedTopic)
}
