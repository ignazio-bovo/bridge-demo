package chains

import (
	"fmt"
	"strings"

	"github.com/Datura-ai/relayer/packages/service/common"
	"github.com/Datura-ai/relayer/packages/service/tss/tss"
	"github.com/btcsuite/btcd/chaincfg"
)

// LoadChains returns chain clients from chain configuration
func LoadChains(
	cfg []common.ChainConfiguration,
	server *tss.TssServer,
) map[common.ChainID]ChainClient {
	chains := make(map[common.ChainID]ChainClient)

	for i, chain := range cfg {
		cli, err := LoadChain(chain, server, uint64(i))
		if err != nil {
			continue
		}
		chains[chain.ChainID] = cli
	}

	return chains
}

func determineBitcoinNetworkParams(rpcHost string) *chaincfg.Params {
	var networkParams *chaincfg.Params
	// determine network params for bitcoin
	if strings.Contains(rpcHost, "testnet") {
		networkParams = &chaincfg.TestNet3Params
	} else {
		networkParams = &chaincfg.MainNetParams
	}
	return networkParams
}

// LoadChains returns chain clients from chain configuration
func LoadChain(
	cfg common.ChainConfiguration,
	server *tss.TssServer,
	handlerID uint64,
) (ChainClient, error) {
	var chain ChainClient
	if !cfg.IsEvm {
		return nil, fmt.Errorf("not an evm chain")
	}
	chain, err := NewEvmClient(cfg, server, cfg.ChainID, common.PubKey(cfg.PoolAddress))
	if err != nil {
		return nil, err
	}
	return chain, nil
}
