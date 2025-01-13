package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	btss "github.com/bnb-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/Datura-ai/relayer/packages/service/tss/blame"
	"github.com/Datura-ai/relayer/packages/service/tss/conversion"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// PartyInfo the information used by tss key gen and key sign
type PartyInfo struct {
	PartyMap      *sync.Map
	NewPartyMap   *sync.Map
	PartyIDMap    map[string]*btss.PartyID
	NewPartyIDMap map[string]*btss.PartyID
	PartyID       *btss.PartyID
	NewPartyID    *btss.PartyID
}

type TssCommon struct {
	conf                        TssConfig
	logger                      zerolog.Logger
	partyLock                   *sync.Mutex
	partyInfo                   *PartyInfo
	partyIDtoP2PID              map[string]peer.ID
	newPartyIDtoP2PID           map[string]peer.ID
	unConfirmedMsgLock          *sync.Mutex
	unConfirmedMessages         map[string]*LocalCacheItem
	localNodeID                 string
	localPeerID                 *peer.ID
	broadcastChannel            chan *messages.BroadcastMsgChan
	TssMsg                      chan *p2p.Message
	P2PPeersLock                *sync.RWMutex
	P2PPeers                    []peer.ID // most of tss message are broadcast, we store the peers ID to avoid iterating
	privateKey                  crypto.PrivKey
	taskDone                    chan struct{}
	blameMgr                    *blame.Manager
	finishedPeers               map[string]bool
	culprits                    []*btss.PartyID
	culpritsLock                *sync.RWMutex
	cachedWireBroadcastMsgLists *sync.Map
	cachedWireUnicastMsgLists   *sync.Map
	msgNum                      int
	numOldCommitte              int
	handleID                    uint64
}

func NewTssCommon(peerID string, broadcastChannel chan *messages.BroadcastMsgChan, conf TssConfig, privKey crypto.PrivKey, msgNum int, handleID uint64) (*TssCommon, error) {
	pID := peer.ID(peerID)
	//localPeerID, err := conversion.GetPeerIDFromPubKey(peerID)
	//if err != nil {
	//	return nil, errors.New("error while converting local node id to peer id")
	//}
	return &TssCommon{
		conf:                        conf,
		logger:                      log.With().Str("module", "tsscommon").Logger(),
		partyLock:                   &sync.Mutex{},
		partyInfo:                   nil,
		partyIDtoP2PID:              make(map[string]peer.ID),
		newPartyIDtoP2PID:           make(map[string]peer.ID),
		unConfirmedMsgLock:          &sync.Mutex{},
		unConfirmedMessages:         make(map[string]*LocalCacheItem),
		broadcastChannel:            broadcastChannel,
		TssMsg:                      make(chan *p2p.Message, msgNum),
		P2PPeersLock:                &sync.RWMutex{},
		P2PPeers:                    nil,
		localNodeID:                 peerID,
		localPeerID:                 &pID,
		privateKey:                  privKey,
		taskDone:                    make(chan struct{}),
		blameMgr:                    blame.NewBlameManager(),
		finishedPeers:               make(map[string]bool),
		culpritsLock:                &sync.RWMutex{},
		cachedWireBroadcastMsgLists: &sync.Map{},
		cachedWireUnicastMsgLists:   &sync.Map{},
		msgNum:                      msgNum,
		numOldCommitte:              0,
		handleID:                    handleID,
	}, nil
}

type BulkWireMsg struct {
	WiredBulkMsgs []byte
	MsgIdentifier string
	Routing       *btss.MessageRouting
}

func NewBulkWireMsg(msg []byte, id string, r *btss.MessageRouting) BulkWireMsg {
	return BulkWireMsg{
		WiredBulkMsgs: msg,
		MsgIdentifier: id,
		Routing:       r,
	}
}

type tssJob struct {
	wireBytes     []byte
	msgIdentifier string
	partyID       *btss.PartyID
	isBroadcast   bool
	localParty    btss.Party
	// acceptedShares map[blame.RoundInfo][]string
}

func newJob(party btss.Party, wireBytes []byte, msgIdentifier string, from *btss.PartyID, isBroadcast bool) *tssJob {
	return &tssJob{
		wireBytes:     wireBytes,
		msgIdentifier: msgIdentifier,
		partyID:       from,
		isBroadcast:   isBroadcast,
		localParty:    party,
	}
}

func (t *TssCommon) doTssJob(tssJobChan chan *tssJob, jobWg *sync.WaitGroup) {
	defer func() {
		jobWg.Done()
	}()

	for tssjob := range tssJobChan {
		party := tssjob.localParty
		wireBytes := tssjob.wireBytes
		partyID := tssjob.partyID
		isBroadcast := tssjob.isBroadcast

		round, err := GetMsgRound(wireBytes, partyID, isBroadcast, t.handleID)
		if err != nil {
			t.logger.Error().Err(err).Msg("broken tss share")
			continue
		}
		round.MsgIdentifier = tssjob.msgIdentifier

		if party != nil {
			if _, errUp := party.UpdateFromBytes(wireBytes, partyID, isBroadcast); errUp != nil {
				err := t.processInvalidMsgBlame(round.RoundMsg, round, errUp)
				t.logger.Error().Err(err).Msgf("fail to apply the share to tss")
				continue
			}
		}

		// we need to retrieve the partylist again as others may update it once we process apply tss share
		t.blameMgr.UpdateAcceptShare(round, partyID.Id)
	}
}

