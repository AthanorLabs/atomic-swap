// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package backend

//nolint:lll
//go:generate mockgen -destination=mock_recovery_db.go -package $GOPACKAGE github.com/athanorlabs/atomic-swap/protocol/backend RecoveryDB
