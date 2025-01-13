package conversion

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/bnb-chain/tss-lib/crypto"
	btss "github.com/bnb-chain/tss-lib/tss"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/edwards/v2"
	crypto2 "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/Datura-ai/relayer/packages/service/tss/messages"
)

const NEW_PARTY_MONIKER_SUFFIX = "P"

// GetPeerIDFromSecp256PubKey convert the given pubkey into a peer.ID
func GetPeerIDFromSecp256PubKey(pk []byte) (peer.ID, error) {
	if len(pk) == 0 {
		return "", errors.New("empty public key raw bytes")
	}

	ppk, err := crypto2.UnmarshalSecp256k1PublicKey(pk)
	if err != nil && !strings.Contains(err.Error(), "invalid length: 34") {
		return "", fmt.Errorf("fail to convert pubkey to the crypto pubkey used in libp2p: %w", err)
	}

	return peer.IDFromPublicKey(ppk)
}

// GetPeerIDFromPartyID returns peer ID and isNewCommittee boolean from partyID
func GetPeerIDFromPartyID(partyID *btss.PartyID) (peer.ID, error) {
	if partyID == nil || !partyID.ValidateBasic() {
		return "", errors.New("invalid partyID")
	}

	pkBytes := partyID.KeyInt().Bytes()
	peerID, err := GetPeerIDFromSecp256PubKey(pkBytes)
	if err != nil {
		return "", err
	}

	return peerID, nil
}

// PartyIDtoPubKey returns pubkey from partyID
func PartyIDtoPubKey(party *btss.PartyID) (string, error) {
	if party == nil || !party.ValidateBasic() {
		return "", errors.New("invalid party")
	}
	partyKeyBytes := party.GetKey()
	return hex.EncodeToString(partyKeyBytes), nil
}

// AccPubKeysFromPartyIDs return public key list from partyID list
func AccPubKeysFromPartyIDs(partyIDs []string, partyIDMap map[string]*btss.PartyID) ([]string, error) {
	pubKeys := make([]string, len(partyIDs))
	for idx, partyID := range partyIDs {
		blameParty, ok := partyIDMap[partyID]
		if !ok {
			return nil, errors.New("cannot find the blame party")
		}
		blamedPubKey, err := PartyIDtoPubKey(blameParty)
		if err != nil {
			return nil, err
		}

		pubKeys[idx] = blamedPubKey
	}
	return pubKeys, nil
}

// SetupPartyIDMap returns partyID map from an array
func SetupPartyIDMap(partiesID []*btss.PartyID) map[string]*btss.PartyID {
	partyIDMap := make(map[string]*btss.PartyID)
	for _, id := range partiesID {
		partyIDMap[id.Id] = id
	}
	return partyIDMap
}

// GetPeersID returns peer list from partyIDs map
func GetPeersID(partyIDtoP2PID map[string]peer.ID, localPeerID string) []peer.ID {
	if partyIDtoP2PID == nil {
		return nil
	}
	peerIDs := make([]peer.ID, 0, len(partyIDtoP2PID)-1)
	for _, value := range partyIDtoP2PID {
		if value.String() == localPeerID {
			continue
		}
		peerIDs = append(peerIDs, value)
	}
	return peerIDs
}

// SetupIDMaps sets partyID in map
func SetupIDMaps(parties map[string]*btss.PartyID, partyIDtoP2PID map[string]peer.ID) error {
	for id, party := range parties {
		peerID, err := GetPeerIDFromPartyID(party)
		if err != nil {
			return err
		}
		partyIDtoP2PID[id] = peerID
	}
	return nil
}

// GetParties returns partyID list and local partyID if the array contains it
func GetParties(keys []string, localPartyKey string, isNewCommittee bool) ([]*btss.PartyID, *btss.PartyID, error) {
	var localPartyID *btss.PartyID
	var unSortedPartiesID []*btss.PartyID
	for idx, item := range keys {
		// Set up the parameters
		// Note: The `id` and `moniker` fields are for convenience to allow you to easily track participants.
		// The `id` should be a unique string representing this party in the network and `moniker` can be anything (even left blank).
		// The `uniqueKey` is a unique identifying key for this peer (such as its p2p public key) as a big.Int.
		partyID, err := NewPartyID(idx, item, isNewCommittee)
		if err != nil {
			return nil, nil, err
		}
		if item == localPartyKey {
			localPartyID = partyID
		}
		partyID.Index = idx
		unSortedPartiesID = append(unSortedPartiesID, partyID)
	}
	partiesID := btss.SortPartyIDs(unSortedPartiesID)
	return partiesID, localPartyID, nil
}

// NewPartyID returns a new partyID
func NewPartyID(idx int, item string, isNewCommittee bool) (*btss.PartyID, error) {
	peerID, err := peer.Decode(item)
	if err != nil {
		return nil, fmt.Errorf("fail to decode peer ID string (%s): %w", item, err)
	}
	pubkey, err := peerID.ExtractPublicKey()
	if err != nil {
		return nil, fmt.Errorf("fail to extract public key from peer ID (%s): %w", item, err)
	}
	pubkeyRaw, err := pubkey.Raw()
	if err != nil {
		return nil, fmt.Errorf("fail to get raw public key from peer ID (%s): %w", item, err)
	}

	key := new(big.Int).SetBytes(pubkeyRaw)
	id := fmt.Sprintf("%d", idx+1)
	moniker := id
	if isNewCommittee {
		moniker = fmt.Sprintf("P[%d]", idx+1)
	}
	return btss.NewPartyID(id, moniker, key), nil
}

// GetPreviousKeySignUicast returns previous unicast key for tss message
func GetPreviousKeySignUicast(current string) string {
	if strings.HasSuffix(current, messages.KEYSIGN1b) {
		return messages.KEYSIGN1aUnicast
	}
	return messages.KEYSIGN2Unicast
}

func isOnCurve(x, y *big.Int) bool {

	curve := btcec.S256()
	return curve.IsOnCurve(x, y)
}

// GetTssPubKey returns public key and address from ECPoint
func GetTssPubKey(pubKeyPoint *crypto.ECPoint) (string, error) {
	pubKeyBytes := elliptic.Marshal(btcec.S256(), pubKeyPoint.X(), pubKeyPoint.Y())
	addrBytes := sha256.Sum256(pubKeyBytes)
	addr := hex.EncodeToString(addrBytes[:])
	return addr, nil
}

func GetTssPubKey_Eddsa(pubKeyPoint *crypto.ECPoint) (string, error) {
	// we check whether the point is on curve according to Kudelski report
	if pubKeyPoint == nil {
		return "", errors.New("invalid points")
	}

	// Initiate eddsa public key
	tssPubKey := edwards.PublicKey{
		Curve: edwards.Edwards(),
		X:     pubKeyPoint.X(),
		Y:     pubKeyPoint.Y(),
	}

	// check if it is on the curve
	if !tssPubKey.IsOnCurve(pubKeyPoint.X(), pubKeyPoint.Y()) {
		return "", errors.New("invalid points")
	}

	publicKeyBytes := tssPubKey.SerializeCompressed()

	// Derivate public key
	pubKey := hex.EncodeToString(publicKeyBytes[:])
	return pubKey, nil
}

func BytesToHashString(msg []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(msg)
	if err != nil {
		return "", fmt.Errorf("fail to caculate sha256 hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// GetThreshold calculates thresold for value
func GetThreshold(value int) (int, error) {
	if value < 0 {
		return 0, errors.New("negative input")
	}
	threshold := int(math.Ceil(float64(value)*2.0/3.0)) - 1
	if threshold < 1 {
		threshold = 1
	}
	return threshold, nil
}
