package backend

//nolint:lll
//go:generate mockgen -destination=mock_recovery_db.go -package $GOPACKAGE github.com/athanorlabs/atomic-swap/protocol/backend RecoveryDB
