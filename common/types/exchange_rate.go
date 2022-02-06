package types

// ExchangeRate defines an exchange rate between ETH and XMR.
// It is defined as the ratio of ETH:XMR that the node wishes to provide.
// ie. an ExchangeRate of 0.1 means that the node considers 1 ETH = 10 XMR.
type ExchangeRate float64

// ToXMR converts an ether amount to a monero amount with the given exchange rate
func (r ExchangeRate) ToXMR(ethAmount float64) float64 {
	return ethAmount / float64(r)
}

// ToETH converts a moenro amount to an eth amount with the given exchange rate
func (r ExchangeRate) ToETH(xmrAmount float64) float64 {
	return float64(r) * xmrAmount
}
