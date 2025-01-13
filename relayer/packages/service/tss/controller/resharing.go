package controller

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	repository "github.com/Datura-ai/relayer/packages/repository/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/resharing"
	"github.com/Datura-ai/relayer/packages/service/tss/tss"
	"github.com/blang/semver"
)

// Keyresharing is instance to reshare key between peers
type Keyresharing struct {
	keys           []string
	logger         zerolog.Logger
	client         *http.Client
	stateManager   repository.TssStateManager
	server         *tss.TssServer
	currentVersion semver.Version
	lastCheck      time.Time
}

// NewTssReSharingGen create a new instance of Keyresharing which will look after TSS key stuff
func NewTssResharing(keys []string, server *tss.TssServer) (*Keyresharing, error) {
	if keys == nil {
		return nil, fmt.Errorf("keys is nil")
	}
	return &Keyresharing{
		keys:   keys,
		logger: log.With().Str("module", "tss_resharing").Logger(),
		client: &http.Client{
			Timeout: time.Second * 130,
		},
		server: server,
	}, nil
}

// getVersion returns message party version for resharing
func (k *Keyresharing) getVersion() semver.Version {
	requestTime := time.Now()
	k.lastCheck = requestTime
	k.currentVersion, _ = semver.Make(messages.NEWJOINPARTYVERSION)

	return k.currentVersion
}

// ResharingKey send a new Keyresharing request to tss server
func (k *Keyresharing) ResharingKey(keygenBlockHeight int64, keys, newKeys []string) error {
	var resp resharing.Response
	var err error

	// No need to do key gen
	if len(keys) == 0 {
		return nil
	}

	currentVersion := k.getVersion()
	keyGenReq := resharing.NewRequestblock(keygenBlockHeight, currentVersion.String(), keys, newKeys)

	errChan := make(chan error, len(newKeys)+len(keys))
	go func() {
		resp, err = k.server.Keyresharing(keyGenReq)
		if errChan != nil {
			errChan <- err
		}
	}()

	timeout := time.After(4 * time.Minute)

	select {
	case err = <-errChan:
		break
	case <-timeout:
		// wait for the goroutine to finish before closing the channel
		close(errChan)
		k.logger.Error().Msg("tss resharing finished timeout")
		return errors.New("tss resharing finished timeout")
	}

	if err != nil {
		k.logger.Debug().Msgf("key resharing finished with error: %v", err)
		return err
	}

	k.logger.Debug().Msgf("key resharing finished successfully, resp: %v", resp)
	return nil
}
