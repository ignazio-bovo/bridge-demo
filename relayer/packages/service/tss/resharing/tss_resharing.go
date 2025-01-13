package resharing

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/ecdsa/resharing"
	btss "github.com/bnb-chain/tss-lib/tss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	types "github.com/Datura-ai/relayer/packages/domain/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/blame"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
	"github.com/Datura-ai/relayer/packages/service/tss/conversion"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	"github.com/Datura-ai/relayer/packages/service/tss/storage"
	"github.com/bnb-chain/tss-lib/crypto"
	tcrypto "github.com/libp2p/go-libp2p/core/crypto"
)

type TssKeyresharing struct {
	logger          zerolog.Logger
	msgID           string
	localNodePubKey string
	tssCommonStruct *common.TssCommon
	stopChan        chan struct{} // channel to indicate whether we should stop
	localParty      *btss.PartyID
	stateManager    storage.LocalStateManager
	commStopChan    chan struct{}
	p2pComm         *p2p.Communication
	preParams       *keygen.LocalPreParams
}

func NewTssKeyresharing(
	localP2PID string,
	localNodePubKey string,
	stopChan chan struct{},
	msgID string,
	stateManager storage.LocalStateManager,
	p2pComm *p2p.Communication,
	preParams *keygen.LocalPreParams,
	privKey tcrypto.PrivKey,
	tssConfig common.TssConfig,
) *TssKeyresharing {
	tssCommon, err := common.NewTssCommon(
		p2pComm.GetLocalPeerID(),
		p2pComm.BroadcastMsgChan,
		tssConfig,
		privKey,
		1,
		common.KEYGEN_TSS_HANDLE_ID,
	)

	if err != nil {
		return nil
	}

	return &TssKeyresharing{
		logger: log.With().
			Str("module", "resharing").
			Str("msgID", msgID).
			Str("localP2PID", localP2PID).
			Str("localNodePubKey", localNodePubKey).Logger(),
		msgID:           msgID,
		localNodePubKey: localNodePubKey,
		tssCommonStruct: tssCommon,
		stopChan:        stopChan,
		localParty:      nil,
		stateManager:    stateManager,
		commStopChan:    make(chan struct{}),
		p2pComm:         p2pComm,
		preParams:       preParams,
	}
}

func (t *TssKeyresharing) GetTssKeyResharingChannels() chan *p2p.Message {
	return t.tssCommonStruct.TssMsg
}

// GetTssCommonStruct returns tss common
func (t *TssKeyresharing) GetTssCommonStruct() *common.TssCommon {
	return t.tssCommonStruct
}

// ResharingKey process key resharing request with other peers and returns public key for new members
func (t *TssKeyresharing) ResharingKey(reSharingReq Request) (err error) {
	var keyReShareWg sync.WaitGroup

	partyIDs, partyID, err := conversion.GetParties(reSharingReq.Keys, t.localNodePubKey, false)
	if err != nil {
		t.logger.Error().Err(err).Msgf("error while getting parties for old committee")
		return err
	}

	newCommittee := append(reSharingReq.Keys, reSharingReq.NewKeys...)
	newPartyIDs, newPartyID, err := conversion.GetParties(newCommittee, t.localNodePubKey, true)
	if err != nil {
		t.logger.Error().Err(err).Msgf("error while getting parties for new committee")
		return err
	}

	errChan := make(chan error)
	defer close(errChan)

	// set the number of old committee
	t.tssCommonStruct.SetNumOldCommitte(len(partyIDs))

	// start resharing party for old and new committee
	if err := t.StartResharingParty(errChan, &keyReShareWg, partyIDs, newPartyIDs, partyID, newPartyID, newCommittee); err != nil {
		t.logger.Error().Err(err).Msg("error while starting resharing party")
		close(t.commStopChan)
		return err
	}

	select {
	case <-time.After(time.Second * 5):
		close(t.commStopChan)

	case <-t.tssCommonStruct.GetTaskDone():
		close(t.commStopChan)

	case err = <-errChan: // when re sharing returns error
		t.logger.Error().Err(err).Msg("re sharing failed")
		close(t.commStopChan)
	}

	keyReShareWg.Wait()

	if err := t.tssCommonStruct.NotifyTaskDone(t.msgID); err != nil {
		t.logger.Error().Err(err).Msg("fail to broadcast the keyresharing done")
	}

	return err
}

