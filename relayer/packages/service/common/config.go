package common

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	maddr "github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"
)

var CONFIRM_CNT = uint64(5)

type Configuration struct {
	Signer    SignerConfiguration  `json:"signer" mapstructure:"signer"`
	Kimachain ClientConfiguration  `json:"kimachain" mapstructure:"kimachain"`
	Chains    []ChainConfiguration `json:"chains" mapstructure:"chains"`
	TssEcdsa  TSSConfiguration     `json:"tss_ecdsa" mapstructure:"tss_ecdsa"`
	TssEddsa  TSSConfiguration     `json:"tss_eddsa" mapstructure:"tss_eddsa"`
}

// SignerConfiguration all the configures need by signer
type SignerConfiguration struct {
	SignerDbPath string                    `json:"signer_db_path" mapstructure:"signer_db_path"`
	BlockScanner BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
}

// ChainConfiguration configuration
type ChainConfiguration struct {
	ChainID      ChainID                   `json:"chain_id" mapstructure:"chain_id"`
	PoolAddress  string                    `json:"pool" mapstructure:"pool"`
	BlockScanner BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
	IsEvm        bool                      `json:"is_evm" mapstructure:"is_evm"`
}

// TSSConfiguration
type TSSConfiguration struct {
	BootstrapPeers []string `json:"bootstrap_peers" mapstructure:"bootstrap_peers"`
	ExternalIP     string   `json:"external_ip" mapstructure:"external_ip"`
	P2PPort        int      `json:"p2p_port" mapstructure:"p2p_port"`
	InfoAddress    string   `json:"info_address" mapstructure:"info_address"`
	Group          string   `json:"group" mapstructure:"group"`
}

// BlockScannerConfiguration settings for BlockScanner
type BlockScannerConfiguration struct {
	RPCHost            string        `json:"rpc_host" mapstructure:"rpc_host"`
	WSSHost            string        `json:"wss_host" mapstructure:"wss_host"`
	HTTPRequestTimeout time.Duration `json:"http_request_timeout" mapstructure:"http_request_timeout"`
}

// ClientConfiguration
type ClientConfiguration struct {
	ChainID      ChainID `json:"chain_id" mapstructure:"chain_id" `
	ChainHost    string  `json:"chain_host" mapstructure:"chain_host"`
	ChainRPC     string  `json:"chain_rpc" mapstructure:"chain_rpc"`
	ChainGRPC    string  `json:"chain_grpc" mapstructure:"chain_grpc"`
	SignerName   string  `json:"signer_name" mapstructure:"signer_name"`
	SignerPasswd string  `json:"signer_passwd" mapstructure:"signer_passwd"`
}

// LoadConfig read the tss-proc configuration from the given file
func LoadConfig(file string, DefaultNodeHome string) (*Configuration, error) {
	applyDefaultConfig()
	var cfg Configuration
	viper.AddConfigPath(DefaultNodeHome)
	viper.AddConfigPath(filepath.Dir(file))
	viper.SetConfigName(strings.TrimRight(path.Base(file), ".json"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fail to read from config file: %w", err)
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("fail to unmarshal: %w", err)
	}

	for _, chain := range cfg.Chains {
		if err := chain.ChainID.Validate(); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

// GetBootstrapPeers return the internal bootstrap peers in a slice of maddr.Multiaddr
func (c TSSConfiguration) GetBootstrapPeers() ([]maddr.Multiaddr, error) {
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
	viper.SetDefault("kimachain.chain_id", "testkima")
	viper.SetDefault("kimachain.chain_host", "localhost:1317")
	applyDefaultSignerConfig()
}

func applyDefaultSignerConfig() {
	viper.SetDefault("signer.signer_db_path", "signer_db")
}
