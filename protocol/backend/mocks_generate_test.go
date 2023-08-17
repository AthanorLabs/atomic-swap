// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package backend

//go:generate mockgen -destination=mock_recovery_db.go -package $GOPACKAGE github.com/athanorlabs/atomic-swap/protocol/backend RecoveryDB
