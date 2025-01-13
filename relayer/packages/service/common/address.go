package common

import (
	"fmt"
	"regexp"
	"strings"

	eth "github.com/ethereum/go-ethereum/common"
)

type Address string

const (
	NoAddress   = Address("")
	NoopAddress = Address("noop")
)

var alphaNumRegex = regexp.MustCompile("^[:A-Za-z0-9]*$")

// NewAddress create a new Address. Supports Binance, Bitcoin, and Ethereum
func NewAddress(address string) (Address, error) {
	if len(address) == 0 {
		return NoAddress, nil
	}

	if !alphaNumRegex.MatchString(address) {
		return NoAddress, fmt.Errorf("address format not supported: %s", address)
	}

	// Check is eth address
	if eth.IsHexAddress(address) {
		return Address(address), nil
	}

	return NoAddress, fmt.Errorf("address format not supported: %s", address)
}

// Initiate solana address
func NewAddressSolana(address string) (Address, error) {
	return Address(address), nil
}

// Note that this can have false positives, such as being unable to distinguish between ETH and AVAX.
func (addr Address) IsChain(chain ChainID) bool {
	return strings.HasPrefix(addr.String(), "0x")
}

func (addr Address) Equals(addr2 Address) bool {
	return strings.EqualFold(addr.String(), addr2.String())
}

func (addr Address) IsEmpty() bool {
	return strings.TrimSpace(addr.String()) == ""
}

func (addr Address) IsNoop() bool {
	return addr.Equals(NoopAddress)
}

func (addr Address) String() string {
	return string(addr)
}