func (t *TssCommon) renderToP2P(broadcastMsg *messages.BroadcastMsgChan) {
	if t.broadcastChannel == nil {
		t.logger.Warn().Msg("broadcast channel is not set")
		return
	}
	t.broadcastChannel <- broadcastMsg
}

// GetConf get current configuration for Tss
func (t *TssCommon) GetConf() TssConfig {
	return t.conf
}

func (t *TssCommon) GetTaskDone() chan struct{} {
	return t.taskDone
}

func (t *TssCommon) WaitForNewTaskDone() {
	t.taskDone = make(chan struct{})
	t.finishedPeers = make(map[string]bool)
}

func (t *TssCommon) GetBlameMgr() *blame.Manager {
	return t.blameMgr
}

func (t *TssCommon) SetPartyInfo(partyInfo *PartyInfo) {
	t.partyLock.Lock()
	defer t.partyLock.Unlock()
	t.partyInfo = partyInfo
}

func (t *TssCommon) getPartyInfo() *PartyInfo {
	t.partyLock.Lock()
	defer t.partyLock.Unlock()
	return t.partyInfo
}

func (t *TssCommon) GetLocalNodeID() string {
	return t.localNodeID
}

func (t *TssCommon) SetLocalPeerID(peerID, newPeerID *peer.ID) {
	t.localPeerID = peerID
}

func (t *TssCommon) processInvalidMsgBlame(roundInfo string, round blame.RoundInfo, err *btss.Error) error {
	// now we get the culprits ID, invalid message and signature the culprits sent
	var culpritsID []string
	var invalidMsgs []*messages.WireMessage
	unicast := checkUnicast(round)
	t.culpritsLock.Lock()
	t.culprits = append(t.culprits, err.Culprits()...)
	t.culpritsLock.Unlock()
	for _, el := range err.Culprits() {
		culpritsID = append(culpritsID, el.Id)
		key := fmt.Sprintf("%s-%s", el.Id, roundInfo)
		storedMsg := t.blameMgr.GetRoundMgr().Get(key)
		invalidMsgs = append(invalidMsgs, storedMsg)
	}
	pubkeys, errBlame := conversion.AccPubKeysFromPartyIDs(culpritsID, t.partyInfo.PartyIDMap)
	if errBlame != nil {
		t.logger.Error().Err(err.Cause()).Msgf("error in get the blame nodes")
		t.blameMgr.GetBlame().SetBlame(blame.TssBrokenMsg, nil, unicast)
		return fmt.Errorf("error in getting the blame nodes")
	}
	// This error indicates the share is wrong, we include this signature to prove that
	// this incorrect share is from the share owner.
	var blameNodes []blame.Node
	var msgBody, sig []byte
	for i, pk := range pubkeys {
		invalidMsg := invalidMsgs[i]
		if invalidMsg == nil {
			t.logger.Error().Msg("we cannot find the record of this curlprit, set it as blank")
			msgBody = []byte{}
			sig = []byte{}
		} else {
			msgBody = invalidMsg.Message
			sig = invalidMsg.Sig
		}
		blameNodes = append(blameNodes, blame.NewNode(pk, msgBody, sig))
	}
	t.blameMgr.GetBlame().SetBlame(blame.TssBrokenMsg, blameNodes, unicast)
	return fmt.Errorf("fail to set bytes to local party: %w", err)
}

// updateLocal will apply the wireMsg to local party
func (t *TssCommon) UpdateLocal(wireMsg *messages.WireMessage, msgType messages.DaturaTSSMessageType) error {
	if wireMsg == nil || wireMsg.Routing == nil || wireMsg.Routing.From == nil {
		t.logger.Warn().Msg("wire msg is nil")
		return errors.New("invalid wireMsg")
	}
	partyInfo := t.getPartyInfo()
	if partyInfo == nil {
		return errors.New("partyInfo is null")
	}

	_, ok := partyInfo.PartyIDMap[wireMsg.Routing.From.Id]
	if !ok {
		return fmt.Errorf("get message from unknown party %s", wireMsg.Routing.From.Id)
	}

	var bulkMsg []BulkWireMsg
	if err := json.Unmarshal(wireMsg.Message, &bulkMsg); err != nil {
		t.logger.Error().Err(err).Msg("error to unmarshal the BulkMsg")
		return err
	}

	worker := runtime.NumCPU()
	tssJobChan := make(chan *tssJob, len(bulkMsg))
	jobWg := sync.WaitGroup{}
	for i := 0; i < worker; i++ {
		jobWg.Add(1)
		go t.doTssJob(tssJobChan, &jobWg)
	}
	for i, msg := range bulkMsg {
		_, err := conversion.GetPeerIDFromPartyID(msg.Routing.From)
		if err != nil {
			t.logger.Error().Err(err).Msg("error while converting partyID to peerID")
			return fmt.Errorf("error while converting partyID to peerID: %w", err)
		}

		var data any
		data, ok = partyInfo.PartyMap.Load("")
		if msgType == messages.TSSKeysignMsg {
			moniker := strconv.Itoa(i)
			data, ok = partyInfo.PartyMap.Load(moniker)
		}
		if !ok {
			t.logger.Error().Msgf("cannot find the party to this wired msg: %d", msgType)
			return errors.New("cannot find the party")
		}

		if msg.Routing.From.Id != wireMsg.Routing.From.Id {
			// this should never happen , if it happened , which ever party did it , should be blamed and slashed
			t.logger.Error().Msgf("all messages in a batch sign should have the same routing ,batch routing party id: %s, however message routing:%s", msg.Routing.From, wireMsg.Routing.From)
		}

		localMsgParty := data.(btss.Party)
		partyID, ok := partyInfo.PartyIDMap[wireMsg.Routing.From.Id]
		if !ok {
			t.logger.Error().Msg("error in find the partyID")
			return errors.New("cannot find the party to handle the message")
		}

		round, err := GetMsgRound(msg.WiredBulkMsgs, partyID, msg.Routing.IsBroadcast, t.handleID)
		if err != nil {
			t.logger.Error().Err(err).Msg("broken tss share")
			return err
		}

		// we only allow a message be updated only once.
		// here we use round + msgIdentifier as the key for the acceptedShares
		round.MsgIdentifier = msg.MsgIdentifier

		// if this share is duplicated, we skip this share
		if t.blameMgr.CheckMsgDuplication(round, partyID.Id) {
			t.logger.Debug().Msgf("we received the duplicated message from party %s", partyID.Id)
			continue
		}

		partyInlist := func(el *btss.PartyID, l []*btss.PartyID) bool {
			for _, each := range l {
				if el == each {
					return true
				}
			}
			return false
		}

		t.culpritsLock.RLock()
		if len(t.culprits) != 0 && partyInlist(partyID, t.culprits) {
			t.logger.Error().Msgf("the malicious party (party ID:%s) try to send incorrect message to me (tss local party ID:%s)", partyID.Id, localMsgParty.PartyID().Id)
			t.culpritsLock.RUnlock()
			return errors.New(blame.TssBrokenMsg)
		}
		t.culpritsLock.RUnlock()
		job := newJob(localMsgParty, msg.WiredBulkMsgs, round.MsgIdentifier, partyID, msg.Routing.IsBroadcast)
		tssJobChan <- job
	}
	close(tssJobChan)
	jobWg.Wait()
	return nil
}

