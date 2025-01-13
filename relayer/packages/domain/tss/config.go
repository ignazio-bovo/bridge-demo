package tss

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	maddr "github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"
)

// TSSConfiguration
type TSSConfig struct {
	BootstrapPeers []string `json:"bootstrap_peers" mapstructure:"bootstrap_peers"`
	ExternalIP     string   `json:"external_ip" mapstructure:"external_ip"`
	P2PPort        int      `json:"p2p_port" mapstructure:"p2p_port"`
	InfoAddress    string   `json:"info_address" mapstructure:"info_address"`
	Group          string   `json:"group" mapstructure:"group"`
}

// ChainClient configuration
type ChainClientConfig struct {
	ChainID  string       `json:"chain_id" mapstructure:"chain_id"`
	Contract string       `json:"contract" mapstructure:"contract"`
	Client   ClientConfig `json:"client" mapstructure:"client"`
	IsEVM    bool         `json:"is_evm" mapstructure:"is_evm"`
}

type ClientConfig struct {
	RPCHost            string        `json:"rpc_host" mapstructure:"rpc_host"`
	WSSHost            string        `json:"wss_host" mapstructure:"wss_host"`
	HTTPRequestTimeout time.Duration `json:"http_request_timeout" mapstructure:"http_request_timeout"`
}

type SignerConfig struct {
	SignerDbPath string            `json:"signer_db_path" mapstructure:"signer_db_path"`
	Client       ChainClientConfig `json:"block_scanner" mapstructure:"block_scanner"`
}

type Configuration struct {
	Signer   SignerConfig        `json:"signer" mapstructure:"signer"`
	Chains   []ChainClientConfig `json:"chains" mapstructure:"chains"`
	TssEcdsa TSSConfig           `json:"tss_ecdsa" mapstructure:"tss_ecdsa"`
}

func (c *Configuration) GetRPCHost(chainID string) (string, error) {
	for _, chain := range c.Chains {
		if chain.ChainID == chainID {
			return chain.Client.RPCHost, nil
		}
	}
	return "", fmt.Errorf("chain not found")
}

// LoadConfig reads the TSS configuration from the given file
func LoadConfig(file string, DefaultNodeHome string) (*Configuration, error) {
	applyDefaultConfig()
	var cfg Configuration

	// Ensure the file has the correct extension
	if filepath.Ext(file) != ".json" {
		return nil, fmt.Errorf("configuration file must have a .json extension")
	}

	viper.SetConfigType("json") // Specify the config file type
	viper.AddConfigPath(DefaultNodeHome)
	viper.AddConfigPath(filepath.Dir(file))
	viper.SetConfigName(strings.TrimSuffix(path.Base(file), ".json"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read from config file: %w", err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &cfg, nil
}

// GetBootstrapPeers return the internal bootstrap peers in a slice of maddr.Multiaddr
func (c TSSConfig) GetBootstrapPeers() ([]maddr.Multiaddr, error) {
	var addrs []maddr.Multiaddr
	for _, item := range c.BootstrapPeers {
		if len(item) > 0 {
			addr, err := maddr.NewMultiaddr(item)
			if err != nil {
				return nil, fmt.Errorf("fail to parse multi addr(%s): %w", item, err)
			}
			addrs = append(addrs, addr)
		}
	}
	return addrs, nil
}

func applyDefaultConfig() {
	viper.SetDefault("signer.signer_db_path", "signer_db")
}
