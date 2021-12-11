package common

// Environment represents the environment the swap will run in (ie. mainnet, stagenet, or development)
type Environment byte

const (
	Mainnet Environment = iota //nolint
	Stagenet
	Development
)

// String ...
func (env Environment) String() string {
	switch env {
	case Mainnet:
		return "mainnet"
	case Stagenet:
		return "stagenet"
	case Development:
		return "development"
	}

	return "unknown"
}
