package xmrmaker

//nolint:lll
//go:generate mockgen -destination=mocks_test.go -package $GOPACKAGE github.com/athanorlabs/atomic-swap/protocol/backend Backend
//go:generate mockgen -destination=net_mock_test.go -package $GOPACKAGE github.com/athanorlabs/atomic-swap/net Host
