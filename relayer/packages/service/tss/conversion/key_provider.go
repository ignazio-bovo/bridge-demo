package conversion

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	btcec "github.com/btcsuite/btcd/btcec/v2"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// GetPeerIDFromPubKey get the peer.ID from bech32 format node pub key
func GetPeerIDFromPubKey(pubkey string) (peer.ID, error) {
	peerID, err := peer.Decode(pubkey)
	if err != nil {
		return "", fmt.Errorf("failed to decode peer ID: %w", err)
	}
	return peerID, nil
}

// GetPeerIDsFromPubKeys convert a list of node pub key to their peer.ID
func GetPeerIDsFromPubKeys(pubkeys []string) ([]peer.ID, error) {
	var peerIDs []peer.ID
	for _, item := range pubkeys {
		peerID, err := GetPeerIDFromPubKey(item)
		if err != nil {
			return nil, err
		}
		peerIDs = append(peerIDs, peerID)
	}
	return peerIDs, nil
}

// GetPeerIDMapFromPubKeys convert a list of node pub key to map of peer.ID
func GetPeerIDMapFromPubKeys(pubkeys []string) (map[string]peer.ID, error) {
	peerIDs := make(map[string]peer.ID)
	for _, item := range pubkeys {
		peerID, err := GetPeerIDFromPubKey(item)
		if err != nil {
			return nil, err
		}
		peerIDs[item] = peerID
	}
	return peerIDs, nil
}

// GetPubKeyFromSecp256PubKey convert the given pubkey into a puk key string
func GetPubKeyFromSecp256PubKey(pk []byte) (string, error) {
	if len(pk) == 0 {
		return "", errors.New("empty public key raw bytes")
	}
	pubKey, err := crypto.UnmarshalSecp256k1PublicKey(pk)
	if err != nil {
		return "", fmt.Errorf("fail to unmarshal pub key: %w", err)
	}
	rawBytes, err := pubKey.Raw()
	if err != nil {
		return "", fmt.Errorf("faail to get pub key raw bytes: %w", err)
	}
	return hex.EncodeToString(rawBytes), nil
}

// GetPeerIDs return a slice of peer id
func GetPeerIDs(pubkeys []string, isNewComittee []bool) ([]peer.ID, error) {
	if len(pubkeys) != len(isNewComittee) {
		return nil, fmt.Errorf("invalid lenght, could not create peerIDs")
	}
	var peerIDs []peer.ID
	for _, item := range pubkeys {
		pID, err := GetPeerIDFromPubKey(item)
		if err != nil {
			return nil, fmt.Errorf("fail to get peer id from pubkey(%s):%w", item, err)
		}
		peerIDs = append(peerIDs, pID)
	}
	return peerIDs, nil
}

// GetPubKeysFromPeerIDs given a list of peer ids, and get a list og pub keys.
func GetPubKeysFromPeerIDs(peers []string) ([]string, error) {
	var result []string
	for _, item := range peers {
		pKey, err := GetPubKeyFromPeerID(item)
		if err != nil {
			return nil, fmt.Errorf("fail to get pubkey from peerID: %w", err)
		}
		result = append(result, pKey)
	}
	return result, nil
}

// GetPubKeyFromPeerID extract the pub key from PeerID
func GetPubKeyFromPeerID(pID string) (string, error) {
	peerID, err := peer.Decode(pID)
	if err != nil {
		return "", fmt.Errorf("fail to decode peer id: %w", err)
	}
	pk, err := peerID.ExtractPublicKey()
	if err != nil {
		return "", fmt.Errorf("fail to extract pub key from peer id: %w", err)
	}
	rawBytes, err := pk.Raw()
	if err != nil {
		return "", fmt.Errorf("faail to get pub key raw bytes: %w", err)
	}
	return hex.EncodeToString(rawBytes), nil
}

func GetPriKey(priKeyString string) (crypto.PrivKey, error) {
	priHexBytes, err := base64.StdEncoding.DecodeString(priKeyString)
	if err != nil {
		return nil, fmt.Errorf("fail to decode private key: %w", err)
	}
	rawBytes, err := hex.DecodeString(string(priHexBytes))
	if err != nil {
		return nil, fmt.Errorf("fail to hex decode private key: %w", err)
	}

	priKey, err := crypto.UnmarshalSecp256k1PrivateKey(rawBytes)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal private key: %w", err)
	}
	return priKey, nil
}

func GetPriKeyRawBytes(priKey crypto.PrivKey) ([]byte, error) {
	return priKey.Raw()
}

func CheckKeyOnCurve(pk string) (bool, error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pk)
	if err != nil {
		return false, fmt.Errorf("fail to decode pub key: %w", err)
	}
	// Convert to btcec public key
	btcecPubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false, fmt.Errorf("fail to parse pub key: %w", err)
	}

	// Check if the point is on the curve
	return btcecPubKey.IsOnCurve(), nil
}
