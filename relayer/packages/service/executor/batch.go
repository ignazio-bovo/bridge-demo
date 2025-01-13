package executor

import (
	"sync"

	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	"github.com/rs/zerolog/log"
)

type BatchHandler interface {
	Start()
}

type DummyBatchHandler struct {
	requestChan  chan domain_common.BridgeTxData
	outputChan   chan []domain_common.BridgeTxData
	currentBatch []domain_common.BridgeTxData
	maxSize      int
	batchMutex   sync.Mutex
}

func NewDummyBatchHandler(
	requestChan chan domain_common.BridgeTxData,
	outputChan chan []domain_common.BridgeTxData,
) *DummyBatchHandler {
	return &DummyBatchHandler{
		requestChan:  requestChan,
		outputChan:   outputChan,
		maxSize:      1,
		currentBatch: make([]domain_common.BridgeTxData, 0),
	}
}

func (h *DummyBatchHandler) Collect() ([]domain_common.BridgeTxData, error) {
	h.batchMutex.Lock()
	defer h.batchMutex.Unlock()

	// Try to read requests until batch is full or channel is empty
	for len(h.currentBatch) < h.maxSize {
		select {
		case req := <-h.requestChan:
			h.currentBatch = append(h.currentBatch, req)
		default:
			// No more requests available immediately
			break
		}
	}

	if len(h.currentBatch) == 0 {
		return nil, nil
	}

	// Create a copy of the current batch
	batch := make([]domain_common.BridgeTxData, len(h.currentBatch))
	copy(batch, h.currentBatch)

	// Clear the current batch
	h.currentBatch = make([]domain_common.BridgeTxData, 0)
	return batch, nil
}

func (h *DummyBatchHandler) Start() {
	for {
		batch, err := h.Collect()
		if err != nil {
			log.Error().Err(err).Msg("Error batching requests")
			continue
		}
		if len(batch) > 0 {
			h.outputChan <- batch
		}
	}
}