// updateLocal will apply the wireMsg to local party
func (t *TssCommon) UpdateLocalResharing(wireMsg *messages.WireMessage, msgType messages.DaturaTSSMessageType) error {
	if wireMsg == nil || wireMsg.Routing == nil || wireMsg.Routing.From == nil {
		t.logger.Warn().Msg("wire msg is nil")
		return errors.New("invalid wireMsg")
	}
	partyInfo := t.getPartyInfo()
	if partyInfo == nil {
		return errors.New("partyInfo is null")
	}

	var ok bool
	var bulkMsg []BulkWireMsg
	if err := json.Unmarshal(wireMsg.Message, &bulkMsg); err != nil {
		t.logger.Error().Err(err).Msg("error to unmarshal the BulkMsg")
		return err
	}

	worker := runtime.NumCPU()
	tssJobChan := make(chan *tssJob, len(bulkMsg))
	jobWg := sync.WaitGroup{}
	for i := 0; i < worker; i++ {
		jobWg.Add(1)
		go t.doTssJob(tssJobChan, &jobWg)
	}
	for _, msg := range bulkMsg {
		_, err := conversion.GetPeerIDFromPartyID(msg.Routing.From)
		if err != nil {
			t.logger.Error().Err(err).Msg("error while converting partyID to peerID")
			return fmt.Errorf("error while converting partyID to peerID: %w", err)
		}

		var data any
		if msgType == messages.TSSKeyresharingMsgNew {
			data, ok = partyInfo.NewPartyMap.Load("")
		} else {
			data, ok = partyInfo.PartyMap.Load("")
		}
		if !ok {
			t.logger.Error().Msgf("cannot find the party to this wired msg: %d", msgType)
			return errors.New("cannot find the party")
		}

		if msg.Routing.From.Id != wireMsg.Routing.From.Id {
			// this should never happen , if it happened , which ever party did it , should be blamed and slashed
			t.logger.Error().Msgf("all messages in a batch sign should have the same routing ,batch routing party id: %s, however message routing:%s", msg.Routing.From, wireMsg.Routing.From)
			return errors.New("all messages in a batch sign should have the same routing")
		}

		partyID := partyInfo.PartyIDMap[msg.Routing.From.Id]
		if strings.Contains(wireMsg.Routing.From.Moniker, conversion.NEW_PARTY_MONIKER_SUFFIX) {
			partyID = partyInfo.NewPartyIDMap[msg.Routing.From.Id]
		}

		if !ok {
			t.logger.Error().Msg("error in find the partyID")
			return errors.New("cannot find the party to handle the message")
		}

		round, err := GetMsgRound(msg.WiredBulkMsgs, partyID, msg.Routing.IsBroadcast, t.handleID)
		if err != nil {
			t.logger.Error().Err(err).Msg("broken tss share")
			return err
		}

		// we only allow a message be updated only once.
		// here we use round + msgIdentifier as the key for the acceptedShares
		round.MsgIdentifier = msg.MsgIdentifier
		partyInlist := func(el *btss.PartyID, l []*btss.PartyID) bool {
			for _, each := range l {
				if el == each {
					return true
				}
			}
			return false
		}

		var localMsgParty btss.Party
		if reflect.TypeOf(data).Implements(reflect.TypeOf((*btss.Party)(nil)).Elem()) {
			localMsgParty = data.(btss.Party)
		}

		t.culpritsLock.RLock()
		if len(t.culprits) != 0 && partyInlist(partyID, t.culprits) {
			localMsgPartyID := ""
			if localMsgParty != nil {
				localMsgPartyID = localMsgParty.PartyID().Id
			}
			t.logger.Error().Msgf("the malicious party (party ID:%s) try to send incorrect message to me (tss local party ID:%s)", partyID.Id, localMsgPartyID)
			t.culpritsLock.RUnlock()
			return errors.New(blame.TssBrokenMsg)
		}
		t.culpritsLock.RUnlock()
		job := newJob(localMsgParty, msg.WiredBulkMsgs, round.MsgIdentifier, partyID, msg.Routing.IsBroadcast)
		tssJobChan <- job
	}
	close(tssJobChan)
	jobWg.Wait()
	return nil
}

