package observer

import (
	"sync"

	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	"github.com/rs/zerolog/log"
)

// EventBus implements a message router with multiple producers and pre-defined destinations, mutex needed for syncing between producers
type EventBus struct {
	routes map[uint64]chan domain_common.BridgeTxData
	mu     sync.RWMutex
}

func NewEmptyEventBus() *EventBus {
	return &EventBus{
		routes: make(map[uint64]chan domain_common.BridgeTxData),
	}
}

func (eb *EventBus) AddDestination(topic uint64, destinationChannel chan domain_common.BridgeTxData) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	log.Debug().Msgf("🎯 Adding destination for topic: %d", topic)
	eb.routes[topic] = destinationChannel
}

func (eb *EventBus) Route(routingMessage domain_common.RoutingMessage) {
	eb.mu.RLock()
	ch, exists := eb.routes[routingMessage.DstChainId]
	eb.mu.RUnlock()

	if exists {
		ch <- routingMessage.TxData
	} else {
		log.Error().Msgf("No route found for topic: %d", routingMessage.DstChainId)
	}
}
