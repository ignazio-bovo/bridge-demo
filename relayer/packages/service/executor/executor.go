package executor

import (
	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	"github.com/Datura-ai/relayer/packages/domain/observer"
	"github.com/rs/zerolog/log"
)

type Executor interface {
	Execute(transferRequest domain_common.BridgeTxData) error
}

type DefaultExecutor struct {
	chainId      uint64
	chainName    string
	batchHandler BatchHandler
	batchOutput  chan []domain_common.BridgeTxData
	endChannel   chan []domain_common.BridgeTxData
}

func NewExecutor(network observer.Network, pendingRequests chan domain_common.BridgeTxData, endChannel chan []domain_common.BridgeTxData) *DefaultExecutor {
	batchOutput := make(chan []domain_common.BridgeTxData)
	return &DefaultExecutor{
		chainId:      network.ChainId,
		chainName:    network.Name,
		batchHandler: NewDummyBatchHandler(pendingRequests, batchOutput),
		batchOutput:  batchOutput,
		endChannel:   endChannel,
	}
}

func (e *DefaultExecutor) Start() error {
	// Start the background batching process
	go e.batchHandler.Start()

	for {
		batch := <-e.batchOutput
		log.Debug().Msgf("Received batch of %d requests in chainId %s", len(batch), e.chainName)
		e.endChannel <- batch
	}
}
