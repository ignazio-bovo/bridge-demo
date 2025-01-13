package tss

import (
	"time"

	domain "github.com/Datura-ai/relayer/packages/domain/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/blame"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
	"github.com/Datura-ai/relayer/packages/service/tss/conversion"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	"github.com/Datura-ai/relayer/packages/service/tss/resharing"
)

func (t *TssServer) Keyresharing(req resharing.Request) (resharing.Response, error) {
	t.tssKeygenLocker.Lock()
	defer t.tssKeygenLocker.Unlock()
	status := common.Success
	msgID, err := t.requestToMsgId(req)
	if err != nil {
		return resharing.Response{}, err
	}

	keyReSharingInstance := resharing.NewTssKeyresharing(
		t.GetLocalPeerID(),
		t.localNodePubKey,
		t.stopChan,
		msgID,
		t.stateManager,
		t.p2pCommunication,
		t.preParams,
		t.privateKey,
		t.conf,
	)

	p2pMessageChannels := keyReSharingInstance.GetTssKeyResharingChannels()
	t.setSubscribeTss(p2pMessageChannels, msgID)

	defer func() {
		t.cancelSubscribeTss(msgID)

		t.p2pCommunication.ReleaseStream(msgID)
		t.partyCoordinator.ReleaseStream(msgID)
	}()
	participants := make([]string, 0)
	participants = append(participants, req.Keys...)
	participants = append(participants, req.NewKeys...)

	threshold, err := conversion.GetThreshold(len(participants) + 1)
	if err != nil {
		return resharing.Response{}, err
	}

	sigChan := make(chan string)
	blameMgr := keyReSharingInstance.GetTssCommonStruct().GetBlameMgr()
	onlinePeers, leader, _, errJoinParty := t.joinParty(msgID, string(domain.KEYSIGN_MSG_TYPE), req.Version, req.BlockHeight, participants, threshold, sigChan, []byte(req.Keys[0]))
	if errJoinParty != nil {
		// this indicate we are processing the leaderless join party
		if leader == "NONE" {
			if onlinePeers == nil {
				t.logger.Error().Err(err).Msg("error before we start join party")
				return resharing.NewResponse(common.Fail, blame.NewBlame(blame.InternalError, []blame.Node{})), errJoinParty
			}
			blameNodes, err := blameMgr.NodeSyncBlame(req.Keys, onlinePeers)
			if err != nil {
				t.logger.Err(errJoinParty).Msg("fail to get peers to blame")
			}
			// make sure we blame the leader as well
			t.logger.Error().Err(errJoinParty).Msgf("fail to form key resharing party with online:%v", onlinePeers)
			return resharing.NewResponse(common.Fail, blameNodes), errJoinParty
		}

		var blameLeader blame.Blame
		var blameNodes blame.Blame
		blameNodes, err = blameMgr.NodeSyncBlame(req.Keys, onlinePeers)
		if err != nil {
			t.logger.Err(err).Msg("fail to get peers to blame")
		}

		leaderPubKey, err := conversion.GetPubKeyFromPeerID(leader)
		if err != nil {
			t.logger.Error().Err(errJoinParty).Msgf("fail to convert the peerID to public key with leader %s", leader)
			blameLeader = blame.NewBlame(blame.TssSyncFail, []blame.Node{})
		} else {
			blameLeader = blame.NewBlame(blame.TssSyncFail, []blame.Node{{Pubkey: leaderPubKey, BlameData: nil, BlameSignature: nil}})
		}

		if len(onlinePeers) != 0 {
			blameNodes.AddBlameNodes(blameLeader.BlameNodes...)
		} else {
			blameNodes = blameLeader
		}

		t.logger.Error().Err(errJoinParty).Msgf("fail to form keygen party with online:%v", onlinePeers)

		return resharing.NewResponse(common.Fail, blameNodes), errJoinParty

	}

	tssCommonStruct := keyReSharingInstance.GetTssCommonStruct()
	if err := tssCommonStruct.SetPeers(onlinePeers); err != nil {
		t.logger.Error().Err(err).Msgf("fail to set online peers for keygen:%v", onlinePeers)
		blameNodes := *blameMgr.GetBlame()
		return resharing.Response{
			Status: common.Fail,
			Blame:  blameNodes,
		}, errJoinParty
	}

	t.logger.Debug().Msg("keyresharing party formed")
	// the statistic of keyresharing only care about Tss it self, even if the
	// following http response aborts, it still counted as a successful keyresharing
	// as the Tss model runs successfully.
	beforeKeyresharing := time.Now()
	err = keyReSharingInstance.ResharingKey(req)
	keyresharingTime := time.Since(beforeKeyresharing)
	if err != nil {
		t.logger.Error().Err(err).Msg("err in keyresharing")
		blameNodes := *blameMgr.GetBlame()
		return resharing.NewResponse(common.Fail, blameNodes), err
	} else {
		t.logger.Info().Msgf("keyresharing success with time: %v", keyresharingTime)
	}

	blameNodes := *blameMgr.GetBlame()
	return resharing.NewResponse(status, blameNodes), nil
}

func (t *TssServer) setSubscribeTss(keyresharingMsgChannels chan *p2p.Message, msgID string) {
	t.p2pCommunication.SetSubscribe(messages.TSSKeyresharingMsgOld, msgID, keyresharingMsgChannels)
	t.p2pCommunication.SetSubscribe(messages.TSSKeyresharingMsgNew, msgID, keyresharingMsgChannels)
	t.p2pCommunication.SetSubscribe(messages.TSSKeyresharingVerMsg, msgID, keyresharingMsgChannels)
	t.p2pCommunication.SetSubscribe(messages.TSSControlMsg, msgID, keyresharingMsgChannels)
	t.p2pCommunication.SetSubscribe(messages.TSSTaskDone, msgID, keyresharingMsgChannels)
}

func (t *TssServer) cancelSubscribeTss(msgID string) {
	t.p2pCommunication.CancelSubscribe(messages.TSSKeyresharingMsgOld, msgID)
	t.p2pCommunication.CancelSubscribe(messages.TSSKeyresharingMsgNew, msgID)
	t.p2pCommunication.CancelSubscribe(messages.TSSKeyresharingVerMsg, msgID)
	t.p2pCommunication.CancelSubscribe(messages.TSSControlMsg, msgID)
	t.p2pCommunication.CancelSubscribe(messages.TSSTaskDone, msgID)
}
