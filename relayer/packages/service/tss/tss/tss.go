package tss

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/Datura-ai/relayer/packages/service/tss/discovery"
	bkeygen "github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	dcommon "github.com/Datura-ai/relayer/packages/domain/common"
	repository "github.com/Datura-ai/relayer/packages/repository/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
	"github.com/Datura-ai/relayer/packages/service/tss/conversion"
	"github.com/Datura-ai/relayer/packages/service/tss/keygen"
	"github.com/Datura-ai/relayer/packages/service/tss/keysign"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	"github.com/Datura-ai/relayer/packages/service/tss/resharing"
	"github.com/Datura-ai/relayer/packages/service/tss/storage"
	tcrypto "github.com/libp2p/go-libp2p/core/crypto"
)

// TssServer is the structure that can provide all keysign and key gen features
type TssServer struct {
	conf              common.TssConfig
	logger            zerolog.Logger
	p2pCommunication  *p2p.Communication
	localNodePubKey   string
	preParams         *bkeygen.LocalPreParams
	tssKeygenLocker   *sync.Mutex
	stopChan          chan struct{}
	partyCoordinator  *p2p.PartyCoordinator
	discovery         *discovery.PeerDiscovery
	stateManager      storage.LocalStateManager
	tssStateManager   repository.TssStateManager
	signatureNotifier *keysign.SignatureNotifier
	privateKey        tcrypto.PrivKey
	handlerID         uint64
	keygenChan        *chan []peer.ID
	keysignChan       *chan dcommon.BridgeTxData
}

// NewTss create a new instance of Tss
func NewTss(
	conf common.TssConfig,
	preParams *bkeygen.LocalPreParams,
	comm *p2p.Communication,
	tssStateManager repository.TssStateManager,
	stateManager *storage.FileStateMgr,
	partyCoordinator *p2p.PartyCoordinator,
	discovery *discovery.PeerDiscovery,
	pubKey string,
	priKey tcrypto.PrivKey,
	handlerID uint64,
	keygenChan *chan []peer.ID,
	keysignChan *chan dcommon.BridgeTxData,
) (*TssServer, *bkeygen.LocalPreParams, error) {
	var err error
	// When using the keygen party it is recommended that you pre-compute the
	// "safe primes" and Paillier secret beforehand because this can take some
	// time.
	// This code will generate those parameters using a concurrency limit equal
	// to the number of available CPU cores.
	if preParams == nil || !preParams.Validate() {
		preParams, err = common.GeneratePreParams(conf.PreParamTimeout)
		if err != nil {
			return nil, nil, fmt.Errorf("fail to generate pre parameters: %w", err)
		}
	}
	if !preParams.Validate() {
		return nil, nil, errors.New("invalid preparams")
	}
	discovery.SetNewKeygenChannel(keygenChan)
	partyCoordinator.SetNotifiers(keygenChan, keysignChan)
	return &TssServer{
		conf:              conf,
		logger:            log.With().Str("module", "tss").Logger(),
		p2pCommunication:  comm,
		localNodePubKey:   pubKey,
		preParams:         preParams,
		tssKeygenLocker:   &sync.Mutex{},
		stopChan:          make(chan struct{}),
		partyCoordinator:  partyCoordinator,
		tssStateManager:   tssStateManager,
		stateManager:      stateManager,
		discovery:         discovery,
		signatureNotifier: keysign.NewSignatureNotifier(comm.GetHost(), handlerID),
		privateKey:        priKey,
		handlerID:         handlerID,
		keygenChan:        keygenChan,
		keysignChan:       keysignChan,
	}, preParams, nil
}

// Start Tss server
func (t *TssServer) Start() error {
	log.Info().Msg("Starting the TSS servers")
	return nil
}

// Stop Tss server
func (t *TssServer) Stop() {
	// stop the p2p and finish the p2p wait group
	err := t.p2pCommunication.Stop()
	if err != nil {
		t.logger.Error().Msgf("error in shutdown the p2p server")
	}
	t.partyCoordinator.Stop()
	log.Info().Msg("The Tss and p2p server has been stopped successfully")
}

