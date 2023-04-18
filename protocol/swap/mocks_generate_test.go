// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package swap

//nolint:lll
//go:generate mockgen -destination=mocks.go -package $GOPACKAGE github.com/athanorlabs/atomic-swap/protocol/swap Database
