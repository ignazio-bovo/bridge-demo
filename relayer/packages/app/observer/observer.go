package observer

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	observer_domain "github.com/Datura-ai/relayer/packages/domain/observer"
	"github.com/Datura-ai/relayer/packages/service/executor"
	executor_service "github.com/Datura-ai/relayer/packages/service/executor"
	observer_service "github.com/Datura-ai/relayer/packages/service/observer"
)

func StartObserver() {
	var networkConfig map[string]observer_domain.Network

	configPath := "./config/networks.yaml"
	networkConfig, err := observer_domain.ParseNetworkConfig(configPath)
	if err != nil {
		log.Error().Msgf("Error parsing network config: %v", err)
	}

	eventBus := observer_service.NewEmptyEventBus()

	// Parse all the keys in the networkConfig map
	if len(networkConfig) != 2 {
		log.Error().Msg("Invalid number of networks found in config")
		return
	}

	ethereumClient, err := ethclient.Dial(networkConfig["ethereum"].RpcEndpoint)
	if err != nil {
		log.Error().Msgf("Error connecting to ethereum client for %s: %v", networkConfig["ethereum"].Name, err)
		return
	}
	subtensorClient, err := ethclient.Dial(networkConfig["subtensor"].RpcEndpoint)
	if err != nil {
		log.Error().Msgf("Error connecting to ethereum client for %s: %v", networkConfig["subtensor"].Name, err)
		return
	}

	subtensorListener, subtensorExecutor, subtensorTxConfirmation := setupNetworkForRelayer(ethereumClient, eventBus, networkConfig["ethereum"], subtensorClient)
	ethereumListener, ethereumExecutor, ethereumTxConfirmation := setupNetworkForRelayer(subtensorClient, eventBus, networkConfig["subtensor"], ethereumClient)

	// Start the listeners
	startServices(ethereumListener, ethereumExecutor, ethereumTxConfirmation, networkConfig["ethereum"])
	startServices(subtensorListener, subtensorExecutor, subtensorTxConfirmation, networkConfig["subtensor"])
	select {}
}

func startServices(listener *observer_service.WebsocketEventListener, executor *executor.DefaultExecutor, txConfirmation *executor.DefaultConfirmTx, network observer_domain.Network) {
	go listener.Start()
	log.Debug().Msgf("👂 Started listener for %s", network.Name)
	go executor.Start()
	log.Debug().Msgf("🚀 Started executor for %s", network.Name)
	go txConfirmation.Run()
	log.Debug().Msgf("🔍 Started tx confirmation for %s", network.Name)
}

func setupNetworkForRelayer(ethClient *ethclient.Client, eventBus *observer_service.EventBus, network observer_domain.Network, otherClient *ethclient.Client) (*observer_service.WebsocketEventListener, *executor.DefaultExecutor, *executor.DefaultConfirmTx) {

	// create a new websocket connector

	parser := observer_service.NewDefaultEventParser(ethClient /*used to get metadata*/, network.ContractAddress)
	websocketListener, err := observer_service.NewWebsocketEventListener(network, ethClient, eventBus, parser)
	if err != nil {
		log.Error().Msgf("Error creating listener for %s: %v", network.Name, err)
		return nil, nil, nil
	}

	// create a new executor
	pendingRequests := make(chan domain_common.BridgeTxData) // request coming from the source chain
	eventBus.AddDestination(network.ChainId, pendingRequests)
	endChannel := make(chan []domain_common.BridgeTxData) // request going to the destination chain
	executor := executor.NewExecutor(network, pendingRequests, endChannel)

	txConfirmation, err := executor_service.NewDefaultConfirmTx(endChannel, ethClient, network, otherClient)
	if err != nil {
		log.Error().Msgf("Error creating tx confirmation for %s: %v", network.Name, err)
		return nil, nil, nil
	}

	return websocketListener, executor, txConfirmation
}
