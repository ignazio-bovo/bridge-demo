package tss

import (
	"fmt"
	"os"
	"time"

	dcommon "github.com/Datura-ai/relayer/packages/domain/common"
	domain "github.com/Datura-ai/relayer/packages/domain/tss"
	lrepository "github.com/Datura-ai/relayer/packages/repository/listener"
	trepository "github.com/Datura-ai/relayer/packages/repository/tss"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
	"github.com/Datura-ai/relayer/packages/service/tss/controller"
	"github.com/Datura-ai/relayer/packages/service/tss/discovery"
	"github.com/Datura-ai/relayer/packages/service/tss/p2p"
	tssserver "github.com/Datura-ai/relayer/packages/service/tss/server"
	"github.com/Datura-ai/relayer/packages/service/tss/storage"
	"github.com/Datura-ai/relayer/packages/service/tss/tss"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

func StartTss() {
	// Load configuration
	cfgFile := flag.StringP("cfg", "c", "tss.json", "configuration file with extension")
	configPath := "config/" + *cfgFile
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Error().Err(err).Msg("configuration file does not exist")
		panic("can't find tss configuration file")
	}

	tssCfg, err := domain.LoadConfig(*cfgFile, "config/")
	if err != nil {
		log.Error().Err(err).Msg("configuration error")
		panic("can't find tss")
	}

	log.Debug().Msg(fmt.Sprintf("%v", tssCfg))

	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Tss:%s", err.Error()))
		panic("Tss: Failed to import configuration file")
	}

	// start state manager here
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get home directory")
		return
	}
	tssStateManager := trepository.NewTssStateManager(fmt.Sprintf("%s/.datura", homeDir))
	listenerStateManager := lrepository.NewListenerStateManager(fmt.Sprintf("%s/.datura", homeDir))
	defer tssStateManager.Close()
	defer listenerStateManager.Close()
	bootstrapPeers, err := tssCfg.TssEcdsa.GetBootstrapPeers()
	if err != nil {
		log.Fatal().Err(err).Msg("fail to get bootstrap peers")
		return
	}

	// check peerId is existing or not
	peerId, err := tssStateManager.ID()
	if err != nil {
		log.Debug().Msg("Peer ID did not exist")
		// generate new ID
		peerId = GeneratePeerID(tssStateManager)
	}

	log.Debug().Msg(fmt.Sprintf("%v", bootstrapPeers))
	log.Debug().Msg(fmt.Sprintf("my peerID:%s", peerId))

	p2pComm, err := p2pCommunication(bootstrapPeers, tssStateManager, tssCfg.TssEcdsa.P2PPort, tssCfg.TssEcdsa.ExternalIP, tssCfg.TssEcdsa.Group)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to establish p2p communication")
		return
	}

	host := p2pComm.GetHost()
	log.Debug().Msg(fmt.Sprintf("host: %v", host))

	rawPrivKey, err := tssStateManager.GetPrivKey()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get private key")
		return
	}

	privKey, err := crypto.UnmarshalSecp256k1PrivateKey(rawPrivKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal private key")
		return
	}

	localStateMgr, err := storage.NewFileStateMgr(fmt.Sprintf("%s/.datura", homeDir))

	// let's generate key with peers
	// Create a stop channel for the keygen process
	//stopChan := make(chan struct{})

	commonCfg := common.TssConfig{
		PartyTimeout:        60 * time.Second,
		EcdsaTimeout:        60 * time.Second,
		KeyresharingTimeout: 60 * time.Second,
		KeysignTimeout:      120 * time.Second,
		KeygenTimeout:       60 * time.Second,
		PreParamTimeout:     60 * time.Second,
		EnableMonitor:       false,
	}

	newPeerChannel := make(chan []peer.ID)
	keysignChan := make(chan dcommon.BridgeTxData)
	partyCoordinator := p2p.NewPartyCoordinator(p2pComm.GetHost(), commonCfg.PartyTimeout)
	peerDiscovery := discovery.NewPeerDiscovery(p2pComm.GetHost(), tssStateManager)

	pubKeyStr := host.ID().String()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get peer ID")
		return
	}
	// Create a new TssKeygen instance
	tssIns, localParams, err := tss.NewTss(
		commonCfg,
		nil,
		p2pComm,
		tssStateManager,
		localStateMgr,
		partyCoordinator,
		peerDiscovery,
		pubKeyStr,
		privKey,
		0,
		&newPeerChannel,
		&keysignChan,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Tss")
		return
	}

	log.Debug().Msg(fmt.Sprintf("localParams: %v", localParams))

	s := tssserver.NewTssHttpServer(tssIns, "0.0.0.0"+tssCfg.TssEcdsa.InfoAddress)
	go func() {
		s.Start()
	}()
	time.Sleep(5 * time.Second)
	controller, err := controller.NewTssController(tssCfg, log.Logger, tssIns, tssStateManager, listenerStateManager)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create TssController")
		return
	}
	controller.Start()
}

func p2pCommunication(cmdBootstrapPeers p2p.AddrList, stateManager trepository.TssStateManager, p2pPort int, externalIP, groupName string) (*p2p.Communication, error) {
	var bootstrapPeers p2p.AddrList
	savedPeers := stateManager.GetPeerAddrList()
	bootstrapPeers = savedPeers
	bootstrapPeers = append(bootstrapPeers, cmdBootstrapPeers...)

	comm, err := p2p.NewCommunication(groupName, bootstrapPeers, p2pPort, externalIP)
	if err != nil {
		return nil, fmt.Errorf("fail to create communication layer: %w", err)
	}
	privKey, err := stateManager.GetPrivKey()
	if err != nil {
		return nil, fmt.Errorf("fail to get private key")
	}
	if err := comm.Start(privKey); nil != err {
		return nil, fmt.Errorf("fail to start p2p network: %w", err)
	}
	return comm, nil
}
