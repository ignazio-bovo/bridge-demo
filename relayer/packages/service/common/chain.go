package common

// Chain struct
type Chain struct {
	ChainID    ChainID
	IsEvm      bool
	IsDisabled bool
}

// Chains to manage chain data
type Chains []Chain

func NewChains(cfg []ChainConfiguration) Chains {
	chains := make([]Chain, len(cfg))
	for idx, c := range cfg {
		chains[idx] = Chain{
			ChainID: c.ChainID,
			IsEvm:   c.IsEvm,
		}
	}
	return chains
}

func (chains Chains) Get(idx int) *Chain {
	return &chains[idx]
}

// Has check whether chain c is in the list
func (chains Chains) Has(c ChainID) bool {
	chainID := c
	for _, ch := range chains {
		if ch.ChainID.Equals(chainID) {
			return true
		}
	}
	return false
}

// Distinct return a distinct set of chains, no duplicates
func (chains Chains) Distinct() (newChains []Chain) {
	for _, chain := range chains {
		if !chains.Has(chain.ChainID) {
			newChains = append(newChains, chain)
		}
	}
	return newChains
}

func (chains Chains) Strings() []string {
	strings := make([]string, len(chains))
	for i, c := range chains {
		strings[i] = c.ChainID.String()
	}
	return strings
}

func (chains Chains) GetChainIdx(chainID ChainID) int {
	for idx, chain := range chains {
		if chainID == chain.ChainID {
			return idx
		}
	}
	return 0
}

func (chains Chains) GetSize() int {
	return len(chains)
}
