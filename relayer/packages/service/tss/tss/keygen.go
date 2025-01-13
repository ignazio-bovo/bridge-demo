package tss

import (
	"errors"
	"fmt"
	"time"

	domain "github.com/Datura-ai/relayer/packages/domain/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/blame"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
	"github.com/Datura-ai/relayer/packages/service/tss/conversion"
	"github.com/Datura-ai/relayer/packages/service/tss/keygen"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
)

func (t *TssServer) Keygen(req keygen.Request) (keygen.Response, error) {
	t.tssKeygenLocker.Lock()
	defer t.tssKeygenLocker.Unlock()
	status := common.Success
	msgID, err := t.requestToMsgId(req)
	if err != nil {
		return keygen.Response{}, err
	}
	keygenInstance := keygen.NewTssKeygen(
		t.GetLocalPeerID(),
		t.localNodePubKey,
		t.stopChan,
		t.preParams,
		msgID,
		t.stateManager,
		t.p2pCommunication,
		t.conf,
		t.privateKey,
	)

	if keygenInstance == nil {
		return keygen.Response{
			Status: common.Fail,
			Blame:  blame.NewBlame(blame.InternalError, []blame.Node{}),
		}, errors.New("keygen instance fail")
	}

	keygenMsgChannel := keygenInstance.GetTssKeyGenChannels()
	t.p2pCommunication.SetSubscribe(messages.TSSKeygenMsg, msgID, keygenMsgChannel)
	t.p2pCommunication.SetSubscribe(messages.TSSKeygenVerMsg, msgID, keygenMsgChannel)
	t.p2pCommunication.SetSubscribe(messages.TSSControlMsg, msgID, keygenMsgChannel)
	t.p2pCommunication.SetSubscribe(messages.TSSTaskDone, msgID, keygenMsgChannel)

	defer func() {
		t.p2pCommunication.CancelSubscribe(messages.TSSKeygenMsg, msgID)
		t.p2pCommunication.CancelSubscribe(messages.TSSKeygenVerMsg, msgID)
		t.p2pCommunication.CancelSubscribe(messages.TSSControlMsg, msgID)
		t.p2pCommunication.CancelSubscribe(messages.TSSTaskDone, msgID)

		t.p2pCommunication.ReleaseStream(msgID)
		t.partyCoordinator.ReleaseStream(msgID)
	}()
	sigChan := make(chan string)
	blameMgr := keygenInstance.GetTssCommonStruct().GetBlameMgr()
	onlinePeers, leader, _, errJoinParty := t.joinParty(msgID, string(domain.KEYGEN_MSG_TYPE), req.Version, req.BlockHeight, req.Keys, len(req.Keys)-1, sigChan, []byte(req.Keys[0]))

	if errJoinParty != nil {
		if leader == "NONE" {
			if onlinePeers == nil {
				t.logger.Error().Err(err).Msg("error before we start join party")
				return keygen.Response{
					Status: common.Fail,
					Blame:  blame.NewBlame(blame.InternalError, []blame.Node{}),
				}, errJoinParty
			}
			blameNodes, err := blameMgr.NodeSyncBlame(req.Keys, onlinePeers)
			if err != nil {
				t.logger.Err(errJoinParty).Msg("fail to get peers to blame")
			}
			// make sure we blame the leader as well
			t.logger.Error().Err(errJoinParty).Msgf("fail to form keygen party with online:%v", onlinePeers)
			return keygen.Response{
				Status: common.Fail,
				Blame:  blameNodes,
			}, errJoinParty

		}

		var blameLeader blame.Blame
		var blameNodes blame.Blame
		blameNodes, err = blameMgr.NodeSyncBlame(req.Keys, onlinePeers)
		if err != nil {
			t.logger.Err(errJoinParty).Msg("fail to get peers to blame")
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

		return keygen.Response{
			Status: common.Fail,
			Blame:  blameNodes,
		}, errJoinParty
	}

	tssCommonStruct := keygenInstance.GetTssCommonStruct()
	if err := tssCommonStruct.SetPeers(onlinePeers); err != nil {
		t.logger.Error().Err(err).Msgf("fail to set online peers for keygen:%v", onlinePeers)
		blameNodes := *blameMgr.GetBlame()
		return keygen.Response{
			Status: common.Fail,
			Blame:  blameNodes,
		}, errJoinParty
	}

	t.logger.Debug().Msg("keygen party formed")
	// the statistic of keygen only care about Tss it self, even if the
	// following http response aborts, it still counted as a successful keygen
	// as the Tss model runs successfully.
	beforeKeygen := time.Now()
	k1, err := keygenInstance.GenerateNewKey_Ecdsa(req)

	keygenTime := time.Since(beforeKeygen)
	if err != nil {
		t.logger.Error().Err(err).Msg("err in keygen")
		blameNodes := *blameMgr.GetBlame()
		return keygen.NewResponse("", "", common.Fail, blameNodes), err
	} else {
		t.logger.Error().Err(err).Msg(fmt.Sprintf("err in keygen at time: %v, %v", keygenTime, err))
	}

	pubPoolKey, err := conversion.GetTssPubKey(k1)
	if err != nil {
		t.logger.Error().Err(err).Msg("fail to generate the new Tss key")
		status = common.Fail
		blameNodes := *blameMgr.GetBlame()
		return keygen.NewResponse("", "", common.Fail, blameNodes), err
	}

	blameNodes := *blameMgr.GetBlame()
	return keygen.NewResponse(
		pubPoolKey,
		pubPoolKey,
		status,
		blameNodes,
	), nil
}
