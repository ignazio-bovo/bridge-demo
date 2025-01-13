package tss

import (
	"crypto/rand"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func GenerateNewPeerID() (peer.ID, crypto.PrivKey, error) {
	// Generate a new Ed25519 key pair
	privateKey, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		return "", nil, err
	}
	// Create a new peer ID from the public key
	peerID, err := peer.IDFromPrivateKey(privateKey)
	if err != nil {
		return "", nil, err
	}

	return peerID, privateKey, nil
}

type PeerMsgType string

const (
	KEYGEN_MSG_TYPE  PeerMsgType = "keygen"
	KEYSIGN_MSG_TYPE PeerMsgType = "keysign"
)