func (t *TssCommon) checkDupAndUpdateVerMsg(bMsg *messages.BroadcastConfirmMessage, peerID string) bool {
	localCacheItem := t.TryGetLocalCacheItem(bMsg.Key)
	// we check whether this node has already sent the VerMsg message to avoid eclipse of others VerMsg
	if localCacheItem == nil {
		bMsg.P2PID = peerID
		return true
	}

	localCacheItem.lock.Lock()
	defer localCacheItem.lock.Unlock()
	if _, ok := localCacheItem.ConfirmedList[peerID]; ok {
		return false
	}
	bMsg.P2PID = peerID
	return true
}

func (t *TssCommon) ProcessOneMessage(wrappedMsg *messages.WrappedMessage, peerID, msgID string) error {
	if nil == wrappedMsg {
		return errors.New("invalid wireMessage")
	}

	t.logger.Info().Msgf("start process one message %s", wrappedMsg.MessageType.String())
	defer t.logger.Info().Msgf("finish processing one message %s", wrappedMsg.MessageType.String())

	switch wrappedMsg.MessageType {
	case messages.TSSKeygenMsg, messages.TSSKeysignMsg:
		var wireMsg messages.WireMessage
		if err := json.Unmarshal(wrappedMsg.Payload, &wireMsg); nil != err {
			return fmt.Errorf("fail to unmarshal wire message: %w", err)
		}

		return t.processTSSMsg(&wireMsg, wrappedMsg.MessageType, true, msgID)
	case messages.TSSKeyresharingMsgOld, messages.TSSKeyresharingMsgNew:
		var wireMsg messages.WireMessage
		if err := json.Unmarshal(wrappedMsg.Payload, &wireMsg); nil != err {
			return fmt.Errorf("fail to unmarshal wire message: %w", err)
		}

		return t.processTSSMsgResharing(&wireMsg, wrappedMsg.MessageType, true, msgID)

	case messages.TSSTaskDone:
		var wireMsg messages.TssTaskNotifier
		err := json.Unmarshal(wrappedMsg.Payload, &wireMsg)
		if err != nil {
			t.logger.Error().Err(err).Msg("fail to unmarshal the notify message")
			return nil
		}
		if wireMsg.TaskDone {
			// if we have already logged this node, we return to avoid close of a close channel
			if t.finishedPeers[peerID] {
				return fmt.Errorf("duplicated notification from peer %s ignored", peerID)
			}
			t.finishedPeers[peerID] = true
			if len(t.finishedPeers) == len(t.partyInfo.PartyIDMap)-1 {
				t.logger.Debug().Msg("we get the confirm of the nodes that generate the signature")
				close(t.taskDone)
			}
			return nil
		}
	}

	return nil
}

func (t *TssCommon) getMsgHash(localCacheItem *LocalCacheItem, threshold int) (string, error) {
	hash, freq, err := getHighestFreq(localCacheItem.ConfirmedList)
	if err != nil {
		t.logger.Error().Err(err).Msg("fail to get the hash freq")
		return "", blame.ErrHashCheck
	}
	if freq < threshold-1 {
		t.logger.Debug().Msgf("fail to have more than 2/3 peers agree on the received message threshold(%d)--total confirmed(%d)\n", threshold, freq)
		return "", blame.ErrHashInconsistency
	}
	return hash, nil
}

func (t *TssCommon) hashCheck(localCacheItem *LocalCacheItem, threshold int, msgType messages.DaturaTSSMessageType) error {
	dataOwner := localCacheItem.Msg.Routing.From
	dataOwnerP2PID, ok := t.partyIDtoP2PID[dataOwner.Id]
	if strings.Contains(dataOwner.Moniker, conversion.NEW_PARTY_MONIKER_SUFFIX) && (msgType != messages.TSSKeysignMsg && msgType != messages.TSSKeygenMsg) {
		dataOwnerP2PID, ok = t.newPartyIDtoP2PID[dataOwner.Id]
	}

	if !ok {
		t.logger.Warn().Msgf("error in find the data Owner P2PID\n")
		return errors.New("error in find the data Owner P2PID")
	}

	// TODO: Need check the possibility of not using local cache.
	// if localCacheItem.TotalConfirmParty() < threshold {
	// 	t.logger.Debug().Msgf("not enough nodes to evaluate the hash, confirmed: %d threshold: %d", localCacheItem.TotalConfirmParty(), threshold)
	// 	return blame.ErrNotEnoughPeer
	// }
	localCacheItem.lock.Lock()
	defer localCacheItem.lock.Unlock()

	targetHashValue := localCacheItem.Hash
	for P2PID := range localCacheItem.ConfirmedList {
		if P2PID == dataOwnerP2PID.String() && msgType != messages.TSSKeyresharingMsgNew && msgType != messages.TSSKeyresharingMsgOld {
			t.logger.Warn().Msgf("we detect that the data owner try to send the hash for his own message\n")
			delete(localCacheItem.ConfirmedList, P2PID)
			return blame.ErrHashFromOwner
		}
	}
	hash, err := t.getMsgHash(localCacheItem, threshold)
	if err != nil {
		return err
	}
	if targetHashValue == hash {
		t.logger.Debug().Msgf("hash check complete")
		return nil
	}
	return blame.ErrNotMajority
}

