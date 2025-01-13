package controller

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/Datura-ai/relayer/packages/service/tss/keysign"
	"github.com/Datura-ai/relayer/packages/service/tss/messages"
	"github.com/Datura-ai/relayer/packages/service/tss/tss"
	"github.com/blang/semver"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	maxKeysignPerRequest = 1 // the maximum number of messages include in one single TSS keysign request
	tssKeysignTimeout    = 5 // in minutes, the maximum time tss-proc is going to wait before the tss result come back
)

// KeySign is a proxy between signer and TSS
type KeySign struct {
	logger         zerolog.Logger
	server         *tss.TssServer
	currentVersion semver.Version
	lastCheck      time.Time

	wg        *sync.WaitGroup
	taskQueue chan *tssKeySignTask
	done      chan struct{}
	handleID  uint64
}

// NewKeySign create a new instance of KeySign
func NewKeySign(server *tss.TssServer) (*KeySign, error) {
	return &KeySign{
		server:    server,
		logger:    log.With().Str("module", "tss_signer").Logger(),
		wg:        &sync.WaitGroup{},
		taskQueue: make(chan *tssKeySignTask),
		done:      make(chan struct{}),
		handleID:  1,
	}, nil
}

// GetPrivKey KIMANode don't actually have any private key , but just return something
func (s *KeySign) GetPrivKey() crypto.PrivKey {
	return nil
}

// ExportAsMnemonic KIMANode don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsMnemonic() (string, error) {
	return "", nil
}

// ExportAsPrivateKey KIMANode don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsPrivateKey() (string, error) {
	return "", nil
}

// Start the keysign workers
func (s *KeySign) Start() {
	s.wg.Add(1)
	go s.processKeySignTasks()
}

// Stop Keysign
func (s *KeySign) Stop() {
	close(s.done)
	s.wg.Wait()
	close(s.taskQueue)

	// Stop p2p party connector
	s.server.StopPartyConnector()
}

// RemoteSign send the request to local task queue
// lastTssInput is the param to tell tss signer that he can close p2p channels or not.
// this will be true by default for normal transaction, but it can only be true for the last input of BTC transaction while former inputs will have false.
// purpose is to use the same msgId and handlerId to sign each inputs of a single BTC transaction.
func (s *KeySign) RemoteSign(msg []byte, poolPubKey string, msgId string, lastTssInput bool) ([]byte, []byte, string, error) {
	if len(msg) == 0 {
		return nil, nil, "", nil
	}

	encodedMsg := base64.StdEncoding.EncodeToString(msg)
	task := tssKeySignTask{
		PoolPubKey:   poolPubKey,
		Msg:          encodedMsg,
		MsgId:        msgId,
		Resp:         make(chan tssKeySignResult, 1),
		LastTssInput: lastTssInput,
	}
	s.taskQueue <- &task
	select {
	case resp := <-task.Resp:
		if resp.Err != nil {
			return nil, nil, "", fmt.Errorf("fail to tss sign: %w", resp.Err)
		}

		if len(resp.R) == 0 && len(resp.S) == 0 {
			// this means the node tried to do keysign , however this node has not been chosen to take part in the keysign committee
			return nil, nil, "", nil
		}
		//s.logger.Debug().Str("R", resp.R).Str("S", resp.S).Str("recovery", resp.RecoveryID).Msg("tss result")
		data, err := getSignature(resp.R, resp.S)
		if err != nil {
			return nil, nil, "", fmt.Errorf("fail to decode tss signature: %w", err)
		}
		bRecoveryId, err := base64.StdEncoding.DecodeString(resp.RecoveryID)
		if err != nil {
			return nil, nil, "", fmt.Errorf("fail to decode recovery id: %w", err)
		}
		return data, bRecoveryId, resp.Msg, nil
	case <-time.After(time.Minute * tssKeysignTimeout):
		return nil, nil, "", fmt.Errorf("TIMEOUT: fail to sign message:%s after %d minutes", encodedMsg, tssKeysignTimeout)
	}
}

// astTssInput is the param to tell tss signer that he can close p2p channels or not.
// this will be true by default for normal transaction, but it can only be true for the last input of BTC transaction while former inputs will have false.
// purpose is to use the same msgId and handlerId to sign each inputs of a single BTC transaction.
type tssKeySignTask struct {
	PoolPubKey   string
	Msg          string
	MsgId        string
	Resp         chan tssKeySignResult
	LastTssInput bool // the param to tell tss signer that he can close p2p channels or not.
}

type tssKeySignResult struct {
	R          string
	S          string
	RecoveryID string
	Err        error
	Msg        string
}

func (s *KeySign) processKeySignTasks() {
	defer s.wg.Done()
	tasks := make(map[string][]*tssKeySignTask)
	taskLock := sync.Mutex{}
	for {
		select {
		case <-s.done:
			// requested to exit
			return
		case t := <-s.taskQueue:
			s.logger.Info().Msg("Entered Task Queue")

			taskLock.Lock()
			_, ok := tasks[t.PoolPubKey]
			if !ok {
				s.logger.Warn().Msg("Cant fetch task by poolpubkey!")
				tasks[t.PoolPubKey] = []*tssKeySignTask{
					t,
				}
			} else {
				s.logger.Info().Msg("Append this task to task queue")
				tasks[t.PoolPubKey] = append(tasks[t.PoolPubKey], t)
			}
			taskLock.Unlock()
		case <-time.After(time.Second):
			// This implementation will check the tasks every second , and send whatever is in the queue to TSS
			// if it has more than maxKeysignPerRequest(1) in the queue , it will only send the first maxKeysignPerRequest(1) of them
			// the reset will be send in the next request

			taskLock.Lock()
			for k, v := range tasks {
				if len(v) == 0 {
					delete(tasks, k)
					continue
				}
				totalTasks := len(v)
				// send no more than maxKeysignPerRequest messages in a single TSS keysign request
				if totalTasks > maxKeysignPerRequest {
					totalTasks = maxKeysignPerRequest
					// when there are more than maxKeysignPerRequest messages in the task queue need to be signed
					// the messages has to be sorted , because the order of messages that get into the slice is not deterministic
					// so it need to sorted to make sure all tss-procs send the same messages to tss
					sort.SliceStable(v, func(i, j int) bool {
						return v[i].Msg < v[j].Msg
					})
				}
				s.wg.Add(1)
				signingTask := v[:totalTasks]
				tasks[k] = v[totalTasks:]

				s.logger.Info().Msg("Before entering LocalTSS Signer")
				go s.toLocalTSSSigner(k, signingTask)
			}
			taskLock.Unlock()
		}
	}
}

