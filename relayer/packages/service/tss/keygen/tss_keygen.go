package keygen

import (
	"errors"
	"fmt"
	"sync"
	"time"

	bcrypto "github.com/bnb-chain/tss-lib/crypto"
	bkg "github.com/bnb-chain/tss-lib/ecdsa/keygen"
	btss "github.com/bnb-chain/tss-lib/tss"

	types "github.com/Datura-ai/relayer/packages/domain/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/blame"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
	"github.com/Datura-ai/relayer/packages/service/tss/conversion"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	"github.com/Datura-ai/relayer/packages/service/tss/storage"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type TssKeygen struct {
	logger          zerolog.Logger
	msgID           string
	localNodePubKey string
	preParams       *bkg.LocalPreParams
	tssCommonStruct *common.TssCommon
	stopChan        chan struct{} // channel to indicate whether we should stop
	localParty      *btss.PartyID
	stateManager    storage.LocalStateManager
	commStopChan    chan struct{}
	p2pComm         *p2p.Communication
}

func NewTssKeygen(
	localP2PID string,
	localNodePubKey string,
	stopChan chan struct{},
	preParam *bkg.LocalPreParams,
	msgID string,
	stateManager storage.LocalStateManager,
	p2pComm *p2p.Communication,
	tssConfig common.TssConfig,
	privateKey crypto.PrivKey,
) *TssKeygen {
	tssCommon, err := common.NewTssCommon(
		p2pComm.GetLocalPeerID(),
		p2pComm.BroadcastMsgChan,
		tssConfig,
		privateKey,
		1,
		common.KEYGEN_TSS_HANDLE_ID,
	)

	if err != nil {
		return nil
	}

	return &TssKeygen{
		logger: log.With().
			Str("module", "keygen").
			Str("msgID", msgID).Logger(),
		msgID:           msgID,
		localNodePubKey: localNodePubKey,
		preParams:       preParam,
		tssCommonStruct: tssCommon,
		stopChan:        stopChan,
		localParty:      nil,
		stateManager:    stateManager,
		commStopChan:    make(chan struct{}),
		p2pComm:         p2pComm,
	}
}

func (tKeyGen *TssKeygen) GetTssKeyGenChannels() chan *p2p.Message {
	return tKeyGen.tssCommonStruct.TssMsg
}

func (tKeyGen *TssKeygen) GetTssCommonStruct() *common.TssCommon {
	return tKeyGen.tssCommonStruct
}

func (t *TssKeygen) GenerateNewKey_Ecdsa(keygenReq Request) (*bcrypto.ECPoint, error) {

	partyIDs, localPartyID, err := conversion.GetParties(keygenReq.Keys, t.localNodePubKey, false)
	if err != nil {
		return nil, fmt.Errorf("fail to get keygen parties: %w", err)
	}

	keyGenLocalStateItem := storage.KeygenLocalState{
		ParticipantKeys: keygenReq.Keys,
		LocalPartyKey:   t.localNodePubKey,
	}

	threshold, err := conversion.GetThreshold(len(keygenReq.Keys))
	if err != nil {
		return nil, err
	}
	keygenPartyMap := new(sync.Map)
	ctx := btss.NewPeerContext(partyIDs)
	params := btss.NewParameters(btss.S256(), ctx, localPartyID, len(partyIDs), threshold)
	outCh := make(chan btss.Message, len(partyIDs))
	endCh := make(chan bkg.LocalPartySaveData, len(partyIDs))
	errChan := make(chan error)

	if t.preParams == nil {
		t.logger.Error().Err(err).Msg("error, empty pre-parameters")
		return nil, errors.New("error, empty pre-parameters")
	}
	blameMgr := t.tssCommonStruct.GetBlameMgr()
	keygenParty := bkg.NewLocalParty(params, outCh, endCh, *t.preParams)

	partyIDMap := conversion.SetupPartyIDMap(partyIDs)
	partyIDtoP2PID := t.tssCommonStruct.GetPartyIDtoP2PIDMap()
	if err1 := conversion.SetupIDMaps(partyIDMap, partyIDtoP2PID); err1 != nil {
		t.logger.Error().Err(err1).Msgf("error in creating mapping between partyID and P2P ID")
		return nil, err1
	}

	t.tssCommonStruct.SetPartyIDtoP2PIDMap(partyIDtoP2PID)
	// we never run multi keygen, so the moniker is set to default empty value
	keygenPartyMap.Store("", keygenParty)
	partyInfo := &common.PartyInfo{
		PartyMap:   keygenPartyMap,
		PartyIDMap: partyIDMap,
		PartyID:    localPartyID,
	}

	t.tssCommonStruct.SetPartyInfo(partyInfo)
	blameMgr.SetPartyInfo(keygenPartyMap, nil, partyIDMap, true)
	var keygenWg sync.WaitGroup
	keygenWg.Add(2)
	// start keygen
	go func() {
		defer keygenWg.Done()
		defer t.logger.Info().Msg(">>>>>>>>>>>>>.keygenParty started")
		if err := keygenParty.Start(); nil != err {
			t.logger.Error().Err(err).Msg("fail to start keygen party")
			close(errChan)
		}
	}()
	go t.tssCommonStruct.ProcessInboundMessages(t.commStopChan, &keygenWg)

	pk, err := t.processKeygen_Ecdsa(errChan, outCh, endCh, keyGenLocalStateItem)
	if err != nil && err != types.ErrExitSignal {
		close(t.stopChan)
		close(t.commStopChan)
		return nil, fmt.Errorf("fail to process key sign: %w", err)
	}

	select {
	case <-time.After(time.Second * 10):
		close(t.commStopChan)

	case <-t.tssCommonStruct.GetTaskDone():
		close(t.commStopChan)
	}

	keygenWg.Wait()
	return pk, err
}

func (t *TssKeygen) processKeygen_Ecdsa(errChan chan error,
	outCh <-chan btss.Message,
	endCh <-chan bkg.LocalPartySaveData,
	keyGenLocalStateItem storage.KeygenLocalState,
) (*bcrypto.ECPoint, error) {
	defer t.logger.Debug().Msg("finished keygen process")
	t.logger.Debug().Msg("start to read messages from local party")
	tssConf := t.tssCommonStruct.GetConf()
	blameMgr := t.tssCommonStruct.GetBlameMgr()
	for {
		select {
		case err := <-errChan: // when keygenParty return
			t.logger.Error().Err(err).Msg("key gen failed")
			return nil, types.ErrChannelClosed

		case <-t.stopChan: // when TSS processor receive signal to quit
			return nil, types.ErrExitSignal

		case <-time.After(tssConf.KeygenTimeout):
			// we bail out after KeygenTimeoutSeconds
			t.logger.Error().Msgf("Keygen - fail to generate message with %s", tssConf.KeygenTimeout.String())
			lastMsg := blameMgr.GetLastMsg()
			failReason := blameMgr.GetBlame().FailReason
			if failReason == "" {
				failReason = blame.TssTimeout
			}
			if lastMsg == nil {
				t.logger.Error().Msg("fail to start the keygen, the last produced message of this node is none")
				return nil, errors.New("timeout before shared message is generated")
			}
			blameNodesUnicast, err := blameMgr.GetUnicastBlame(messages.KEYGEN2aUnicast)
			if err != nil {
				t.logger.Error().Err(err).Msg("error in get unicast blame")
			}
			t.tssCommonStruct.P2PPeersLock.RLock()
			threshold, err := conversion.GetThreshold(len(t.tssCommonStruct.P2PPeers) + 1)
			t.tssCommonStruct.P2PPeersLock.RUnlock()
			if err != nil {
				t.logger.Error().Err(err).Msg("error in get the threshold to generate blame")
			}

			if len(blameNodesUnicast) > 0 && len(blameNodesUnicast) <= threshold {
				blameMgr.GetBlame().SetBlame(failReason, blameNodesUnicast, true)
			}
			blameNodesBroadcast, err := blameMgr.GetBroadcastBlame(lastMsg.Type())
			if err != nil {
				t.logger.Error().Err(err).Msg("error in get broadcast blame")
			}
			blameMgr.GetBlame().AddBlameNodes(blameNodesBroadcast...)

			// if we cannot find the blame node, we check whether everyone send me the share
			if len(blameMgr.GetBlame().BlameNodes) == 0 {
				blameNodesMisingShare, isUnicast, err := blameMgr.TssMissingShareBlame(messages.TSSKEYGENROUNDS)
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
			t.logger.Debug().Msgf(">>>>>>>>>>msg: %s", msg.String())
			blameMgr.SetLastMsg(msg)
			err := t.tssCommonStruct.ProcessOutCh(msg, messages.TSSKeygenMsg, t.msgID)
			if err != nil {
				t.logger.Error().Err(err).Msg("fail to process the message")
				return nil, err
			}

		case msg := <-endCh:
			t.logger.Debug().Msgf("keygen finished successfully: %s", msg.ECDSAPub.Y().String())
			if err := t.tssCommonStruct.NotifyTaskDone(t.msgID); err != nil {
				t.logger.Error().Err(err).Msg("fail to broadcast the keysign done")
			}

			pubKey, err := conversion.GetTssPubKey(msg.ECDSAPub)
			if err != nil {
				return nil, fmt.Errorf("fail to get tss pubkey: %w", err)
			}

			t.logger.Info().Msg(pubKey)

			keyGenLocalStateItem.LocalData = msg
			keyGenLocalStateItem.PubKey = pubKey
			if err := t.stateManager.SaveLocalState(keyGenLocalStateItem); err != nil {
				return nil, fmt.Errorf("fail to save keygen result to storage: %w", err)
			}
			address := t.p2pComm.ExportPeerAddress()
			if err := t.stateManager.SaveAddressBook(address); err != nil {
				t.logger.Error().Err(err).Msg("fail to save the peer addresses")
			}

			return msg.ECDSAPub, nil
		}
	}
}
