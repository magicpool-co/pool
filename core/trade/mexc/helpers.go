package mexc

import (
	"fmt"
	"strings"
)

func formatChain(chain string) string {
	chain = strings.ToUpper(chain)
	switch chain {
	case "FLUX":
		return "FLUX1"
	default:
		return chain
	}
}

func unformatChain(chain string) string {
	chain = strings.ToUpper(chain)
	switch chain {
	case "FLUX1":
		return "FLUX"
	default:
		return chain
	}
}

func chainToNetwork(chain string) string {
	chain = strings.ToUpper(chain)
	switch chain {
	case "ETH":
		return "ERC20"
	case "KAS":
		return "KASPA"
	case "FLUX":
		return "ZEL"
	case "FIRO":
		return "XZC"
	default:
		return chain
	}
}

func networkToChain(network string) string {
	network = strings.ToUpper(network)
	switch network {
	case "ERC20":
		return "ETH"
	case "KASPA":
		return "KAS"
	case "ZEL", "FLUX1":
		return "FLUX"
	case "XZC":
		return "FIRO"
	default:
		return network
	}
}

func parseIncrement(increment int) (int, error) {
	switch increment {
	case 1:
		return 1e1, nil
	case 2:
		return 1e2, nil
	case 3:
		return 1e3, nil
	case 4:
		return 1e4, nil
	case 5:
		return 1e5, nil
	case 6:
		return 1e6, nil
	case 7:
		return 1e7, nil
	case 8:
		return 1e8, nil
	default:
		return 1, fmt.Errorf("unknown increment %d", increment)
	}
}
