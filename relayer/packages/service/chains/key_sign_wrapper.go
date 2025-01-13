package chains

import (
	"errors"
	"fmt"

	"github.com/Datura-ai/relayer/packages/service/common"
	tss "github.com/Datura-ai/relayer/packages/service/tss/tss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// KeySignWrapper is a wrap of private key and also tss instance
type KeySignWrapper struct {
	pubKey        common.PubKey
	tssKeyManager tss.TssKeyManager
	logger        zerolog.Logger
}

// newKeySignWrapper create a new instance of keysign wrapper
func NewKeySignWrapper(pubKey common.PubKey, keyManager tss.TssKeyManager) (*KeySignWrapper, error) {
	return &KeySignWrapper{
		pubKey:        pubKey,
		tssKeyManager: keyManager,
		logger:        log.With().Str("module", "signer").Str("chain", common.ETHChainID.String()).Logger(),
	}, nil
}

// GetPubKey return the public key
func (w *KeySignWrapper) GetPubKey() common.PubKey {
	return w.pubKey
}

// Sign the given transaction
func (w *KeySignWrapper) SignTSS(msgToSign []byte, poolPubKey common.PubKey, msgId string) ([]byte, []byte, error) {
	if msgToSign == nil {
		return nil, nil, errors.New("msgToSign is nil")
	}

	// If pool public key is unavailable
	if poolPubKey.IsEmpty() {
		return nil, nil, errors.New("pool public key is empty")
	}

	sig, originMsg, err := w.signTSS(msgToSign, poolPubKey.String(), msgId)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to sign tx: %w", err)
	}

	return sig, originMsg, nil
}

func (w *KeySignWrapper) signTSS(msgToSign []byte, poolPubKey string, msgId string) ([]byte, []byte, error) {
	sig, originMsg, _, err := w.tssKeyManager.RemoteSign(msgToSign[:], poolPubKey, msgId, false)
	if err != nil || sig == nil {
		return nil, nil, fmt.Errorf("fail to TSS sign: %w", err)
	}

	// add the recovery id at the end
	return sig, originMsg, nil
}
