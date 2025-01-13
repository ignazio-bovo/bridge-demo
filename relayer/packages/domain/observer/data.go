package observer

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// TODO : create a common part for Observer & TSS
type TransferData struct {
	TransferHash       common.Hash
	SourceChainID      uint64
	From               common.Address
	DestinationChainID uint64
	To                 common.Address
	Amount             *big.Int
	Nonce              *big.Int
	Timestamp          time.Time
}

type ConfirmTransferRequest struct {
	Data TransferData
}

type TransferConfirmation struct {
	Data TransferData
}
