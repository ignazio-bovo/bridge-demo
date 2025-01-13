package tss

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	tsslibcommon "github.com/bnb-chain/tss-lib/common"
	"github.com/libp2p/go-libp2p/core/peer"

	domain "github.com/Datura-ai/relayer/packages/domain/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/blame"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
	"github.com/Datura-ai/relayer/packages/service/tss/conversion"
	"github.com/Datura-ai/relayer/packages/service/tss/keysign"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	"github.com/Datura-ai/relayer/packages/service/tss/storage"
)

func (t *TssServer) waitForSignatures(msgID, poolPubKey string, msgsToSign [][]byte, sigChan chan string) (keysign.Response, error) {
	// TSS keysign include both form party and keysign itself, thus we wait twice of the timeout
	data, err := t.signatureNotifier.WaitForSignature(msgID, msgsToSign, poolPubKey, t.conf.KeysignTimeout, sigChan)
	if err != nil {
		return keysign.Response{}, err
	}
	// for gg20, it wrap the signature R,S into SignatureData structure
	if len(data) == 0 {
		return keysign.Response{}, errors.New("keysign failed")
	}

	return t.batchSignatures(data, msgsToSign), nil
}

func (t *TssServer) generateSignature(msgID string, msgsToSign [][]byte, req keysign.Request, threshold int, allParticipants []string, localStateItem storage.KeygenLocalState, blameMgr *blame.Manager, keysignInstance *keysign.TssKeysign, sigChan chan string) (keysign.Response, error) {
	allPeersID, err := conversion.GetPeerIDsFromPubKeys(allParticipants)
	if err != nil {
		t.logger.Error().Msg("invalid block height or public key")
		return keysign.Response{
			Status: common.Fail,
			Blame:  blame.NewBlame(blame.InternalError, []blame.Node{}),
		}, err
	}

	oldJoinParty, err := conversion.VersionLTCheck(req.Version, messages.NEWJOINPARTYVERSION)
	if err != nil {
		return keysign.Response{
			Status: common.Fail,
			Blame:  blame.NewBlame(blame.InternalError, []blame.Node{}),
		}, errors.New("fail to parse the version")
	}
	// we use the old join party
	if oldJoinParty {
		allParticipants = req.SignerPubKeys
		myPk, err := conversion.GetPubKeyFromPeerID(t.p2pCommunication.GetHost().ID().String())
		if err != nil {
			t.logger.Info().Msgf("fail to convert the p2p id(%s) to pubkey, turn to wait for signature", t.p2pCommunication.GetHost().ID().String())
			return keysign.Response{}, p2p.ErrNotActiveSigner
		}
		isSignMember := false
		for _, el := range allParticipants {
			if myPk == el {
				isSignMember = true
				break
			}
		}
		if !isSignMember {
			t.logger.Info().Msgf("we(%s) are not the active signer", t.p2pCommunication.GetHost().ID().String())
			return keysign.Response{}, p2p.ErrNotActiveSigner
		}
	}

	onlinePeers, leader, msgToSign, errJoinParty := t.joinParty(msgID, string(domain.KEYSIGN_MSG_TYPE), req.Version, req.BlockHeight, allParticipants, threshold, sigChan, msgsToSign[0])
	for _, op := range onlinePeers {
		t.logger.Info().Msgf("online peers: (%s)", op.String())
	}

	t.logger.Info().Msgf("leader: (%s)", leader)

	if errJoinParty != nil {
		// we received the signature from waiting for signature
		if errors.Is(errJoinParty, p2p.ErrSignReceived) {
			return keysign.Response{}, errJoinParty
		}

		// this indicate we are processing the leaderness join party
		if leader == "NONE" {
			if onlinePeers == nil {
				t.logger.Error().Err(errJoinParty).Msg("error before we start join party")
				t.broadcastKeysignFailure(msgID, allPeersID)
				return keysign.Response{
					Status: common.Fail,
					Blame:  blame.NewBlame(blame.InternalError, []blame.Node{}),
				}, errJoinParty
			}

			blameNodes, err := blameMgr.NodeSyncBlame(req.SignerPubKeys, onlinePeers)
			if err != nil {
				t.logger.Err(err).Msg("fail to get peers to blame")
			}
			t.broadcastKeysignFailure(msgID, allPeersID)
			// make sure we blame the leader as well
			t.logger.Error().Err(err).Msgf("fail to form keysign party with online:%v", onlinePeers)
			return keysign.Response{
				Status: common.Fail,
				Blame:  blameNodes,
			}, errJoinParty
		}

		var blameLeader blame.Blame
		leaderPubKey, err := conversion.GetPubKeyFromPeerID(leader)
		if err != nil {
			t.logger.Error().Err(errJoinParty).Msgf("fail to convert the peerID to public key %s", leader)
			blameLeader = blame.NewBlame(blame.TssSyncFail, []blame.Node{})
		} else {
			blameLeader = blame.NewBlame(blame.TssSyncFail, []blame.Node{{Pubkey: leaderPubKey, BlameData: nil, BlameSignature: nil}})
		}

		t.broadcastKeysignFailure(msgID, allPeersID)
		// make sure we blame the leader as well
		t.logger.Error().Err(errJoinParty).Msgf("messagesID(%s)fail to form keysign party with online:%v", msgID, onlinePeers)
		return keysign.Response{
			Status: common.Fail,
			Blame:  blameLeader,
		}, errJoinParty

	}

	isKeySignMember := false
	for _, el := range onlinePeers {
		if el == t.p2pCommunication.GetHost().ID() {
			isKeySignMember = true
		}
	}
	if !isKeySignMember {
		// we are not the keysign member so we quit keysign and waiting for signature
		t.logger.Info().Msgf("we(%s) are not the active signer", t.p2pCommunication.GetHost().ID().String())
		return keysign.Response{}, p2p.ErrNotActiveSigner
	}
	parsedPeers := make([]string, len(onlinePeers))
	for i, el := range onlinePeers {
		parsedPeers[i] = el.String()
	}

	signers, err := conversion.GetPubKeysFromPeerIDs(parsedPeers)
	if err != nil {
		sigChan <- "signature generated"
		return keysign.Response{
			Status: common.Fail,
			Blame:  blame.Blame{},
		}, err
	}

	// clear and use the common one that is shared by tss p2p.
	msgsToSign = make([][]byte, 0)
	msgsToSign = append(msgsToSign, msgToSign)
	signatureData, err := keysignInstance.SignMessage(msgsToSign, localStateItem, signers, msgID)
	// the statistic of keygen only care about Tss it self, even if the following http response aborts,
	// it still counted as a successful keygen as the Tss model runs successfully.
	if err != nil {
		t.logger.Error().Err(err).Msg("err in keysign")
		sigChan <- "signature generated"
		t.broadcastKeysignFailure(msgID, allPeersID)
		blameNodes := *blameMgr.GetBlame()
		return keysign.Response{
			Status: common.Fail,
			Blame:  blameNodes,
		}, err
	}

	sigChan <- "signature generated"
	// update signature notification
	if err := t.signatureNotifier.BroadcastSignature(msgID, signatureData, allPeersID); err != nil {
		return keysign.Response{}, fmt.Errorf("fail to broadcast signature:%w", err)
	}

	return t.batchSignatures(signatureData, msgsToSign), nil
}

