package tss

import (
	"fmt"
	"os"

	domain "github.com/Datura-ai/relayer/packages/domain/tss"
	repository "github.com/Datura-ai/relayer/packages/repository/tss"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
)

func GenerateKey() {
	// start state manager here
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get home directory")
		return
	}
	stateManager := repository.NewTssStateManager(fmt.Sprintf("%s/.datura", homeDir))
	defer stateManager.Close()
	// check peerId is existing or not
	peerId, err := stateManager.ID()
	if err != nil {
		log.Debug().Msg("Peer ID did not exist")
		// generate new ID
		peerId = GeneratePeerID(stateManager)
	}
	log.Info().Str("peerId", peerId.String()).Msg("Peer ID")
}

func GeneratePeerID(stateManager repository.TssStateManager) *peer.ID {
	newID, privKey, err := domain.GenerateNewPeerID()
	if err != nil {
		panic("Failed to generate new Peer ID")
	}
	success := stateManager.SavePrivKey(privKey, true)
	if !success {
		panic("Failed to generate new Peer ID")
	}
	log.Info().Str("newPeerID", newID.String()).Msg("Generated and saved new Peer ID")
	return &newID
}