func (t *TssKeyresharing) StartResharingParty(
	errChan chan error,
	wg *sync.WaitGroup,
	partyIDs, newPartyIDs []*btss.PartyID,
	partyID, newPartyID *btss.PartyID,
	newCommitee []string,
) error {
	var resharingParty btss.Party

	resharingPartyMap, newResharingPartyMap := new(sync.Map), new(sync.Map)
	blameMgr := t.tssCommonStruct.GetBlameMgr()
	partyIDMap := conversion.SetupPartyIDMap(partyIDs)
	newPartyIDMap := conversion.SetupPartyIDMap(newPartyIDs)

	partyIDtoP2PID := t.tssCommonStruct.GetPartyIDtoP2PIDMap()
	newPartyIDtoP2PID := t.tssCommonStruct.GetNewPartyIDtoP2PIDMap()
	if err1 := conversion.SetupIDMaps(partyIDMap, partyIDtoP2PID); err1 != nil {
		t.logger.Error().Err(err1).Msg("error in creating mapping between partyID and P2P ID")
		return err1
	}
	if err2 := conversion.SetupIDMaps(newPartyIDMap, newPartyIDtoP2PID); err2 != nil {
		t.logger.Error().Err(err2).Msg("error in creating mapping between partyID and blame P2P ID")
		return err2
	}

	t.tssCommonStruct.SetPartyIDtoP2PIDMap(partyIDtoP2PID)
	t.tssCommonStruct.SetNewPartyIDtoP2PIDMap(newPartyIDtoP2PID)

	ctx, partyIDsCount := btss.NewPeerContext(partyIDs), len(partyIDs)
	threshold, err := conversion.GetThreshold(partyIDsCount)
	if err != nil {
		return err
	}

	newCtx, newPartyIDsCount := btss.NewPeerContext(newPartyIDs), len(newPartyIDs)
	newThreshold, err := conversion.GetThreshold(newPartyIDsCount)
	if err != nil {
		return err
	}

	bothCommiteeLen := partyIDsCount + newPartyIDsCount
	outCh := make(chan btss.Message, bothCommiteeLen*2)
	endCh := make(chan keygen.LocalPartySaveData, bothCommiteeLen)

	// If it is an existing node,
	if partyID != nil {
		params := btss.NewReSharingParameters(btss.S256(), ctx, newCtx, partyID, newPartyIDsCount, threshold, newPartyIDsCount, newThreshold)
		if params == nil {
			t.logger.Error().Err(err).Msg("error, empty re-sharing-parameters")
			return errors.New("error, empty pre-parameters")
		}

		localStateItem, err := t.stateManager.GetLocalState(t.localNodePubKey)
		if err != nil {
			return err
		}

		resharingParty = resharing.NewLocalParty(params, localStateItem.LocalData, outCh, endCh)
		// we never run multi key resharing, so the moniker is set to default empty value
		resharingPartyMap.Store("", resharingParty)
	}

	newParams := btss.NewReSharingParameters(btss.S256(), ctx, newCtx, newPartyID, newPartyIDsCount, threshold, newPartyIDsCount, newThreshold)
	if newParams == nil {
		t.logger.Error().Err(err).Msg("error, empty re-sharing-parameters")
		return errors.New("error, empty pre-parameters")
	}

	newLocalState := t.newLocalPartySaveData(newPartyIDsCount)
	newResharingParty := resharing.NewLocalParty(newParams, newLocalState, outCh, endCh)
	// we never run multi key resharing, so the moniker is set to default empty value
	newResharingPartyMap.Store("", newResharingParty)

	partyInfo := &common.PartyInfo{
		PartyMap:      resharingPartyMap,
		NewPartyMap:   newResharingPartyMap,
		PartyIDMap:    partyIDMap,
		NewPartyIDMap: newPartyIDMap,
		PartyID:       partyID,
		NewPartyID:    newPartyID,
	}
	t.tssCommonStruct.SetPartyInfo(partyInfo)
	blameMgr.SetPartyInfo(resharingPartyMap, newResharingPartyMap, partyIDMap, true)

	wg.Add(1)
	// start key resharing
	go func() {
		defer wg.Done()
		defer t.logger.Debug().Msg(">>>>>>>>>>>>>new resharingParty started")
		if err := newResharingParty.Start(); nil != err {
			t.logger.Error().Err(err).Msg("fail to start re sharing party")
			errChan <- err
		}
	}()

	if partyID != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer t.logger.Debug().Msg(">>>>>>>>>>>>>old resharingParty started")
			if err := resharingParty.Start(); nil != err {
				t.logger.Error().Err(err).Msg("fail to start re sharing party")
				errChan <- err
			}
		}()
	}

	wg.Add(1)
	go t.tssCommonStruct.ProcessInboundMessages(t.commStopChan, wg)

	return t.ProcessResharingMessages(errChan, outCh, endCh, newCommitee)
}

