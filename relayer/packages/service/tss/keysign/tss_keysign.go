package keysign

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"

	tsslibcommon "github.com/bnb-chain/tss-lib/common"
	"github.com/bnb-chain/tss-lib/ecdsa/signing"
	btss "github.com/bnb-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p/core/crypto"
	"go.uber.org/atomic"

	types "github.com/Datura-ai/relayer/packages/domain/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/blame"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
	"github.com/Datura-ai/relayer/packages/service/tss/conversion"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	"github.com/Datura-ai/relayer/packages/service/tss/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type TssKeysign struct {
	localNodeID     string
	logger          zerolog.Logger
	tssCommonStruct *common.TssCommon
	stopChan        chan struct{} // channel to indicate whether we should stop
	localParties    []*btss.PartyID
	commStopChan    chan struct{}
	p2pComm         *p2p.Communication
	stateManager    storage.LocalStateManager
	handleID        uint64
}

func NewTssKeysign(
	localNodeID string,
	stopChan chan struct{},
	msgID string,
	p2pComm *p2p.Communication,
	stateManager storage.LocalStateManager,
	privKey crypto.PrivKey,
	tssConfig common.TssConfig,
	handleID uint64,
) *TssKeysign {
	logItems := []string{"keySign", msgID}
	tssCommon, err := common.NewTssCommon(
		p2pComm.GetLocalPeerID(),
		p2pComm.BroadcastMsgChan,
		tssConfig,
		privKey,
		1,
		handleID,
	)
	if err != nil {
		return nil
	}

	return &TssKeysign{
		localNodeID:     localNodeID,
		logger:          log.With().Strs("module", logItems).Logger(),
		tssCommonStruct: tssCommon,
		stopChan:        stopChan,
		localParties:    make([]*btss.PartyID, 0),
		commStopChan:    make(chan struct{}),
		p2pComm:         p2pComm,
		stateManager:    stateManager,
		handleID:        handleID,
	}
}

func (t *TssKeysign) GetTssKeySignChannels() chan *p2p.Message {
	return t.tssCommonStruct.TssMsg
}

func (t *TssKeysign) GetTssCommonStruct() *common.TssCommon {
	return t.tssCommonStruct
}

func (t *TssKeysign) startBatchSigning(keySignPartyMap *sync.Map, msgNum int) bool {
	// start the batch sign
	var keySignWg sync.WaitGroup
	ret := atomic.NewBool(true)
	keySignWg.Add(msgNum)
	keySignPartyMap.Range(func(key, value interface{}) bool {
		eachParty := value.(btss.Party)
		go func(eachParty btss.Party) {
			defer keySignWg.Done()
			if err := eachParty.Start(); err != nil {
				t.logger.Error().Err(err).Msg("fail to start key sign party")
				ret.Store(false)
			}
			t.logger.Info().Msgf("local party(%s) %s is ready", eachParty.PartyID().Id, eachParty.PartyID().Moniker)
		}(eachParty)
		return true
	})
	keySignWg.Wait()
	return ret.Load()
}

