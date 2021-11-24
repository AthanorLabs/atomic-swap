package common

type Environment byte

const (
	Mainnet Environment = iota
	Stagenet
	Development
)

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
