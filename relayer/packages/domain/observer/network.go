package observer

import (
	"fmt"
	"io/ioutil"
	"time"

	geth_comm "github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v2"
)

type Network struct {
	Name                string            `yaml:"name"`
	ChainId             uint64            `yaml:"chain_id"`
	ContractAddress     geth_comm.Address `yaml:"address"`
	BlockConfirmations  int               `yaml:"block_confirmations"`
	RpcEndpoint         string            `yaml:"rpc_endpoint"`
	WSSEndpoint         string            `yaml:"wss_endpoint"`
	BlockProductionTime time.Duration     `yaml:"block_production_time"`
	StartBlock          uint64            `yaml:"start_block"`
}

type NetworkConfig map[string]Network

func ParseNetworkConfig(filename string) (NetworkConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var config NetworkConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %w", err)
	}

	return config, nil
}

func (n Network) String() string {
	return fmt.Sprintf("Network: %s, ChainId: %d, ContractAddress: %s, BlockConfirmations: %d, RpcEndpoint: %s, WSSEndpoint: %s", n.Name, n.ChainId, n.ContractAddress.Hex(), n.BlockConfirmations, n.RpcEndpoint, n.WSSEndpoint)
}
