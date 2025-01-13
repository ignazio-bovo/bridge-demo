package observer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	"github.com/Datura-ai/relayer/packages/domain/observer"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EventListener interface {
	Start() error
}

// Observer struct with injected dependencies
type WebsocketEventListener struct {
	ethClient           *ethclient.Client
	parser              EventParser
	messageBus          *EventBus
	blockConfirmations  uint64
	lastBlockProcessed  uint64
	lastBlockMutex      sync.RWMutex
	blockProductionTime time.Duration
	address             common.Address
	chainId             uint64
	logger              *zerolog.Logger
	blockBatchSize      uint64
}

// NewObserver constructor with dependency injection
func NewWebsocketEventListener(
	network observer.Network,
	ethClient *ethclient.Client,
	messageBus *EventBus,
	parser EventParser,
) (*WebsocketEventListener, error) {
	logger := log.Logger.With().Str("Network", network.Name).Str("Module", "👂 Listener").Logger()

	return &WebsocketEventListener{
		ethClient:           ethClient,
		messageBus:          messageBus,
		parser:              parser,
		blockConfirmations:  uint64(network.BlockConfirmations),
		address:             network.ContractAddress,
		blockProductionTime: network.BlockProductionTime,
		lastBlockProcessed:  network.StartBlock, // optional, will be loaded from state file
		chainId:             network.ChainId,
		logger:              &logger,
		blockBatchSize:      100,
	}, nil
}

// Start method with error return for better testability
func (l *WebsocketEventListener) Start() error {
	defer l.ethClient.Close()

	// Create filter query for the contract
	query := ethereum.FilterQuery{
		Addresses: []common.Address{l.address},
		Topics: [][]common.Hash{
			{domain_common.GetTransferRequestedSignature()},
		},
	}

	for {
		chainTipBlock, err := l.ethClient.BlockNumber(context.Background())
		if err != nil {
			l.logger.Error().Err(err).Msg("Failed to get latest block number")
			return err
		}

		headBlock := uint64(0)
		if chainTipBlock > l.blockConfirmations {
			headBlock = chainTipBlock - l.blockConfirmations
		}

		if headBlock > l.lastBlockProcessed {
			l.lastBlockMutex.RLock()
			toBlock := min(headBlock, l.lastBlockProcessed+l.blockBatchSize)
			fromBlock := big.NewInt(int64(l.lastBlockProcessed + 1)) // +1 because we already processed the last block in the previous iteration
			l.lastBlockMutex.RUnlock()

			toBlockBig := big.NewInt(int64(toBlock))
			query.FromBlock = fromBlock
			query.ToBlock = toBlockBig

			// Get logs for the block range
			l.logger.Debug().Msgf("Fetching logs from block %d to %d with chainTipBlock %d", fromBlock, toBlock, chainTipBlock)
			logs, err := l.ethClient.FilterLogs(context.Background(), query)
			if err != nil {
				l.logger.Error().Err(err).Msg("Failed to filter logs")
				return err
			}

			// Process each log
			for _, vLog := range logs {
				routingMessage, err := l.parser.ParseEvent(vLog.Topics, vLog.Data)
				if err != nil {
					l.logger.Error().Err(err).Msg("Error parsing event")
					return err
				}
				l.logger.Debug().Msgf("Received routing message: %+v", routingMessage)
				l.messageBus.Route(routingMessage)
			}

			// Update last block processed with mutex protection
			l.lastBlockMutex.Lock()
			l.lastBlockProcessed = toBlock
			l.lastBlockMutex.Unlock()

			l.saveState()
		}

		time.Sleep(l.blockProductionTime)
	}
}

func (l *WebsocketEventListener) saveState() error {
	// TODO: make this a constant not dependeing on the execution dir
	stateFile := fmt.Sprintf("./config/listenerState_%d.json", l.chainId)
	state := struct {
		LastBlockProcessed uint64 `json:"lastBlockProcessed"`
		Timestamp          uint64 `json:"timestamp"`
	}{LastBlockProcessed: l.lastBlockProcessed, Timestamp: uint64(time.Now().Unix())}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}