func (t *TssServer) updateKeySignResult(result keysign.Response, timeSpent time.Duration) {
	if result.Status == common.Success {
		return
	}
	t.logger.Error().Msgf("keysign failed with error: %v", result.Blame)
}

// Process keysigning through tss library
// Params:
// msgId : Used in channel subscribe to process p2p streams
// handleId: Used in p2p party construction to make p2p connection
// lastTssInput: Used in BTC transaction mainly to preserve p2p connection while signing the multiple inputs for the single transaction.
// it is mostly true to ensure closing all channels and streams but set false in the middle of BTC multiple inputs signing.
// for the last input of BTC multple inputs, it will use true to close stream.
func (t *TssServer) KeySign(req keysign.Request, msgID string, handleID uint64, lastTssInput bool) (keysign.Response, error) {
	t.logger.Info().Str("pool pub key", req.PoolPubKey).
		Str("signer pub keys", strings.Join(req.SignerPubKeys, ",")).
		Str("msg", strings.Join(req.Messages, ",")).
		Msg("received keysign request")
	emptyResp := keysign.Response{}

	keysignInstance := keysign.NewTssKeysign(
		t.localNodePubKey,
		t.stopChan,
		msgID,
		t.p2pCommunication,
		t.stateManager,
		t.privateKey,
		t.conf,
		handleID,
	)

	if keysignInstance == nil {
		return keysign.Response{
			Status: common.Fail,
			Blame:  blame.NewBlame(blame.InternalError, []blame.Node{}),
		}, errors.New("keysign instance fail")
	}

	keySignChannels := keysignInstance.GetTssKeySignChannels()

	t.p2pCommunication.SetSubscribe(messages.TSSKeysignMsg, msgID, keySignChannels)
	t.p2pCommunication.SetSubscribe(messages.TSSKeysignVerMsg, msgID, keySignChannels)
	t.p2pCommunication.SetSubscribe(messages.TSSControlMsg, msgID, keySignChannels)
	t.p2pCommunication.SetSubscribe(messages.TSSTaskDone, msgID, keySignChannels)

	defer func() {
		t.p2pCommunication.CancelSubscribe(messages.TSSKeysignMsg, msgID)
		t.p2pCommunication.CancelSubscribe(messages.TSSKeysignVerMsg, msgID)
		t.p2pCommunication.CancelSubscribe(messages.TSSControlMsg, msgID)
		t.p2pCommunication.CancelSubscribe(messages.TSSTaskDone, msgID)

		if lastTssInput {
			t.p2pCommunication.ReleaseStream(msgID)
			t.signatureNotifier.ReleaseStream(msgID)
			t.partyCoordinator.ReleaseStream(msgID)
		}
	}()

	localStateItem, err := t.stateManager.GetLocalState(req.PoolPubKey)
	if err != nil {
		return keysign.Response{}, fmt.Errorf("fail to get local keygen state: %w", err)
	}
	localPubKey, err := t.tssStateManager.ID()
	if err != nil {
		return keysign.Response{}, fmt.Errorf("fail to get local keygen state: %w", err)
	}

	peers, err := t.tssStateManager.GetSignerPubKeys()
	if err != nil {
		return keysign.Response{}, fmt.Errorf("fail to get signer pub keys: %w", err)
	}
	// replace local party key with the local peerId
	localStateItem.LocalPartyKey = localPubKey.String()
	localStateItem.ParticipantKeys = peers

	var msgsToSign [][]byte
	for _, val := range req.Messages {
		msgToSign, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return keysign.Response{}, fmt.Errorf("fail to decode message(%s): %w", strings.Join(req.Messages, ","), err)
		}
		msgsToSign = append(msgsToSign, msgToSign)
	}

	sort.SliceStable(msgsToSign, func(i, j int) bool {
		ma, err := common.MsgToHashInt(msgsToSign[i])
		if err != nil {
			t.logger.Error().Err(err).Msgf("fail to convert the hash value")
		}
		mb, err := common.MsgToHashInt(msgsToSign[j])
		if err != nil {
			t.logger.Error().Err(err).Msgf("fail to convert the hash value")
		}
		if ma.Cmp(mb) == -1 {
			return false
		}
		return true
	})

	oldJoinParty, err := conversion.VersionLTCheck(req.Version, messages.NEWJOINPARTYVERSION)
	if err != nil {
		return keysign.Response{
			Status: common.Fail,
			Blame:  blame.NewBlame(blame.InternalError, []blame.Node{}),
		}, errors.New("fail to parse the version")
	}

	if len(req.SignerPubKeys) == 0 && oldJoinParty {
		return emptyResp, errors.New("empty signer pub keys")
	}

	threshold, err := conversion.GetThreshold(len(localStateItem.ParticipantKeys))
	if err != nil {
		t.logger.Error().Err(err).Msg("fail to get the threshold")
		return emptyResp, errors.New("fail to get threshold")
	}
	if len(req.SignerPubKeys) <= threshold && oldJoinParty {
		t.logger.Error().Msgf("not enough signers, threshold=%d and signers=%d", threshold, len(req.SignerPubKeys))
		return emptyResp, errors.New("not enough signers")
	}

	blameMgr := keysignInstance.GetTssCommonStruct().GetBlameMgr()

	var receivedSig, generatedSig keysign.Response
	var errWait, errGen error
	sigChan := make(chan string, 2)
	wg := sync.WaitGroup{}
	wg.Add(2)
	keysignStartTime := time.Now()
	// we wait for signatures
	go func() {
		defer wg.Done()
		receivedSig, errWait = t.waitForSignatures(msgID, req.PoolPubKey, msgsToSign, sigChan)
		// we received an valid signature indeed
		if errWait == nil {
			sigChan <- "signature received"
			t.logger.Log().Msgf("for message %s we get the signature from the peer", msgID)
			return
		}
		t.logger.Log().Msgf("we fail to get the valid signature with error %v", errWait)
	}()

	// we generate the signature ourselves
	go func() {
		defer wg.Done()
		generatedSig, errGen = t.generateSignature(msgID, msgsToSign, req, threshold, localStateItem.ParticipantKeys, localStateItem, blameMgr, keysignInstance, sigChan)
	}()
	wg.Wait()
	close(sigChan)
	keysignTime := time.Since(keysignStartTime)
	// we received the generated verified signature, so we return
	if errWait == nil {
		t.updateKeySignResult(receivedSig, keysignTime)
		return receivedSig, nil
	}
	// for this round, we are not the active signer
	if errors.Is(errGen, p2p.ErrSignReceived) || errors.Is(errGen, p2p.ErrNotActiveSigner) {
		t.updateKeySignResult(receivedSig, keysignTime)
		return receivedSig, nil
	}
	// we get the signature from our tss keysign
	t.updateKeySignResult(generatedSig, keysignTime)
	return generatedSig, errGen
}

func (t *TssServer) broadcastKeysignFailure(messageID string, peers []peer.ID) {
	if err := t.signatureNotifier.BroadcastFailed(messageID, peers); err != nil {
		t.logger.Err(err).Msg("fail to broadcast keysign failure")
	}
}

func (t *TssServer) batchSignatures(sigs []*tsslibcommon.SignatureData, msgsToSign [][]byte) keysign.Response {
	var signatures []keysign.Signature
	for i, sig := range sigs {
		msg := base64.StdEncoding.EncodeToString(msgsToSign[i])
		r := base64.StdEncoding.EncodeToString(sig.R)
		s := base64.StdEncoding.EncodeToString(sig.S)
		recovery := base64.StdEncoding.EncodeToString(sig.SignatureRecovery)

		signature := keysign.NewSignature(msg, r, s, recovery)
		signatures = append(signatures, signature)
	}
	return keysign.NewResponse(
		signatures,
		common.Success,
		blame.Blame{},
	)
}