func (t *TssCommon) sendBulkMsg(wiredMsgType string, tssMsgType messages.DaturaTSSMessageType, wiredMsgList []BulkWireMsg, msgID string) error {
	// since all the messages in the list is the same round, so it must have the same dest
	// we just need to get the routing info of the first message
	r := wiredMsgList[0].Routing

	buf, err := json.Marshal(wiredMsgList)
	if err != nil {
		return fmt.Errorf("error in marshal the cachedWireMsg: %w", err)
	}

	sig, err := generateSignature(buf, msgID, t.privateKey)
	if err != nil {
		t.logger.Error().Err(err).Msg("fail to generate the share's signature")
		return err
	}

	wireMsg := messages.WireMessage{
		Routing:   r,
		RoundInfo: wiredMsgType,
		Message:   buf,
		Sig:       sig,
	}
	wireMsgBytes, err := json.Marshal(wireMsg)
	if err != nil {
		return fmt.Errorf("fail to convert tss msg to wire bytes: %w", err)
	}
	wrappedMsg := messages.WrappedMessage{
		MsgID:       msgID,
		MessageType: tssMsgType,
		Payload:     wireMsgBytes,
		HandleID:    t.handleID,
	}

	peerIDs := make([]peer.ID, 0)
	if len(r.To) == 0 {
		t.P2PPeersLock.RLock()
		peerIDs = t.P2PPeers
		t.P2PPeersLock.RUnlock()
	} else {
		if tssMsgType == messages.TSSKeyresharingMsgOld || tssMsgType == messages.TSSKeyresharingMsgNew {
			for i, each := range r.To {
				if tssMsgType == messages.TSSKeyresharingMsgOld && i >= t.numOldCommitte {
					break
				}

				// Same Id - itself, old -> new or new -> old
				if r.From.Id == each.Id {
					// Update local itself as P2P won't send message to himself
					if r.From.Moniker != each.Moniker {
						t.UpdateLocalResharing(&wireMsg, tssMsgType)
					}
					continue
				}

				peerID, ok := t.partyIDtoP2PID[each.Id]
				if tssMsgType == messages.TSSKeyresharingMsgNew {
					peerID, ok = t.newPartyIDtoP2PID[each.Id]
				}
				if !ok {
					t.logger.Error().Msg("error in find the P2P ID")
					continue
				}
				peerIDs = append(peerIDs, peerID)
			}
		} else { // key gen and key sign
			each := r.To[0]
			if each.Id == r.From.Id {
				return nil
			}
			peerID, ok := t.partyIDtoP2PID[each.Id]
			if !ok {
				t.logger.Error().Msg("error in find the P2P ID")
				return nil
			}
			peerIDs = append(peerIDs, peerID)
		}
	}
	t.renderToP2P(&messages.BroadcastMsgChan{
		WrappedMessage: wrappedMsg,
		PeersID:        peerIDs,
	})

	return nil
}

func (t *TssCommon) ProcessOutCh(msg btss.Message, msgType messages.DaturaTSSMessageType, msgID string) error {
	msgData, r, err := msg.WireBytes()
	// if we cannot get the wire share, the tss will fail, we just quit.
	if err != nil {
		return fmt.Errorf("fail to get wire bytes: %w", err)
	}

	if r.IsBroadcast {
		cachedWiredMsg := NewBulkWireMsg(msgData, msg.GetFrom().Id, r)
		// now we store this message in cache
		dat, ok := t.cachedWireBroadcastMsgLists.Load(msg.Type())
		if !ok {
			l := []BulkWireMsg{cachedWiredMsg}
			t.cachedWireBroadcastMsgLists.Store(msg.Type(), l)
		} else {
			cachedList := dat.([]BulkWireMsg)
			cachedList = append(cachedList, cachedWiredMsg)
			t.cachedWireBroadcastMsgLists.Store(msg.Type(), cachedList)
		}
	} else {
		cachedWiredMsg := NewBulkWireMsg(msgData, msg.GetFrom().Id, r)
		dat, ok := t.cachedWireUnicastMsgLists.Load(msg.Type() + ":" + r.To[0].String())
		if !ok {
			l := []BulkWireMsg{cachedWiredMsg}
			t.cachedWireUnicastMsgLists.Store(msg.Type()+":"+r.To[0].String(), l)
		} else {
			cachedList := dat.([]BulkWireMsg)
			cachedList = append(cachedList, cachedWiredMsg)
			t.cachedWireUnicastMsgLists.Store(msg.Type()+":"+r.To[0].String(), cachedList)
		}
	}
	t.cachedWireUnicastMsgLists.Range(func(key, value interface{}) bool {
		wiredMsgList := value.([]BulkWireMsg)
		ret := strings.Split(key.(string), ":")
		wiredMsgType := ret[0]
		if len(wiredMsgList) == t.msgNum {
			err := t.sendBulkMsg(wiredMsgType, msgType, wiredMsgList, msgID)
			if err != nil {
				t.logger.Error().Err(err).Msg("error in send bulk message")
				return true
			}
			t.cachedWireUnicastMsgLists.Delete(key)
		}
		return true
	})

	t.cachedWireBroadcastMsgLists.Range(func(key, value interface{}) bool {
		wiredMsgList := value.([]BulkWireMsg)
		wiredMsgType := key.(string)
		if len(wiredMsgList) == t.msgNum {
			err := t.sendBulkMsg(wiredMsgType, msgType, wiredMsgList, msgID)
			if err != nil {
				t.logger.Error().Err(err).Msg("error in send bulk message")
				return true
			}
			t.cachedWireBroadcastMsgLists.Delete(key)
		}
		return true
	})

	return nil
}

