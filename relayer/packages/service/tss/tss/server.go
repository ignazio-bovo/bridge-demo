package tss

import (
	"github.com/Datura-ai/relayer/packages/service/tss/keygen"
	"github.com/Datura-ai/relayer/packages/service/tss/keysign"
	"github.com/Datura-ai/relayer/packages/service/tss/resharing"
)

// Server define the necessary functionality should be provide by a TSS Server implementation
type Server interface {
	Start() error
	Stop()
	GetLocalPeerID() string
	Keygen(req keygen.Request) (keygen.Response, error)
	KeySign(req keysign.Request, msgId string, handleID uint64, isLastInput bool) (keysign.Response, error)
	Keyresharing(req resharing.Request) (resharing.Response, error)
}