func (t *TssServer) requestToMsgId(request interface{}) (string, error) {
	var dat []byte
	var keys []string
	switch value := request.(type) {
	case keygen.Request:
		keys = append(keys, value.Keys...)
	case resharing.Request:
		keys = append(keys, value.Keys...)
		keys = append(keys, value.NewKeys...)
	case keysign.Request:
		sort.Strings(value.Messages)
		dat = []byte(strings.Join(value.Messages, ","))
		keys = value.SignerPubKeys
	case []string:
		sort.Strings(value)
		keys = append(keys, value...)
	default:
		t.logger.Error().Msg("unknown request type")
		return "", errors.New("unknown request type")
	}
	keyAccumulation := ""
	sort.Strings(keys)
	for _, el := range keys {
		keyAccumulation += el
	}
	dat = append(dat, []byte(keyAccumulation)...)
	return common.MsgToHashString(dat)
}

func (t *TssServer) joinParty(msgID, requestType, version string, blockHeight int64, participants []string, threshold int, sigChan chan string, msgToSign []byte) ([]peer.ID, string, []byte, error) {
	var peersIDStr []string
	oldJoinParty, err := conversion.VersionLTCheck(version, messages.NEWJOINPARTYVERSION)
	if err != nil {
		return nil, "", nil, fmt.Errorf("fail to parse the version with error:%w", err)
	}
	if oldJoinParty {
		peerIDs, err := conversion.GetPeerIDsFromPubKeys(participants)
		if err != nil {
			return nil, "NONE", nil, fmt.Errorf("fail to convert pub key to peer id: %w", err)
		}
		for _, el := range peerIDs {
			peersIDStr = append(peersIDStr, el.String())
		}
		t.logger.Info().Msgf("we apply the leadless join party, peers: %+v", peersIDStr)
		onlines, err := t.partyCoordinator.JoinPartyWithRetry(msgID, peersIDStr)
		return onlines, "NONE", nil, err
	} else {
		t.logger.Info().Msg("we apply the join party with a leader")

		if len(participants) == 0 {
			t.logger.Error().Msg("we fail to have any participants or passed by request")
			return nil, "", nil, errors.New("no participants can be found")
		}
		peersID, err := conversion.GetPeerIDsFromPubKeys(participants)
		if err != nil {
			return nil, "", nil, fmt.Errorf("fail to convert the public key to peer ID: %s", err)
		}
		for _, el := range peersID {
			peersIDStr = append(peersIDStr, el.String())
		}

		t.logger.Info().Msgf("string here peer ids: (%s)", peersIDStr)

		// Assuming the missing return value is a byte slice, you can return nil or an appropriate value.
		peers, str, err := t.partyCoordinator.JoinPartyWithLeader(msgID, requestType, blockHeight, peersIDStr, threshold, msgToSign, sigChan)
		return peers, str, nil, err
	}
}

// GetLocalPeerID return the local peer
func (t *TssServer) GetLocalPeerID() string {
	return t.p2pCommunication.GetLocalPeerID()
}

// Stop party co ordinator
func (t *TssServer) StopPartyConnector() {
	//
	// finish the p2p wait group
	t.partyCoordinator.Stop()
	t.signatureNotifier.RemoveStreamHandler()
	log.Info().Msg("The Tss and p2p server has been stopped successfully")
}

func (t *TssServer) GetNewPeerChannel() *chan []peer.ID {
	return t.keygenChan
}

func (t *TssServer) GetKeysignChan() *chan dcommon.BridgeTxData {
	return t.keysignChan
}

func (t *TssServer) GetStateManager() repository.TssStateManager {
	return t.tssStateManager
}

func (t *TssServer) IsLeader(msgID string, blockHeight int64) (bool, error) {
	peersStr := []string{}
	host := t.p2pCommunication.GetHost()
	peers := host.Network().Peers()
	for _, peer := range peers {
		peersStr = append(peersStr, peer.String())
	}
	return t.partyCoordinator.IsLeader(msgID, blockHeight, peersStr)
}