func (t *TssCommon) applyShare(localCacheItem *LocalCacheItem, threshold int, key string, msgType messages.DaturaTSSMessageType, msgID string) error {
	unicast := true
	if localCacheItem.Msg.Routing.IsBroadcast {
		unicast = false
	}
	err := t.hashCheck(localCacheItem, threshold, msgType)
	if err != nil {
		if errors.Is(err, blame.ErrNotEnoughPeer) {
			return nil
		}
		if errors.Is(err, blame.ErrNotMajority) {
			t.logger.Error().Err(err).Msg("we send request to get the message match with majority")
			localCacheItem.Msg = nil
			return t.requestShareFromPeer(localCacheItem, threshold, key, msgType, msgID)
		}
		blamePk, err := t.blameMgr.TssWrongShareBlame(localCacheItem.Msg)
		if err != nil {
			t.logger.Error().Err(err).Msg("error in get the blame nodes")
			t.blameMgr.GetBlame().SetBlame(blame.HashCheckFail, nil, unicast)
			return fmt.Errorf("error in getting the blame nodes %w", blame.ErrHashCheck)
		}
		blameNode := blame.NewNode(blamePk, localCacheItem.Msg.Message, localCacheItem.Msg.Sig)
		t.blameMgr.GetBlame().SetBlame(blame.HashCheckFail, []blame.Node{blameNode}, unicast)
		return blame.ErrHashCheck
	}

	t.blameMgr.GetRoundMgr().Set(key, localCacheItem.Msg)
	if err := t.UpdateLocal(localCacheItem.Msg, msgType); nil != err {
		return fmt.Errorf("fail to update the message to local party: %w", err)
	}
	t.logger.Debug().Msgf("remove key: %s", key)
	// the information had been confirmed by all party , we don't need it anymore
	t.removeKey(key)
	return nil
}

func (t *TssCommon) requestShareFromPeer(localCacheItem *LocalCacheItem, threshold int, key string, msgType messages.DaturaTSSMessageType, msgID string) error {
	targetHash, err := t.getMsgHash(localCacheItem, threshold)
	if err != nil {
		t.logger.Debug().Msg("we do not know which message to request, so we quit")
		return nil
	}
	var peersIDs []peer.ID
	for thisPeerStr, hash := range localCacheItem.ConfirmedList {
		if hash == targetHash {
			thisPeer, err := peer.Decode(thisPeerStr)
			if err != nil {
				t.logger.Error().Err(err).Msg("fail to convert the p2p id")
				return err
			}
			peersIDs = append(peersIDs, thisPeer)
		}
	}

	msg := &messages.TssControl{
		ReqHash:     targetHash,
		ReqKey:      key,
		RequestType: 0,
		Msg:         nil,
	}

	t.blameMgr.GetShareMgr().Set(targetHash)
	return t.processMessage(msgType, msg, peersIDs, msgID)
}

func (t *TssCommon) processMessage(msgType messages.DaturaTSSMessageType, msg *messages.TssControl, peersIDs []peer.ID, msgID string) error {
	switch msgType {
	case messages.TSSKeygenVerMsg:
		msg.RequestType = messages.TSSKeygenMsg
	case messages.TSSKeysignVerMsg:
		msg.RequestType = messages.TSSKeysignMsg
	case messages.TSSKeyresharingVerMsg:
		msg.RequestType = messages.TSSKeyresharingMsgNew
	case messages.TSSEcdsaPartVerMsg:
		msg.RequestType = messages.TSSEcdsaPartMsg
	case messages.TSSKeysignMsg, messages.TSSKeygenMsg,
		messages.TSSKeyresharingMsgOld, messages.TSSKeyresharingMsgNew, messages.TSSEcdsaPartMsg:
		msg.RequestType = msgType
	default:
		t.logger.Debug().Msgf("unknown message type: %d", msgType)
		return nil
	}

	return t.processRequestMsgFromPeer(peersIDs, msg, true, msgID, t.handleID)
}

func (t *TssCommon) broadcastHashToPeers(peerIDs []peer.ID, msgType messages.DaturaTSSMessageType, key, msgID, msgHash string) error {
	if len(peerIDs) == 0 {
		t.logger.Error().Msg("fail to get any peer ID")
		return errors.New("fail to get any peer ID")
	}

	broadcastConfirmMsg := &messages.BroadcastConfirmMessage{
		// P2PID will be filled up by the receiver.
		P2PID: "",
		Key:   key,
		Hash:  msgHash,
	}
	buf, err := json.Marshal(broadcastConfirmMsg)
	if err != nil {
		return fmt.Errorf("fail to marshal borad cast confirm message: %w", err)
	}
	t.logger.Debug().Msgf("broadcast VerMsg to all other parties %+v", peerIDs)

	p2pWrappedMSg := messages.WrappedMessage{
		MessageType: msgType,
		MsgID:       msgID,
		Payload:     buf,
		HandleID:    t.handleID,
	}
	t.renderToP2P(&messages.BroadcastMsgChan{
		WrappedMessage: p2pWrappedMSg,
		PeersID:        peerIDs,
	})

	return nil
}

