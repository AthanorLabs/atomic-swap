package contracts

// Gas prices for our operations. Most of these are set by the highest value we
// ever see in a test, so you would need to adjust upwards a little to use as a
// gas limit. We use these values to estimate minimum required balances.
const (
	swapCreatorDeployGas = 1004649 // constant, so no "max" prefix
	MaxNewSwapETHGas     = 50589
	MaxNewSwapTokenGas   = 86218
	MaxSetReadyGas       = 31872
	MaxClaimETHGas       = 43349
	MaxClaimTokenGas     = 47522
	MaxRefundETHGas      = 43120
	MaxRefundTokenGas    = 47282
	MaxTokenApproveGas   = 47000 // 46223 with our contract
)