func getSignature(r, s string) ([]byte, error) {
	rBytes, err := base64.StdEncoding.DecodeString(r)
	if err != nil {
		return nil, err
	}
	sBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	R := new(big.Int).SetBytes(rBytes)
	S := new(big.Int).SetBytes(sBytes)
	N := btcec.S256().N
	halfOrder := new(big.Int).Rsh(N, 1)
	// see: https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
	if S.Cmp(halfOrder) == 1 {
		S.Sub(N, S)
	}

	// Serialize signature to R || S.
	// R, S are padded to 32 bytes respectively.
	rBytes = R.Bytes()
	sBytes = S.Bytes()

	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)
	return sigBytes, nil
}

func (s *KeySign) getVersion() semver.Version {
	requestTime := time.Now()
	s.lastCheck = requestTime
	s.currentVersion, _ = semver.Make(messages.NEWJOINPARTYVERSION)

	return s.currentVersion
}

func (s *KeySign) setTssKeySignTasksFail(tasks []*tssKeySignTask, err error) {
	for _, item := range tasks {
		select {
		case item.Resp <- tssKeySignResult{
			Err: err,
		}:
		case <-time.After(time.Second):
			// this is a fallback , if fail to send a failed result back to caller , it doesn't stuck
			continue
		}
	}
}

// toLocalTSSSigner will send the request to local signer
func (s *KeySign) toLocalTSSSigner(poolPubKey string, tasks []*tssKeySignTask) {
	s.logger.Info().Msg("Entered local tss singer")

	defer s.wg.Done()
	var msgToSign []string
	for _, item := range tasks {
		msgToSign = append(msgToSign, item.Msg)
	}

	pubKeys, err := s.server.GetStateManager().GetSignerPubKeys()
	if err != nil {
		s.logger.Error().Err(err).Msg("could not fetch public keys")
		return
	}

	// Check if valid data
	if len(pubKeys) < 1 {
		return
	}

	tssMsg := keysign.Request{
		PoolPubKey:    poolPubKey,
		Messages:      msgToSign,
		SignerPubKeys: pubKeys,
	}

	currentVersion := s.getVersion()
	tssMsg.Version = currentVersion.String()
	s.logger.Debug().Msg("new TSS join party, version:" + tssMsg.Version)

	tssMsg.BlockHeight = (int64)(1)

	s.logger.Info().Msgf("msgToSign to tss Local node PoolPubKey: %s, Messages: %+v, block height: %d", tssMsg.PoolPubKey, tssMsg.Messages, tssMsg.BlockHeight)
	s.logger.Info().Msg("Before entering TSS server keysign")

	// If there is no tx to process
	if len(tasks) < 1 {
		return
	}

	// Fetch TSS Msg ID
	// TODO: Need to make sure we handle one task at a time.
	msgID := tasks[0].MsgId
	lastTssInput := tasks[0].LastTssInput

	keySignResp, err := s.server.KeySign(tssMsg, msgID, s.handleID, lastTssInput)
	if err != nil {
		s.setTssKeySignTasksFail(tasks, fmt.Errorf("fail tss keysign: %w", err))
		return
	}

	s.logger.Info().Msg("Get result from TSS server keysign")

	// 1 means success,2 means fail , 0 means NA
	if keySignResp.Status == 1 && len(keySignResp.Blame.BlameNodes) == 0 {
		s.logger.Info().Msgf("response: %+v", keySignResp)
		// success
		for _, t := range tasks {
			found := false
			for _, sig := range keySignResp.Signatures {
				// Commmented this line for tron as it uses the leader's hash instead of its local hash
				// Because signature won't be matches with the local task content.
				s.logger.Info().Msg("Keysign result")

				t.Resp <- tssKeySignResult{
					R:          sig.R,
					S:          sig.S,
					RecoveryID: sig.RecoveryID,
					Err:        nil,
					Msg:        sig.Msg,
				}
				found = true
			}
			// Didn't find the signature in the tss keysign result , notify the task , so it doesn't get stuck
			if !found {
				s.logger.Error().Msg("didn't find signature for message")

				t.Resp <- tssKeySignResult{
					Err: fmt.Errorf("didn't find signature for message %s in the keysign result", t.Msg),
				}
			}
		}
		return
	}

	// Blame need to be passed back to Kimachain , so as Kimachain can use the information to slash relevant node account
	errMsg := "TSS keysign failed"
	if keySignResp.Status == 2 {
		errMsg = "TSS keysign explicitly failed"
	} else if len(keySignResp.Blame.BlameNodes) > 0 {
		errMsg = fmt.Sprintf("TSS keysign failed with blame: %v", keySignResp.Blame.BlameNodes)
	}
	s.logger.Error().Msg(errMsg)
	s.setTssKeySignTasksFail(tasks, fmt.Errorf("fail tss keysign: %w", errMsg))
}
