package xmrmaker

//nolint:lll
//go:generate mockgen -destination=mocks_test.go -package $GOPACKAGE github.com/noot/atomic-swap/protocol/backend Backend
