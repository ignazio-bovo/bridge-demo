package common

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	secp256k1 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/libp2p/go-libp2p/core/peer"

	eth "github.com/ethereum/go-ethereum/crypto"
)

// PubKey used in kimachain, it should be bech32 encoded string
// thus it will be something like
type (
	PubKey  string
	PubKeys []PubKey
)

// PubKeySet contains two pub keys , secp256k1 and ed25519
type PubKeySet struct {
	Secp256k1 PubKey `protobuf:"bytes,1,opt,name=secp256k1,proto3,casttype=PubKey" json:"secp256k1,omitempty"`
	Ed25519   PubKey `protobuf:"bytes,2,opt,name=ed25519,proto3,casttype=PubKey" json:"ed25519,omitempty"`
}

var EmptyPubKey PubKey

var EmptyPubKeySet PubKeySet

// NewPubKey create a new instance of PubKey
// key is bech32 encoded string
func NewPubKey(key string) (PubKey, error) {
	if len(key) == 0 {
		return EmptyPubKey, nil
	}
	return PubKey(key), nil
}

// Equals check whether two are the same
func (pubKey PubKey) Equals(pubKey1 PubKey) bool {
	return pubKey == pubKey1
}

// IsEmpty to check whether it is empty
func (pubKey PubKey) IsEmpty() bool {
	return len(pubKey) == 0
}

// String stringer implementation
func (pubKey PubKey) String() string {
	return string(pubKey)
}

// EVMPubkeyToAddress converts a pubkey of an EVM chain to the corresponding address
func (pubKey PubKey) EVMPubkeyToAddress() (Address, error) {
	pk, err := peer.Decode(string(pubKey))
	if err != nil {
		return NoAddress, err
	}
	peerPubKey, err := pk.ExtractPublicKey()
	if err != nil {
		return NoAddress, err
	}
	peerPubKeyRaw, err := peerPubKey.Raw()
	if err != nil {
		return NoAddress, err
	}
	pub, err := secp256k1.ParsePubKey(peerPubKeyRaw)
	if err != nil {
		return NoAddress, err
	}
	str := strings.ToLower(eth.PubkeyToAddress(*pub.ToECDSA()).String())
	return NewAddress(str)
}

// GetAddress will return an address for the given chain
func (pubKey PubKey) GetAddress(chain ChainID) (Address, error) {
	if chain.String() == "" {
		return NoAddress, nil
	}

	return pubKey.EVMPubkeyToAddress()
}

// MarshalJSON to Marshals to JSON using Bech32
func (pubKey PubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(pubKey.String())
}

// UnmarshalJSON to Unmarshal from JSON assuming Bech32 encoding
func (pubKey *PubKey) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	pk, err := NewPubKey(s)
	if err != nil {
		return err
	}
	*pubKey = pk
	return nil
}

func (pks PubKeys) Valid() error {
	for _, pk := range pks {
		if _, err := NewPubKey(pk.String()); err != nil {
			return err
		}
	}
	return nil
}

func (pks PubKeys) Contains(pk PubKey) bool {
	for _, p := range pks {
		if p.Equals(pk) {
			return true
		}
	}
	return false
}

// Equals check whether two pub keys are identical
func (pks PubKeys) Equals(newPks PubKeys) bool {
	if len(pks) != len(newPks) {
		return false
	}

	source := make(PubKeys, len(pks))
	dest := make(PubKeys, len(newPks))
	copy(source, pks)
	copy(dest, newPks)

	// sort both lists
	sort.Slice(source[:], func(i, j int) bool {
		return source[i].String() < source[j].String()
	})
	sort.Slice(dest[:], func(i, j int) bool {
		return dest[i].String() < dest[j].String()
	})
	for i := range source {
		if !source[i].Equals(dest[i]) {
			return false
		}
	}
	return true
}

// String implement stringer interface
func (pks PubKeys) String() string {
	strs := make([]string, len(pks))
	for i := range pks {
		strs[i] = pks[i].String()
	}
	return strings.Join(strs, ", ")
}

func (pks PubKeys) Strings() []string {
	allStrings := make([]string, len(pks))
	for i, pk := range pks {
		allStrings[i] = pk.String()
	}
	return allStrings
}

// NewPubKeySet create a new instance of PubKeySet , which contains two keys
func NewPubKeySet(secp256k1, ed25519 PubKey) PubKeySet {
	return PubKeySet{
		Secp256k1: secp256k1,
		Ed25519:   ed25519,
	}
}

// IsEmpty will determinate whether PubKeySet is an empty
func (pks PubKeySet) IsEmpty() bool {
	return pks.Secp256k1.IsEmpty() || pks.Ed25519.IsEmpty()
}

// Equals check whether two PubKeySet are the same
func (pks PubKeySet) Equals(pks1 PubKeySet) bool {
	return pks.Ed25519.Equals(pks1.Ed25519) && pks.Secp256k1.Equals(pks1.Secp256k1)
}

func (pks PubKeySet) Contains(pk PubKey) bool {
	return pks.Ed25519.Equals(pk) || pks.Secp256k1.Equals(pk)
}

// String implement fmt.Stinger
func (pks PubKeySet) String() string {
	return fmt.Sprintf(`
	secp256k1: %s
	ed25519: %s
`, pks.Secp256k1.String(), pks.Ed25519.String())
}