// signMessage
func (t *TssKeysign) SignMessage(msgsToSign [][]byte, localStateItem storage.KeygenLocalState, parties []string, msgID string) ([]*tsslibcommon.SignatureData, error) {
	partiesID, localPartyID, err := conversion.GetParties(localStateItem.ParticipantKeys, localStateItem.LocalPartyKey, false)
	if err != nil {
		return nil, fmt.Errorf("fail to form key sign party: %w", err)
	}

	if !common.Contains(partiesID, localPartyID) {
		t.logger.Info().Msgf("we are not in this rounds key sign")
		return nil, nil
	}
	threshold, err := conversion.GetThreshold(len(localStateItem.ParticipantKeys))
	if err != nil {
		return nil, errors.New("fail to get threshold")
	}

	outCh := make(chan btss.Message, 2*len(partiesID)*len(msgsToSign))
	endCh := make(chan tsslibcommon.SignatureData, len(partiesID)*len(msgsToSign))
	errCh := make(chan struct{}, len(partiesID)*len(msgsToSign))

	keySignPartyMap := new(sync.Map)
	for i, val := range msgsToSign {
		// Use pure []byte to big.Int conversion
		// m, err := common.MsgToHashInt(val)
		m, err := common.ConvertMsgToBigInt(val)
		if err != nil {
			return nil, fmt.Errorf("fail to convert msg to hash int: %w", err)
		}
		encodedString := hex.EncodeToString(m.Bytes())
		t.logger.Info().Msgf("Hash encoded string: %s", encodedString)

		moniker := strconv.Itoa(i)
		ctx := btss.NewPeerContext(partiesID)
		t.logger.Info().Msgf("message: (%s) keysign parties: %+v", m.String(), parties)
		t.localParties = nil
		params := btss.NewParameters(btss.S256(), ctx, localPartyID, len(partiesID), threshold)
		keySignParty := signing.NewLocalParty(m, params, localStateItem.LocalData, outCh, endCh)
		keySignPartyMap.Store(moniker, keySignParty)
	}

	blameMgr := t.tssCommonStruct.GetBlameMgr()
	partyIDtoP2PID := t.tssCommonStruct.GetPartyIDtoP2PIDMap()
	partyIDMap := conversion.SetupPartyIDMap(partiesID)
	err1 := conversion.SetupIDMaps(partyIDMap, partyIDtoP2PID)
	if err1 != nil {
		t.logger.Error().Err(err).Msgf("error in creating mapping between partyID and P2P ID")
		return nil, err
	}

	t.tssCommonStruct.SetPartyIDtoP2PIDMap(partyIDtoP2PID)
	t.tssCommonStruct.SetPartyInfo(&common.PartyInfo{
		PartyMap:   keySignPartyMap,
		PartyIDMap: partyIDMap,
		PartyID:    localPartyID,
	})

	blameMgr.SetPartyInfo(keySignPartyMap, nil, partyIDMap, true)

	t.tssCommonStruct.P2PPeersLock.Lock()
	t.tssCommonStruct.P2PPeers = conversion.GetPeersID(partyIDtoP2PID, t.tssCommonStruct.GetLocalPeerID().String())
	t.tssCommonStruct.P2PPeersLock.Unlock()

	var keySignWg sync.WaitGroup
	keySignWg.Add(2)
	// start the key sign
	go func() {
		defer keySignWg.Done()
		ret := t.startBatchSigning(keySignPartyMap, len(msgsToSign))
		if !ret {
			close(errCh)
		}
	}()
	go t.tssCommonStruct.ProcessInboundMessages(t.commStopChan, &keySignWg)
	results, err := t.processKeySign(len(msgsToSign), errCh, outCh, endCh, msgID)
	if err != nil {
		close(t.commStopChan)
		return nil, fmt.Errorf("fail to process key sign: %w", err)
	}

	select {
	case <-time.After(time.Second * 5):
		close(t.commStopChan)
	case <-t.tssCommonStruct.GetTaskDone():
		close(t.commStopChan)
	}
	keySignWg.Wait()

	t.logger.Info().Msgf("%s successfully sign the message", t.p2pComm.GetHost().ID().String())
	sort.SliceStable(results, func(i, j int) bool {
		a := new(big.Int).SetBytes(results[i].M)
		b := new(big.Int).SetBytes(results[j].M)

		return a.Cmp(b) != -1
	})

	return results, nil
}

