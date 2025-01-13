package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/blang/semver"

	"github.com/Datura-ai/relayer/packages/service/tss/keygen"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/tss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Keygen
type Keygen struct {
	logger         zerolog.Logger
	client         *http.Client
	server         *tss.TssServer
	currentVersion semver.Version
	lastCheck      time.Time
}

// NewTssKeygen create a new instance of TssKeyGen which will look after TSS key stuff
func NewTssKeygen(server *tss.TssServer) (*Keygen, error) {
	return &Keygen{
		logger: log.With().Str("module", "tss_keygen").Logger(),
		client: &http.Client{
			Timeout: time.Second * 130,
		},
		server: server,
	}, nil
}

func (kg *Keygen) getVersion() semver.Version {
	requestTime := time.Now()
	kg.lastCheck = requestTime
	kg.currentVersion, _ = semver.Make(messages.NEWJOINPARTYVERSION)

	return kg.currentVersion
}

func (kg *Keygen) GenerateNewKey(pKeys []string) (string, error) {
	var resp keygen.Response
	var err error

	// No need to do key gen
	if len(pKeys) == 0 {
		return "", nil
	}

	keys := pKeys
	keyGenReq := keygen.Request{
		Keys:        keys,
		BlockHeight: 1,
		Version:     kg.getVersion().String(),
	}

	currentVersion := kg.getVersion()
	keyGenReq.Version = currentVersion.String()

	errChan := make(chan error, 1)

	go func() {
		resp, err = kg.server.Keygen(keyGenReq)
		errChan <- err
	}()

	timeout := time.After(2 * time.Minute)

	select {
	case err = <-errChan:
		break
	case <-timeout:
		// wait for the goroutine to finish before closing the channel
		close(errChan)

		kg.logger.Error().Msg("tss key keygen timeout")
		return "", errors.New("tss key keygen timeout")
	}

	if err != nil {
		kg.logger.Debug().Msgf("key keygen finished with error: %v", err)
		return "", err
	}

	kg.logger.Debug().Msgf("keygen finished successfully, resp: %v", resp.PubKey)

	// TODO later on KIMANode need to have both secp256k1 key and ed25519
	return resp.PubKey, nil
}
