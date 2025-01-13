package chains

import (
	types "github.com/Datura-ai/relayer/packages/domain/common"
	"github.com/Datura-ai/relayer/packages/service/common"
)

type ChainClient interface {
	SignTx(tx types.BridgeTxData, poolPubKey common.PubKey, msgId string) ([]byte, error)
	BroadcastTx(txOutItem types.BridgeTxData, hexTx []byte) (string, error)
	Start()
	Stop()
}
