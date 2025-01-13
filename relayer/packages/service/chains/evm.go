package chains

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ecommon "github.com/ethereum/go-ethereum/common"
	ecore "github.com/ethereum/go-ethereum/core"
	etypes "github.com/ethereum/go-ethereum/core/types"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	tssp "github.com/Datura-ai/relayer/packages/service/tss/tss"

	domain "github.com/Datura-ai/relayer/packages/domain/common"
	"github.com/Datura-ai/relayer/packages/service/common"
	tss "github.com/Datura-ai/relayer/packages/service/tss/controller"

	erc20 "github.com/Datura-ai/relayer/packages/service/chains/abis/erc20"
)

const (
	MaxContractGas = 2000000
)

// Client is a structure to sign and broadcast tx to Ethereum chain used by signer mostly
type Client struct {
	logger zerolog.Logger
	cfg    common.ChainConfiguration
	client *ethclient.Client
	*KeySignWrapper
	tssKeySigner *tss.KeySign
	wg           *sync.WaitGroup
	erc20ABI     *abi.ABI
	stopchan     chan struct{}
	chain        common.ChainID
	eipSigner    etypes.EIP155Signer
	handleID     uint64
}

// NewEvmClient create new instance of Ethereum client
func NewEvmClient(
	cfg common.ChainConfiguration,
	server *tssp.TssServer,
	chain common.ChainID,
	pubKey common.PubKey,
) (*Client, error) {

	tssKm, err := tss.NewKeySign(server)
	if err != nil {
		return nil, fmt.Errorf("fail to create tss signer: %w", err)
	}

	ethClient, err := ethclient.Dial(cfg.BlockScanner.RPCHost)
	if err != nil {
		return nil, fmt.Errorf("fail to dial ETH rpc host(%s): %w", cfg.BlockScanner.RPCHost, err)
	}
	chainID, err := getChainID(ethClient, cfg.BlockScanner.HTTPRequestTimeout)
	if err != nil {
		return nil, err
	}

	KeySignWrapper, err := NewKeySignWrapper(pubKey, tssKm)
	if err != nil {
		return nil, fmt.Errorf("fail to create "+chain.String()+" key sign wrapper: %w", err)
	}

	erc20ABI, err := domain.GetContractABI()
	if err != nil {
		return nil, fmt.Errorf("fail to get contract abi: %w", err)
	}

	eipSigner := etypes.NewEIP155Signer(chainID)

	c := &Client{
		logger:         log.With().Str("module", chain.String()).Logger(),
		cfg:            cfg,
		client:         ethClient,
		KeySignWrapper: KeySignWrapper,
		tssKeySigner:   tssKm,
		wg:             &sync.WaitGroup{},
		stopchan:       make(chan struct{}),
		erc20ABI:       erc20ABI,
		chain:          chain,
		eipSigner:      eipSigner,
		handleID:       1,
	}

	c.logger.Info().Msgf("current chain id: %d", chainID.Uint64())
	if chainID.Uint64() == 0 {
		return nil, fmt.Errorf("chain id is: %d , invalid", chainID.Uint64())
	}
	return c, nil
}

// Start to monitor Ethereum block chain
func (c *Client) Start() {
	c.tssKeySigner.Start()
	// c.blockScanner.Start(globalTxsQueue)
}

// Stop ETH client
func (c *Client) Stop() {
	c.tssKeySigner.Stop()
	// c.blockScanner.Stop()
	c.client.Close()
	close(c.stopchan)
	c.wg.Wait()
}

// GetConfig return the configurations used by ETH chain
func (c *Client) GetConfig() common.ChainConfiguration {
	return c.cfg
}

func (c *Client) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.cfg.BlockScanner.HTTPRequestTimeout)
}

// getChainID retrieve the chain id from ETH node, and determinate whether we are running on test net by checking the status
// when it failed to get chain id , it will assume LocalNet
func getChainID(client *ethclient.Client, timeout time.Duration) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to get chain id ,err: %w", err)
	}
	return chainID, err
}

// GetChain get chain
func (c *Client) GetChain() common.ChainID {
	return c.chain
}

// GetAddress return current signer address, it will be bech32 encoded address
func (c *Client) GetAddress(poolPubKey common.PubKey) string {
	addr, err := poolPubKey.GetAddress(common.ETHChainID)
	if err != nil {
		c.logger.Error().Err(err).Str("pool_pub_key", poolPubKey.String()).Msg("fail to get pool address")
		return ""
	}
	return addr.String()
}

// GetNonce gets nonce
func (c *Client) GetNonce(addr string) (uint64, error) {
	ctx, cancel := c.getContext()
	defer cancel()
	nonce, err := c.client.PendingNonceAt(ctx, ecommon.HexToAddress(addr))
	if err != nil {
		return 0, fmt.Errorf("fail to get account nonce: %w", err)
	}
	return nonce, nil
}

