package common

import "errors"

type ProvidesCoin string

var (
	ProvidesXMR ProvidesCoin = "XMR"
	ProvidesETH ProvidesCoin = "ETH"
)

func NewProvidesCoin(s string) (ProvidesCoin, error) {
	switch s {
	case "XMR":
		return ProvidesXMR, nil
	case "ETH":
		return ProvidesETH, nil
	default:
		return "", errors.New("invalid ProvidesCoin")
	}
}