// processResharing handles resharing channels for out and end messages
func (t *TssKeyresharing) ProcessResharingMessages(
	errChan chan error,
	outCh <-chan btss.Message,
	endCh <-chan keygen.LocalPartySaveData,
	newCommitee []string,
) error {
	timeout := t.tssCommonStruct.GetConf().KeyresharingTimeout

	t.logger.Debug().Msg("start to read messages from local party")
	defer t.logger.Debug().Msg("finished resharing process")

	for {
		select {
		case <-errChan: // when key sign return
			t.logger.Error().Msg("key sign failed")
			return types.ErrChannelClosed
		case <-t.stopChan: // when TSS processor receive signal to quit
			return types.ErrExitSignal

		case <-time.After(timeout):
			// we bail out after KeyresharingTimeoutSeconds
			return blame.ErrTssTimeOut

		case msg := <-outCh:
			t.logger.Debug().Msgf(">>>>>>>>>>msg: %s", msg.String())
			blameMgr := t.tssCommonStruct.GetBlameMgr()
			blameMgr.SetLastMsg(msg)

			dest := msg.GetTo()
			if dest == nil {
				return errors.New("did not expect a msg to have a nil destination during resharing")
			}

			if msg.IsToOldCommittee() || msg.IsToOldAndNewCommittees() {
				err := t.tssCommonStruct.ProcessOutCh(msg, messages.TSSKeyresharingMsgOld, t.msgID)
				if err != nil {
					t.logger.Error().Err(err).Msg("fail to process the message")
					return err
				}
			}

			if !msg.IsToOldCommittee() || msg.IsToOldAndNewCommittees() {
				err := t.tssCommonStruct.ProcessOutCh(msg, messages.TSSKeyresharingMsgNew, t.msgID)
				if err != nil {
					t.logger.Error().Err(err).Msg("fail to process the message")
					return err
				}
			}
		case msg := <-endCh:
			// only new committee members saves new local state
			if msg.Xi != nil {
				originalIdx, err := msg.OriginalIndex()
				if err != nil {
					errChan <- fmt.Errorf("fail to get original index: %w", err)
				}

				// xj test: BigXj == xj*G
				xj := msg.Xi
				gXj := crypto.ScalarBaseMult(btss.S256(), xj)
				BigXj := msg.BigXj[originalIdx]
				if !BigXj.Equals(gXj) {
					t.logger.Error().Msgf("ensure BigX_j == g^x_j: %s", msg.ECDSAPub.Y().String())
					return errors.New("ensure BigX_j == g^x_j")
				}

				pubKey, err := conversion.GetTssPubKey(msg.ECDSAPub)
				if err != nil {
					t.logger.Error().Msgf("fail to get kimachain pubkey: %s", err.Error())

					return fmt.Errorf("fail to get kimachain pubkey: %w", err)
				}

				keyGenLocalStateItem := storage.KeygenLocalState{
					ParticipantKeys: newCommitee,
					LocalPartyKey:   t.localNodePubKey,
					PubKey:          pubKey,
					LocalData:       msg,
					OriginalIdx:     originalIdx,
				}

				// localdata.TrimLocalData(&keyGenLocalStateItem.LocalData, keyGenLocalStateItem.KeyIdx)
				if err := t.stateManager.SaveLocalState(keyGenLocalStateItem); err != nil {
					return fmt.Errorf("fail to save keygen result to storage: %w", err)
				}

				address := t.p2pComm.ExportPeerAddress()
				if err := t.stateManager.SaveAddressBook(address); err != nil {
					t.logger.Error().Err(err).Msg("fail to save the peer addresses")
					return err
				}

				t.logger.Debug().Msgf("key re-sharing finished successfully: %s", msg.ECDSAPub.Y().String())

				return nil
			}
		}
	}
}

func (t *TssKeyresharing) newLocalPartySaveData(partyIDsLen int) keygen.LocalPartySaveData {
	newLocalData := keygen.NewLocalPartySaveData(partyIDsLen)
	if t.preParams != nil {
		newLocalData.LocalPreParams = *t.preParams
	}

	return newLocalData
}
