package tss

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	bkeygen "github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
	db "go.mills.io/bitcask/v2"
)

type TssStateManager interface {
	GetPeerAddrList() p2p.AddrList
	RetrievePeers() ([]peer.ID, error)
	AddPeer(peerId peer.ID)

	ID() (*peer.ID, error)
	SavePrivKey(privKey crypto.PrivKey, override bool) bool
	GetPrivKey() ([]byte, error)

	GetSignerPubKeys() ([]string, error)

	UpdatePeers(peerIds []peer.ID) error

	SavePreParams(preParams *bkeygen.LocalPreParams) error
	GetPreParams() (*bkeygen.LocalPreParams, error)

	Close()
	GetTssPubKey() (string, error)
	SaveTssPubKey(pubKey string) error

	AddPendingMsgID(msgID string) error
	RemovePendingMsgID(msgID string) error
	IsExistPendingMsgID(msgID string) bool
}

type tssStateManager struct {
	store *db.Bitcask
}

func NewTssStateManager(home string) TssStateManager {
	option := []db.Option{
		db.WithMaxKeySize(1024),
		db.WithMaxValueSize(1024 * 1024),
		db.WithAutoRecovery(true),
		db.WithDirMode(0755),
		db.WithFileMode(0644),
	}
	store, _ := db.Open(fmt.Sprintf("%s%s", home, "/tss_db"), option...)
	//defer store.Close()
	return &tssStateManager{
		store: store,
	}
}

func DefaultTssStateManager() TssStateManager {
	// start state manager here
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get home directory")
		return nil
	}
	stateManager := NewTssStateManager(fmt.Sprintf("%s/.datura", homeDir))
	return stateManager
}

func (ts *tssStateManager) RetrievePeers() ([]peer.ID, error) {
	peerData, err := ts.store.Get([]byte(TSS_PEER))
	if err != nil {
		log.Error().Err(err).Msg("Tss: Retrieve peers failed")
		return []peer.ID{}, nil
	}

	var peerIDs []peer.ID
	if err := json.Unmarshal(peerData, &peerIDs); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal peer IDs")
		return nil, err
	}
	// Sort the peer IDs
	sort.Slice(peerIDs, func(i, j int) bool {
		return peerIDs[i].String() < peerIDs[j].String()
	})
	return peerIDs, nil
}

func (ts *tssStateManager) AddPeer(peerId peer.ID) {
	// Retrieve existing peers
	existingPeers, err := ts.RetrievePeers()
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve existing peers")
		return
	}

	// Check if the peer already exists
	for _, existingPeer := range existingPeers {
		if existingPeer == peerId {
			// Peer already exists, no need to add
			return
		}
	}

	// Add the new peer
	updatedPeers := append(existingPeers, peerId)

	// Marshal the updated peer list
	peerData, err := json.Marshal(updatedPeers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal updated peer list")
		return
	}

	// Save the updated peer list to the store
	if err := ts.store.Put(TSS_PEER, peerData); err != nil {
		log.Error().Err(err).Msg("Failed to save updated peer list to database")
		return
	}

	log.Debug().Str("peerId", peerId.String()).Msg("Added new peer to the store")
}

func (ts *tssStateManager) RemovePeer(peerId peer.ID) {
	// Retrieve existing peers
	existingPeers, err := ts.RetrievePeers()
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve existing peers")
		return
	}

	// Create a new slice to hold the updated peer list
	var updatedPeers []peer.ID

	// Iterate through existing peers and exclude the one to be removed
	for _, existingPeer := range existingPeers {
		if existingPeer != peerId {
			updatedPeers = append(updatedPeers, existingPeer)
		}
	}

	// Marshal the updated peer list
	peerData, err := json.Marshal(updatedPeers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal updated peer list")
		return
	}

	// Save the updated peer list to the store
	if err := ts.store.Put(TSS_PEER, peerData); err != nil {
		log.Error().Err(err).Msg("Failed to save updated peer list to database")
		return
	}

	log.Debug().Str("peerId", peerId.String()).Msg("Removed peer from the store")
}

func (ts *tssStateManager) UpdatePeers(peerIds []peer.ID) error {
	peerData, err := json.Marshal(peerIds)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal updated peer list")
		return err
	}

	if err := ts.store.Put([]byte(TSS_PEER), peerData); err != nil {
		log.Error().Err(err).Msg("Failed to save updated peer list to database")
		return err
	}

	log.Debug().Msg("Updated peer list in the store")
	return nil
}

