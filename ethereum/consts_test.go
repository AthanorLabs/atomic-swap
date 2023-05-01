package contracts

// We don't deploy SwapCreator contracts or ERC20 token contracts in swaps, so
// these constants are only compiled in for test files.
const (
	maxSwapCreatorDeployGas = 1062999
	maxTestERC20DeployGas   = 798286 // using long token names or symbols will increase this
)
