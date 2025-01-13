package resharing

import (
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
)

// Request request to do keygen
type Request struct {
	Keys        []string                    `json:"keys"`
	NewKeys     []string                    `json:"new_key"`
	BlockHeight int64                       `json:"block_height"`
	Version     string                      `json:"tss_version"`
	LocalData   []keygen.LocalPartySaveData `json:"local_data"`
	AddressBook map[peer.ID]p2p.AddrList    `json:"address_book"`
}

// NewRequestblock create a new instance of keygen.Request
func NewRequestblock(blockHeight int64, version string, keys, newKeys []string) Request {
	return Request{
		Keys:        keys,
		NewKeys:     newKeys,
		BlockHeight: blockHeight,
		Version:     version,
	}
}