func (ts *tssStateManager) ID() (*peer.ID, error) {
	idBytes, err := ts.store.Get(TSS_ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve TSS ID from store")
		return nil, err
	}

	if len(idBytes) == 0 {
		log.Warn().Msg("TSS ID not found in store")
		return nil, err
	}

	id, err := peer.IDFromBytes(idBytes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert bytes to peer.ID")
		return nil, err
	}
	return &id, nil
}

func (ts *tssStateManager) SavePrivKey(privKey crypto.PrivKey, override bool) bool {
	// Check if ID already exists
	_, err := ts.ID()
	if err == nil && !override {
		log.Warn().Msg("TSS ID already exists and override is false. Not saving new ID.")
		return false
	}

	// Get raw bytes of the private key
	privKeyRaw, err := privKey.Raw() // Get the 32-byte raw private key data
	if err != nil {
		log.Error().Err(err).Msg("Failed to get raw bytes of private key")
		return false
	}

	// Save raw private key to the store
	if err := ts.store.Put(TSS_KEY, privKeyRaw); err != nil {
		log.Error().Err(err).Msg("Failed to save private key to database")
		return false
	}

	// Get peer.ID from the private key and save it
	id, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get peer ID from private key")
		return false
	}
	idBytes, err := id.Marshal()
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert peer.ID to bytes")
		return false
	}

	if err := ts.store.Put(TSS_ID, idBytes); err != nil {
		log.Error().Err(err).Msg("Failed to save TSS ID to database")
		return false
	}

	log.Debug().Str("id", id.String()).Msg("Saved TSS ID and private key to the store")
	return true
}

func (ts *tssStateManager) Close() {
	ts.store.Close()
}

func (ts *tssStateManager) GetPeerAddrList() p2p.AddrList {
	peerAddrData, err := ts.store.Get(TSS_PEER_ADDR)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("DB: %s", err.Error()))
		log.Error().Msg("Tss:Failed to retrieve peer Address from database")
		return nil
	}
	var peerAddrList p2p.AddrList
	if err := json.Unmarshal(peerAddrData, &peerAddrList); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal peer IDs")
		return nil
	}
	return peerAddrList
}

func (ts *tssStateManager) GetPrivKey() ([]byte, error) {
	// Retrieve raw private key bytes from the store
	return ts.store.Get(TSS_KEY)
}

func (ts *tssStateManager) GetSignerPubKeys() ([]string, error) {
	// Implement the logic to retrieve signer public keys
	// For now, return an empty slice and nil error as a placeholder
	peerIDs, err := ts.RetrievePeers()
	if err != nil {
		return nil, err
	}
	pubKeys := []string{}
	for _, peerID := range peerIDs {
		pubKeys = append(pubKeys, peerID.String())
	}
	log.Info().Msgf("pubKeys: %v", pubKeys)
	if err != nil {
		return nil, err
	}
	return pubKeys, nil
}

func (ts *tssStateManager) GetTssPubKey() (string, error) {
	pubKey, err := ts.store.Get(TSS_PUB_KEY)
	if err != nil {
		return "", err
	}
	return string(pubKey), nil
}

func (ts *tssStateManager) SaveTssPubKey(pubKey string) error {
	return ts.store.Put(TSS_PUB_KEY, []byte(pubKey))
}

func (ts *tssStateManager) SavePreParams(preParams *bkeygen.LocalPreParams) error {
	// Serialize preParams to JSON
	preParamsData, err := json.Marshal(preParams)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal preParams")
		return err
	}
	return ts.store.Put(TSS_PRE_PARAMS, preParamsData)
}

func (ts *tssStateManager) GetPreParams() (*bkeygen.LocalPreParams, error) {
	preParamsData, err := ts.store.Get(TSS_PRE_PARAMS)
	if err != nil {
		return nil, err
	}
	var preParams bkeygen.LocalPreParams
	if err := json.Unmarshal(preParamsData, &preParams); err != nil {
		return nil, err
	}
	return &preParams, nil
}

func (ts *tssStateManager) AddPendingMsgID(msgID string) error {
	return ts.store.Put([]byte(msgID), []byte(msgID))
}

func (ts *tssStateManager) RemovePendingMsgID(msgID string) error {
	return ts.store.Delete([]byte(msgID))
}

func (ts *tssStateManager) IsExistPendingMsgID(msgID string) bool {
	_, err := ts.store.Get([]byte(msgID))
	if err != nil {
		return false
	}
	return true
}
