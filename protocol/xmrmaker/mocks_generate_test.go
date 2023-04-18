// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

//go:generate mockgen -destination=mock_net_test.go -package $GOPACKAGE github.com/athanorlabs/atomic-swap/net P2pHost