// SignTx sign the the given TxArrayItem
func (c *Client) SignTx(tx domain.BridgeTxData, poolPubKey common.PubKey, msgId string) ([]byte, error) {
	if tx.From == "" {
		return nil, fmt.Errorf("from address is empty")
	}

	token_addr := tx.SrcToken

	// TODO: should do address check instead of length validation.
	if len(token_addr) < 1 {
		return nil, fmt.Errorf("to address is empty")
	}

	// Token Address
	tokenAddr := ecommon.HexToAddress(token_addr)
	// instance, err := token.NewMain(tokenAddr, c.client)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to create an instance")
	// }

	//poolPubKey := common.PubKey(tx.VaultAddress)

	// Standardize transferring amount according to the decimal count

	chainID, err := common.NewChainID(fmt.Sprintf("%d", tx.DestChainId))
	if err != nil {
		return nil, fmt.Errorf("cannot create chain id (%s): %w", tx.DestChainId, err)
	}

	dest, err := poolPubKey.GetAddress(chainID)
	if err != nil {
		return nil, fmt.Errorf("fail to get address for pub key(%s): %w", dest, err)
	}

	fromAddr := ecommon.HexToAddress(tx.From)
	toAddress := ecommon.HexToAddress(dest.String())

	data, err := c.erc20ABI.Pack("transferFrom", fromAddr, toAddress, tx.Amount)
	if err != nil {
		return nil, fmt.Errorf("fail to create data to call smart contract(transferFrom): %w", err)
	}

	nonce, err := c.GetNonce(dest.String())
	if err != nil {
		return nil, fmt.Errorf("fail to fetch account(%s) nonce : %w", dest, err)
	}
	c.logger.Info().Uint64("nonce", nonce).Msg("account info")

	gasPrice, _ := c.client.SuggestGasPrice(context.Background())

	estimatedNativeValue := big.NewInt(0)

	// Create Tx
	createdTx := etypes.NewTransaction(nonce, tokenAddr, estimatedNativeValue, MaxContractGas, gasPrice, data)
	// calculate hash to sign
	hashToSign := c.eipSigner.Hash(createdTx).Bytes()
	// Sign the hash with TSS
	// we set lastInput flag to true as we have only single input
	signature, hashInBytes, err := c.SignTSS(hashToSign, poolPubKey, msgId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign message")
	}

	poolPubKeyBytes := []byte(poolPubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert pool public key to bytes")
	}

	// sanity check for signature
	if ecrypto.VerifySignature(poolPubKeyBytes, hashInBytes, signature[:64]) {
		c.logger.Info().Msg("we can successfully verify the bytes")
	} else {
		c.logger.Error().Msg("Oops! we cannot verify the bytes")
		// return nil, errors.New("signature verification failed")
	}

	signedTx, err := createdTx.WithSignature(c.eipSigner, signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to append signature to tx")
	}
	return signedTx.MarshalJSON() // return signed tx as JSON []byte
}

// BroadcastTx decodes tx using rlp and broadcasts too Ethereum chain
func (c *Client) BroadcastTx(txOutItem domain.BridgeTxData, hexTx []byte) (string, error) {
	tx := &etypes.Transaction{}
	if err := tx.UnmarshalJSON(hexTx); err != nil {
		return "", err
	}
	ctx, cancel := c.getContext()
	defer cancel()
	if err := c.client.SendTransaction(ctx, tx); err != nil && err.Error() != ecore.ErrKnownBlock.Error() && err.Error() != ecore.ErrNonceTooLow.Error() {
		return "", err
	}
	txID := tx.Hash().String()
	c.logger.Info().Msgf("Signex tx was successfuly broadcasted to %s chain, hash: %s", txOutItem.DestChainId, txID)

	return txID, nil
}

func (c *Client) convertSigningAmount(amt *big.Int, token string) *big.Int {

	var value big.Int
	amt = amt.Mul(amt, value.Exp(big.NewInt(10), big.NewInt(int64(18)), nil))
	amt = amt.Div(amt, value.Exp(big.NewInt(10), big.NewInt(18), nil))
	return amt
}

// ERC20Token represents an ERC20 token contract.
type ERC20Token struct {
	address  ecommon.Address
	instance *erc20.Clients
}

func NewERC20Token(address ecommon.Address, client *ethclient.Client) (*ERC20Token, error) {
	instance, err := erc20.NewClients(address, client)
	if err != nil {
		return nil, err
	}
	return &ERC20Token{address: address, instance: instance}, nil
}
