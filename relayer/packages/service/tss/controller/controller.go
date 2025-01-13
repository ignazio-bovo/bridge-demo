package controller

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/Datura-ai/relayer/packages/domain/common"
	tssdomain "github.com/Datura-ai/relayer/packages/domain/tss"
	lrepository "github.com/Datura-ai/relayer/packages/repository/listener"
	trepository "github.com/Datura-ai/relayer/packages/repository/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/tss"
	geth_comm "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
)

type TssController struct {
	cfg              *tssdomain.Configuration
	logger           zerolog.Logger
	msgsLock         sync.Mutex
	keygen           *Keygen
	signer           *KeySign
	resharing        *Keyresharing
	stopChan         chan struct{}
	tssStateMgr      trepository.TssStateManager
	listenerStateMgr lrepository.ListenerStateManager
}

func NewTssController(tssConfig *tssdomain.Configuration, logger zerolog.Logger, tssServer *tss.TssServer, tssStateMgr trepository.TssStateManager, listenerStateMgr lrepository.ListenerStateManager) (*TssController, error) {
	keygen, err := NewTssKeygen(tssServer)
	if err != nil {
		return nil, err
	}
	signer, err := NewKeySign(tssServer)
	if err != nil {
		return nil, err
	}
	return &TssController{
		cfg:              tssConfig,
		logger:           logger,
		msgsLock:         sync.Mutex{},
		keygen:           keygen,
		signer:           signer,
		stopChan:         make(chan struct{}),
		tssStateMgr:      tssStateMgr,
		listenerStateMgr: listenerStateMgr,
	}, nil
}

func (c *TssController) Start() error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go c.startPeerDiscovery()
	wg.Add(2)
	go c.processKeygenRequest(c.keygen.server.GetNewPeerChannel())
	wg.Add(3)
	go c.signer.Start()
	wg.Add(4)
	go c.processKeysignRequest(c.keygen.server.GetKeysignChan())

	// It will be removed after testing
	wg.Add(5)
	go c.GenerateMockKeysignRequest()
	wg.Wait()
	return nil
}

func (c *TssController) startPeerDiscovery() {
	err := c.keygen.server.StartPeerDiscovery()
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to start peer discovery")
	}
	c.logger.Info().Msg("Waiting for other parties to join...")
}

func (c *TssController) processKeygenRequest(peerChan *chan []peer.ID) error {
	for {
		peerChan := c.keygen.server.GetNewPeerChannel()
		if peerChan == nil {
			continue
		}
		select {
		case peers, ok := <-*peerChan:
			if !ok {
				c.logger.Warn().Msg("Peer channel closed unexpectedly")
				return fmt.Errorf("peer channel closed")
			}
			if len(peers) == 0 || len(peers) == 1 {
				c.logger.Warn().Msg("Received empty list")
				continue
			}
			c.logger.Info().Msgf("Received peers: %v", len(peers))
			localPeers, err := c.tssStateMgr.RetrievePeers()
			if err != nil {
				return fmt.Errorf("failed to retrieve local peers: %w", err)
			}
			_, err = c.tssStateMgr.GetTssPubKey()
			if slicesEqual(peers, localPeers) && err == nil {
				continue
			}
			if err := c.tssStateMgr.UpdatePeers(peers); err != nil {
				return fmt.Errorf("failed to update peers: %w", err)
			}
			peerPubKeys := make([]string, len(peers))
			for i, peer := range peers {
				peerPubKeys[i] = peer.String()
			}
			poolPubKey, err := c.keygen.GenerateNewKey(peerPubKeys)
			if err != nil {
				return fmt.Errorf("failed to generate new key: %w", err)
			}
			c.logger.Info().Msgf("Generated new key: %s", poolPubKey)
			c.tssStateMgr.SaveTssPubKey(poolPubKey)
			return nil
		case <-c.stopChan:
			c.logger.Info().Msg("Received stop signal")
			return nil
		}
	}
}

func (c *TssController) processKeysignRequest(keysignChan *chan common.BridgeTxData) error {
	pubKey, err := c.tssStateMgr.GetTssPubKey()
	if err != nil {
		return fmt.Errorf("failed to get tss pubkey: %w", err)
	}
	for {
		select {
		case msg, ok := <-*keysignChan:
			if !ok {
				c.logger.Warn().Msg("Keysign channel closed unexpectedly")
				return fmt.Errorf("keysign channel closed")
			}
			msgId := msg.GetMsgId()
			//stateMgr.AddPendingMsgID(msgId)
			data, bRecoveryId, resmsg, err := c.signer.RemoteSign(msg.BuildReleaseTx(), pubKey, msgId, false)
			if err != nil {
				return fmt.Errorf("failed to remote sign: %w", err)
			}
			//stateMgr.RemovePendingMsgID(msgId)
			c.logger.Info().Msgf("Received message: %v", msg)
			c.logger.Info().Msgf("data: %v", data)
			c.logger.Info().Msgf("bRecoveryId: %v", bRecoveryId)
			c.logger.Info().Msgf("resmsg: %v", resmsg)
			c.logger.Info().Msgf("err: %v", err)

			// Create EVM client and broadcast the tx
			rpcHost, err := c.cfg.GetRPCHost(fmt.Sprintf("%d", msg.SrcChainId))
			if err != nil {
				return fmt.Errorf("failed to get rpc host: %w", err)
			}
			evmClient, err := NewEVMClient(rpcHost, &msg)
			if err != nil {
				return fmt.Errorf("failed to create evm client: %w", err)
			}
			txID, err := evmClient.BroadcastTx(data)
			if err != nil {
				return fmt.Errorf("failed to broadcast tx: %w", err)
			}
			c.logger.Info().Msgf("txID: %s", txID)

			// Only leader node will broadcast the tx
			isLeader, err := c.signer.server.IsLeader(msgId, 0)
			if err != nil {
				return fmt.Errorf("failed to check if leader: %w", err)
			}
			if isLeader {
				c.logger.Info().Msg("I am the leader, broadcast the tx")
				c.listenerStateMgr.SaveBroadcastMsg(msg)
			}
		case <-c.stopChan:
			c.logger.Info().Msg("Received stop signal")
			return nil
		}
	}
}

func slicesEqual(a, b []peer.ID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// MockKeysignRequest is used for testing only
func (c *TssController) GenerateMockKeysignRequest() {
	time.Sleep(20 * time.Second)
	mockTxMsg := c.mockTxMsg()
	c.listenerStateMgr.SaveBroadcastMsg(common.BridgeTxData{})
	keysignChan := c.signer.server.GetKeysignChan()
	*keysignChan <- mockTxMsg
}
func (c *TssController) mockTxMsg() common.BridgeTxData {
	return common.BridgeTxData{
		SrcChainId: 2332,
		To:         geth_comm.HexToAddress("0x1234567890123456789012345678901234567890"),
		TokenKey:   geth_comm.HexToHash("0x1234567890123456789012345678901234567890"),
		Amount:     big.NewInt(1000000000000000000),
		Nonce:      big.NewInt(1),
		TokenMetadata: common.TokenMetadata{
			Name:     "Mock Token",
			Symbol:   "MOCK",
			Decimals: 18,
		},
	}
}