func (t *TssCommon) receiverBroadcastHashToPeers(wireMsg *messages.WireMessage, msgType messages.DaturaTSSMessageType, msgID string) error {
	var peerIDs []peer.ID
	dataOwnerPartyID := wireMsg.Routing.From.Id
	dataOwnerPeerID, ok := t.partyIDtoP2PID[dataOwnerPartyID]
	if strings.Contains(wireMsg.Routing.From.Moniker, conversion.NEW_PARTY_MONIKER_SUFFIX) {
		dataOwnerPeerID, ok = t.newPartyIDtoP2PID[dataOwnerPartyID]
	}
	if !ok {
		return errors.New("error in find the data owner peerID")
	}
	t.P2PPeersLock.RLock()
	for _, el := range t.P2PPeers {
		if el == dataOwnerPeerID {
			continue
		}
		peerIDs = append(peerIDs, el)
	}
	t.P2PPeersLock.RUnlock()
	msgVerType := getBroadcastMessageType(msgType)
	key := wireMsg.GetCacheKey()
	msgHash, err := conversion.BytesToHashString(wireMsg.Message)
	if err != nil {
		return fmt.Errorf("fail to calculate hash of the wire message: %w", err)
	}
	err = t.broadcastHashToPeers(peerIDs, msgVerType, key, msgID, msgHash)
	if err != nil {
		t.logger.Error().Err(err).Msg("fail to broadcast the hash to peers")
		return err
	}
	return nil
}

// processTSSResharingMsg
func (t *TssCommon) processTSSMsg(wireMsg *messages.WireMessage, msgType messages.DaturaTSSMessageType, forward bool, msgID string) error {
	t.logger.Debug().Msg("process wire message")
	defer t.logger.Debug().Msg("finish process wire message")

	if wireMsg == nil || wireMsg.Routing == nil || wireMsg.Routing.From == nil {
		t.logger.Warn().Msg("received msg invalid")
		return errors.New("invalid wireMsg")
	}

	partyIDMap := t.getPartyInfo().PartyIDMap
	dataOwner, ok := partyIDMap[wireMsg.Routing.From.Id]
	if !ok {
		t.logger.Error().Msg("error in find the data owner")
		return errors.New("error in find the data owner")
	}

	keyBytes := dataOwner.GetKey()
	pk, err := crypto.UnmarshalSecp256k1PublicKey(keyBytes)
	if err != nil {
		t.logger.Error().Msg("fail to unmarshal the public key")
		return fmt.Errorf("fail to unmarshal the public key: %w", err)
	}
	if ok = verifySignature(pk, wireMsg.Message, wireMsg.Sig, msgID); !ok {
		t.logger.Error().Msg("fail to verify the signature")
		return fmt.Errorf("signature verify failed for %s", msgType)
	}

	t.logger.Debug().Msgf("received msg from %s to %+v", wireMsg.Routing.From, wireMsg.Routing.To)

	// for the unicast message, we only update it local party
	if !wireMsg.Routing.IsBroadcast {
		t.logger.Debug().Msgf("msg from %s to %+v", wireMsg.Routing.From, wireMsg.Routing.To)
		return t.UpdateLocal(wireMsg, msgType)
	}

	// if not received the broadcast message , we save a copy locally , and then tell all others what we got
	if !forward {
		err := t.receiverBroadcastHashToPeers(wireMsg, msgType, msgID)
		if err != nil {
			t.logger.Error().Err(err).Msg("fail to broadcast msg to peers")
		}
	}

	partyInfo := t.getPartyInfo()
	key := wireMsg.GetCacheKey()
	msgHash, err := conversion.BytesToHashString(wireMsg.Message)
	if err != nil {
		return fmt.Errorf("fail to calculate hash of the wire message: %w", err)
	}
	localCacheItem := t.TryGetLocalCacheItem(key)
	if nil == localCacheItem {
		t.logger.Debug().Msgf("++%s doesn't exist yet,add a new one", key)
		localCacheItem = NewLocalCacheItem(wireMsg, msgHash)
		t.updateLocalUnconfirmedMessages(key, localCacheItem)
	} else {
		// this means we received the broadcast confirm message from other party first
		t.logger.Debug().Msgf("==%s exist", key)
		if localCacheItem.Msg == nil {
			t.logger.Debug().Msgf("==%s exist, set message", key)
			localCacheItem.Msg = wireMsg
			localCacheItem.Hash = msgHash
		}
	}

	localCacheItem.UpdateConfirmList(t.localPeerID.String(), msgHash)

	threshold, err := conversion.GetThreshold(len(partyInfo.PartyIDMap))
	if err != nil {
		return err
	}
	return t.applyShare(localCacheItem, threshold, key, msgType, msgID)
}

