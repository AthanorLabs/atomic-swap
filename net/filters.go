package net

import (
	"net"

	ma "github.com/multiformats/go-multiaddr"
)

func newPrivateIPFilters() (privateIPs *ma.Filters, err error) {
	privateCIDRs := []string{
		"10.0.0.0/8",
		"127.0.0.1/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10",
		"198.18.0.0/15",
		"169.254.0.0/16",
	}
	privateIPs = ma.NewFilters()
	for _, cidr := range privateCIDRs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return privateIPs, err
		}
		privateIPs.AddFilter(*ipnet, ma.ActionDeny)
	}
	return
}

var (
	privateIPs *ma.Filters
	_          = privateIPs // TODO: remove before merge
)

func init() {
	var err error
	privateIPs, err = newPrivateIPFilters()
	if err != nil {
		log.Panic(err)
	}
}