func (t *TssKeysign) processKeySign(reqNum int, errChan chan struct{}, outCh <-chan btss.Message, endCh <-chan tsslibcommon.SignatureData, msgID string) ([]*tsslibcommon.SignatureData, error) {
	defer t.logger.Debug().Msg("key sign finished")
	t.logger.Debug().Msg("start to read messages from local party")
	var signatures []*tsslibcommon.SignatureData

	tssConf := t.tssCommonStruct.GetConf()
	blameMgr := t.tssCommonStruct.GetBlameMgr()

	for {
		select {
		case <-errChan: // when key sign return
			t.logger.Error().Msg("key sign failed")
			return nil, types.ErrChannelClosed
		case <-t.stopChan: // when TSS processor receive signal to quit
			return nil, types.ErrExitSignal
		case <-time.After(tssConf.KeysignTimeout):
			// we bail out after KeySignTimeoutSeconds
			t.logger.Error().Msgf("fail to sign message with %s", tssConf.KeysignTimeout.String())
			lastMsg := blameMgr.GetLastMsg()
			failReason := blameMgr.GetBlame().FailReason
			if failReason == "" {
				failReason = blame.TssTimeout
			}

			t.tssCommonStruct.P2PPeersLock.RLock()
			threshold, err := conversion.GetThreshold(len(t.tssCommonStruct.P2PPeers) + 1)
			t.tssCommonStruct.P2PPeersLock.RUnlock()
			if err != nil {
				t.logger.Error().Err(err).Msg("error in get the threshold for generate blame")
			}
			if !lastMsg.IsBroadcast() {
				blameNodesUnicast, err := blameMgr.GetUnicastBlame(lastMsg.Type())
				if err != nil {
					t.logger.Error().Err(err).Msg("error in get unicast blame")
				}
				if len(blameNodesUnicast) > 0 && len(blameNodesUnicast) <= threshold {
					blameMgr.GetBlame().SetBlame(failReason, blameNodesUnicast, true)
				}
			} else {
				blameNodesUnicast, err := blameMgr.GetUnicastBlame(conversion.GetPreviousKeySignUicast(lastMsg.Type()))
				if err != nil {
					t.logger.Error().Err(err).Msg("error in get unicast blame")
				}
				if len(blameNodesUnicast) > 0 && len(blameNodesUnicast) <= threshold {
					blameMgr.GetBlame().SetBlame(failReason, blameNodesUnicast, true)
				}
			}

			blameNodesBroadcast, err := blameMgr.GetBroadcastBlame(lastMsg.Type())
			if err != nil {
				t.logger.Error().Err(err).Msg("error in get broadcast blame")
			}
			blameMgr.GetBlame().AddBlameNodes(blameNodesBroadcast...)

			// if we cannot find the blame node, we check whether everyone send me the share
			if len(blameMgr.GetBlame().BlameNodes) == 0 {
				blameNodesMisingShare, isUnicast, err := blameMgr.TssMissingShareBlame(messages.TSSKEYSIGNROUNDS)
				if err != nil {
					t.logger.Error().Err(err).Msg("fail to get the node of missing share ")
				}

				if len(blameNodesMisingShare) > 0 && len(blameNodesMisingShare) <= threshold {
					blameMgr.GetBlame().AddBlameNodes(blameNodesMisingShare...)
					blameMgr.GetBlame().IsUnicast = isUnicast
				}
			}

			return nil, blame.ErrTssTimeOut
		case msg := <-outCh:
			t.logger.Debug().Msgf(">>>>>>>>>>key sign msg: %s", msg.String())
			t.tssCommonStruct.GetBlameMgr().SetLastMsg(msg)
			err := t.tssCommonStruct.ProcessOutCh(msg, messages.TSSKeysignMsg, msgID)
			if err != nil {
				return nil, err
			}
		case msg := <-endCh:
			signatures = append(signatures, &msg)
			if len(signatures) == reqNum {
				t.logger.Debug().Msg("we have done the key sign")
				err := t.tssCommonStruct.NotifyTaskDone(msgID)
				if err != nil {
					t.logger.Error().Err(err).Msg("fail to broadcast the keysign done")
				}
				//export the address book
				address := t.p2pComm.ExportPeerAddress()
				if err := t.stateManager.SaveAddressBook(address); err != nil {
					t.logger.Error().Err(err).Msg("fail to save the peer addresses")
				}
				return signatures, nil
			}
		}
	}
}
