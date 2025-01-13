package common

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type RoutingMessage struct {
	TxData     BridgeTxData
	DstChainId uint64
}

type TokenMetadata struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

type BridgeTxData struct {
	To            common.Address `json:"to"`
	Amount        *big.Int       `json:"amount"`
	Nonce         *big.Int       `json:"nonce"`
	TokenKey      common.Hash    `json:"tokenKey"`
	SrcChainId    uint64         `json:"srcChainId"`
	TokenMetadata TokenMetadata  `json:"tokenMetadata"`
}

func (t *BridgeTxData) GetMsgId() string {
	//Let's make hash based on BridgeTxData
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d-%d", t.SrcChainId, t.Nonce)))
	return hex.EncodeToString(hash.Sum(nil))
}

func (t *BridgeTxData) BuildReleaseTx() []byte {
	//TODO: build the release tx at bridge contract
	// Prepare broadcast raw Tx data
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil
	}
	return bytes
}

func (t *BridgeTxData) BuildReleaseResultTx() []byte {
	//TODO: build the release result tx at bridge contract
	// Prepare broadcast raw Tx data
	return []byte{}
}

func (t *BridgeTxData) ToBytes() []byte {
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil
	}
	return bytes
}

func (t *BridgeTxData) FromBytes(bytes []byte) error {
	data := &BridgeTxData{}
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return err
	}
	*t = *data
	return nil
}