// processTSSResharingMsg
func (t *TssCommon) processTSSMsgResharing(wireMsg *messages.WireMessage, msgType messages.DaturaTSSMessageType, forward bool, msgID string) error {
	t.logger.Debug().Msg("process wire message")
	defer t.logger.Debug().Msg("finish process wire message")

	if wireMsg == nil || wireMsg.Routing == nil || wireMsg.Routing.From == nil {
		t.logger.Warn().Msg("received msg invalid")
		return errors.New("invalid wireMsg")
	}

	partyIDMap := t.getPartyInfo().PartyIDMap
	if strings.Contains(wireMsg.Routing.From.Moniker, conversion.NEW_PARTY_MONIKER_SUFFIX) {
		partyIDMap = t.getPartyInfo().NewPartyIDMap
	}

	dataOwner, ok := partyIDMap[wireMsg.Routing.From.Id]
	if !ok {
		t.logger.Error().Msg("error in find the data owner")
		return errors.New("error in find the data owner")
	}

	keyBytes := dataOwner.GetKey()
	pk, err := crypto.UnmarshalSecp256k1PublicKey(keyBytes)
	if err != nil {
		t.logger.Error().Msg("fail to unmarshal the public key")
		return fmt.Errorf("fail to unmarshal the public key: %w", err)
	}
	if ok = verifySignature(pk, wireMsg.Message, wireMsg.Sig, msgID); !ok {
		t.logger.Error().Msg("fail to verify the signature")
		return fmt.Errorf("signature verify failed for %s", msgType.String())
	}

	t.logger.Debug().Msgf("received msg from %s to %+v", wireMsg.Routing.From, wireMsg.Routing.To)

	return t.UpdateLocalResharing(wireMsg, msgType)
}

func getBroadcastMessageType(msgType messages.DaturaTSSMessageType) messages.DaturaTSSMessageType {
	switch msgType {
	case messages.TSSKeygenMsg:
		return messages.TSSKeygenVerMsg
	case messages.TSSKeysignMsg:
		return messages.TSSKeysignVerMsg
	case messages.TSSKeyresharingMsgOld:
		return messages.TSSKeyresharingVerMsg
	case messages.TSSKeyresharingMsgNew:
		return messages.TSSKeyresharingVerMsg
	case messages.TSSEcdsaPartMsg:
		return messages.TSSEcdsaPartVerMsg
	default:
		return messages.Unknown // this should not happen
	}
}

func (t *TssCommon) TryGetLocalCacheItem(key string) *LocalCacheItem {
	t.unConfirmedMsgLock.Lock()
	defer t.unConfirmedMsgLock.Unlock()
	localCacheItem, ok := t.unConfirmedMessages[key]
	if !ok {
		return nil
	}
	return localCacheItem
}

func (t *TssCommon) TryGetAllLocalCached() []*LocalCacheItem {
	var localCachedItems []*LocalCacheItem
	t.unConfirmedMsgLock.Lock()
	defer t.unConfirmedMsgLock.Unlock()
	for _, value := range t.unConfirmedMessages {
		localCachedItems = append(localCachedItems, value)
	}
	return localCachedItems
}

func (t *TssCommon) updateLocalUnconfirmedMessages(key string, cacheItem *LocalCacheItem) {
	t.unConfirmedMsgLock.Lock()
	defer t.unConfirmedMsgLock.Unlock()
	t.unConfirmedMessages[key] = cacheItem
}

func (t *TssCommon) removeKey(key string) {
	t.unConfirmedMsgLock.Lock()
	defer t.unConfirmedMsgLock.Unlock()
	delete(t.unConfirmedMessages, key)
}

func (t *TssCommon) ProcessInboundMessages(finishChan chan struct{}, wg *sync.WaitGroup) {
	t.logger.Debug().Msg("start processing inbound messages")
	defer wg.Done()
	defer t.logger.Debug().Msg("stop processing inbound messages")
	for {
		select {
		case <-finishChan:
			return
		case m, ok := <-t.TssMsg:
			if !ok {
				return
			}
			var wrappedMsg messages.WrappedMessage
			if err := json.Unmarshal(m.Payload, &wrappedMsg); nil != err {
				t.logger.Error().Err(err).Msg("fail to unmarshal wrapped message bytes")
				return
			}

			if err := t.ProcessOneMessage(&wrappedMsg, m.PeerID.String(), wrappedMsg.MsgID); err != nil {
				t.logger.Error().Err(err).Msg("fail to process the received message")
			}
		}
	}
}

func (t *TssCommon) GetPartyIDtoP2PIDMap() map[string]peer.ID {
	return t.partyIDtoP2PID
}

func (t *TssCommon) GetNewPartyIDtoP2PIDMap() map[string]peer.ID {
	return t.newPartyIDtoP2PID
}

func (t *TssCommon) SetPartyIDtoP2PIDMap(partyIDtoP2PID map[string]peer.ID) {
	t.partyIDtoP2PID = partyIDtoP2PID
}

func (t *TssCommon) SetNewPartyIDtoP2PIDMap(partyIDtoP2PID map[string]peer.ID) {
	t.newPartyIDtoP2PID = partyIDtoP2PID
}

func (t *TssCommon) GetLocalPeerID() *peer.ID {
	return t.localPeerID
}

func (t *TssCommon) SetPeers(peers []peer.ID) error {
	t.P2PPeersLock.Lock()
	defer t.P2PPeersLock.Unlock()
	t.P2PPeers = peers

	return nil
}

// Set number of old committe
func (t *TssCommon) SetNumOldCommitte(oldCommitte int) {
	t.numOldCommitte = oldCommitte
}
