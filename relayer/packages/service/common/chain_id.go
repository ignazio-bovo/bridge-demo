package common

import (
	"errors"
	"strings"
)

var (
	ETHChainID           = ChainID("ETH")
	SUBTENSORID          = ChainID("SUBTENSOR")
	SigningAlgoSecp256k1 = SigningAlgo("secp256k1")
	SigningAlgoEd25519   = SigningAlgo("ed25519")
)

// SigningAlgo signing algorithm
type SigningAlgo string

// Chain is an alias of string, represent a block chain
type ChainID string

// NewChain create a new Chain and default the siging_algo to Secp256k1
func NewChainID(chainID string) (ChainID, error) {
	chain := ChainID(strings.ToUpper(chainID))
	if err := chain.Validate(); err != nil {
		return chain, err
	}
	return chain, nil
}

// Validate validates chain format, should consist only of uppercase letters
func (c ChainID) Validate() error {
	if len(c) < 3 {
		return errors.New("chain id len is less than 3")
	}
	if len(c) > 10 {
		return errors.New("chain id len is more than 10")
	}
	for _, ch := range c.String() {
		if ch < 'A' || ch > 'Z' {
			return errors.New("chain id can consist only of uppercase letters")
		}
	}
	return nil
}

// Equals compare two chain to see whether they represent the same chain
func (c ChainID) Equals(c2 ChainID) bool {
	return strings.EqualFold(c.String(), c2.String())
}

// IsEmpty is to determinate whether the chain is empty
func (c ChainID) IsEmpty() bool {
	return strings.TrimSpace(c.String()) == ""
}

// String implement fmt.Stringer
func (c ChainID) String() string {
	// convert it to upper case again just in case someone created a ticker via Chain("rune")
	return strings.ToUpper(string(c))
}

// GetSigningAlgo get the signing algorithm for the given chain
func (c ChainID) GetSigningAlgo(isEvm bool) SigningAlgo {
	if isEvm {
		// Only SigningAlgoSecp256k1 is supported for now
		return SigningAlgoSecp256k1
	}
	return SigningAlgoEd25519
}

// GetGasUnits returns name of the gas unit for each chain
func (c ChainID) GetGasUnits() string {
	switch c {
	case ETHChainID:
		return "gwei"
	default:
		return ""
	}
}
