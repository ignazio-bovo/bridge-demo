package executor

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	observer_domain "github.com/Datura-ai/relayer/packages/domain/observer"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ConfirmTxSender interface {
	Run()
	ConfirmConfirmTransferRequest(batch []domain_common.BridgeTxData) (*types.Receipt, error)
}

type DefaultConfirmTx struct {
	endChannel             chan []domain_common.BridgeTxData
	destinationClient      *ethclient.Client
	sourceClient           *ethclient.Client
	contractAddress        common.Address
	sleepDuration          time.Duration
	chainId                uint64
	contractABI            *abi.ABI
	destinationNativeToken common.Address
	logger                 *zerolog.Logger
}

func NewDefaultConfirmTx(endChannel chan []domain_common.BridgeTxData, destinationClient *ethclient.Client, network observer_domain.Network, sourceClient *ethclient.Client) (*DefaultConfirmTx, error) {
	// Get the contract ABI
	contractABI, err := domain_common.GetContractABI()
	if err != nil {
		log.Error().AnErr("Error getting contract ABI", err)
		return nil, err
	}
	logger := log.Logger.With().Str("Network", network.Name).Str("Module", "Listener").Logger()

	return &DefaultConfirmTx{
		endChannel:             endChannel,
		destinationClient:      destinationClient,
		contractAddress:        network.ContractAddress,
		sleepDuration:          network.BlockProductionTime,
		chainId:                network.ChainId,
		contractABI:            contractABI,
		sourceClient:           sourceClient,
		destinationNativeToken: common.Address{},
		logger:                 &logger,
	}, nil
}

func (c *DefaultConfirmTx) Run() {
	for {
		select {
		case batch := <-c.endChannel:
			log.Debug().Msgf("Received batch of %d requests in chainId %d", len(batch), c.chainId)
			if err := c.ConfirmTransferRequest(batch); err != nil {
				log.Error().AnErr("Error confirming transfer request", err)
				continue
			}
		default:
			// Non-blocking default case
			time.Sleep(c.sleepDuration)

		}
	}
}

func (c *DefaultConfirmTx) getSignerCredentials() (*ecdsa.PrivateKey, common.Address, error) {
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		return nil, common.Address{}, fmt.Errorf("private key is not set in .env file")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to load private key: %w", err)
	}

	fromAddress := common.HexToAddress(os.Getenv("ADDRESS"))
	if fromAddress == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf("from address is not set in .env file")
	}

	return privateKey, fromAddress, nil
}

func (c *DefaultConfirmTx) getGasPrice() (*big.Int, error) {
	gasPrice, err := c.destinationClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error().AnErr("Error getting suggested gas price", err)
		return nil, err
	}
	// 10% increase
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(110))
	gasPrice = gasPrice.Div(gasPrice, big.NewInt(100))
	c.logger.Debug().Msgf("💰 Gas price: %d", gasPrice)
	return gasPrice, nil
}

func (c *DefaultConfirmTx) estimateGasLimit(from common.Address, gasPrice *big.Int, data []byte) (uint64, error) {
	msg := ethereum.CallMsg{
		From:     from,
		To:       &c.contractAddress,
		Gas:      0,
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     data,
	}

	gasLimit, err := c.destinationClient.EstimateGas(context.Background(), msg)
	if err != nil {
		c.logger.Error().AnErr("Error estimating gas limit", err)
		gasLimit = uint64(400000) // Fallback
		c.logger.Warn().Msgf("Using fallback gas limit of %d on chainId %d", gasLimit, c.chainId)
	}
	c.logger.Debug().Msgf("⛽ Gas limit: %d", gasLimit)
	return gasLimit, nil
}

func (c *DefaultConfirmTx) calculateTotalValue(batch []domain_common.BridgeTxData) *big.Int {
	totalValue := big.NewInt(0)
	for _, request := range batch {
		if c.isNativeToken(request.TokenKey) {
			c.logger.Debug().Msgf("💰 Amount for user %s: %d", request.To, request.Amount)
			totalValue.Add(totalValue, request.Amount)
		}
	}
	return totalValue
}

func (c *DefaultConfirmTx) checkBalance(from common.Address, gasPrice *big.Int, gasLimit uint64, totalValue *big.Int) error {
	balance, err := c.destinationClient.BalanceAt(context.Background(), from, nil)
	if err != nil {
		c.logger.Error().AnErr("Error getting signer balance", err)
		return err
	}

	gasCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	totalCost := new(big.Int).Add(gasCost, totalValue)

	c.logger.Info().
		Str("signer_balance", balance.String()).
		Str("total_cost", totalCost.String()).
		Str("gas_cost", gasCost.String()).
		Str("transfer_value", totalValue.String()).
		Msg("💰 Transaction cost breakdown")

	if balance.Cmp(totalCost) < 0 {
		return fmt.Errorf("insufficient balance for transaction: have %s, need %s", balance.String(), totalCost.String())
	}
	return nil
}

func (c *DefaultConfirmTx) buildAndSignTransaction(nonce uint64, gasLimit uint64, value *big.Int, data []byte, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	// Get the suggested priority fee (tip)
	gasTipCap, err := c.destinationClient.SuggestGasTipCap(context.Background())
	if err != nil {
		c.logger.Error().AnErr("Error getting suggested tip cap", err)
		return nil, err
	}

	// Get the latest header to fetch current base fee
	header, err := c.destinationClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		c.logger.Error().AnErr("Error getting latest header", err)
		return nil, err
	}

	// Calculate maxFeePerGas = (2 * baseFee) + priorityFee
	baseFee := header.BaseFee
	maxFeePerGas := new(big.Int).Mul(baseFee, big.NewInt(2))
	maxFeePerGas.Add(maxFeePerGas, gasTipCap)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(c.chainId)),
		Nonce:     nonce,
		To:        &c.contractAddress,
		Value:     value,
		Gas:       gasLimit,
		GasFeeCap: maxFeePerGas, // Max fee per gas (baseFee + priority fee)
		GasTipCap: gasTipCap,    // Max priority fee per gas (tip)
		Data:      data,
	})

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(big.NewInt(int64(c.chainId))), privateKey)
	if err != nil {
		c.logger.Error().AnErr("Error signing transaction", err)
		return nil, err
	}

	c.logger.Debug().
		Str("baseFee", baseFee.String()).
		Str("maxFeePerGas", maxFeePerGas.String()).
		Str("gasTipCap", gasTipCap.String()).
		Msg("Transaction signed with gas parameters")

	return signedTx, nil
}

func (c *DefaultConfirmTx) ConfirmTransferRequest(batch []domain_common.BridgeTxData) error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		c.logger.Warn().Msg("Error loading .env file")
	}

	// Get private key and address from env
	privateKey, fromAddress, err := c.getSignerCredentials()
	if err != nil {
		return err
	}

	// Get transaction parameters
	nonce, err := c.destinationClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		c.logger.Error().AnErr("Error getting pending nonce", err)
		return err
	}

	// Get gas parameters
	gasPrice, err := c.getGasPrice()
	if err != nil {
		return err
	}

	// Pack transaction data
	data, err := c.contractABI.Pack("executeTransferRequests", batch)
	if err != nil || len(data) == 0 {
		c.logger.Error().AnErr("Error packing transaction data", err)
		return err
	}

	// Get gas limit
	gasLimit, err := c.estimateGasLimit(fromAddress, gasPrice, data)
	if err != nil {
		return err
	}

	// Calculate total value to transfer
	totalValue := c.calculateTotalValue(batch)

	// Verify sufficient balance
	if err := c.checkBalance(fromAddress, gasPrice, gasLimit, totalValue); err != nil {
		return err
	}

	// Build and sign transaction
	signedTx, err := c.buildAndSignTransaction(
		nonce,
		gasLimit,
		totalValue,
		data,
		privateKey,
	)
	if err != nil {
		return err
	}

	// Send and monitor transaction
	if err := c.sendAndMonitorTransaction(signedTx); err != nil {
		c.logger.Debug().Err(err).Msg("Error sending transaction")
		return err
	}

	c.logger.Debug().Str("tx_hash", signedTx.Hash().Hex()).Msg("🎯 Transaction sent")
	return nil
}

func (c *DefaultConfirmTx) sendAndMonitorTransaction(signedTx *types.Transaction) error {
	// Send the transaction
	err := c.destinationClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		// Check for specific error types first
		if err == ethereum.NotFound {
			return fmt.Errorf("transaction not found: %w", err)
		}

		// Check for revert strings in the error
		if strings.Contains(err.Error(), "execution reverted") {
			// Try to extract the revert reason
			reason, extractErr := extractRevertReason(err.Error())
			if extractErr == nil {
				return fmt.Errorf("transaction reverted: %s", reason)
			}
		}

		c.logger.Debug().Err(err).Msg("Error sending transaction")
		return err
	}

	// Wait for transaction receipt
	receipt, err := waitForTxReceipt(c.destinationClient, signedTx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Check transaction status
	if receipt.Status == 0 {
		// Transaction failed - try to get revert reason
		reason, err := getRevertReason(c.destinationClient, signedTx, receipt.BlockNumber)
		if err == nil {
			return fmt.Errorf("transaction failed: %s", reason)
		}
		return fmt.Errorf("transaction failed: status 0")
	}

	return nil
}

// Helper function to wait for transaction receipt
func waitForTxReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for i := 0; i < 50; i++ { // Retry up to 50 times
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, err
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("transaction receipt not found after timeout")
}

// Helper function to get revert reason
func getRevertReason(client *ethclient.Client, tx *types.Transaction, blockNumber *big.Int) (string, error) {
	// Create a message call
	msg := ethereum.CallMsg{
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}

	// Call the transaction
	_, err := client.CallContract(context.Background(), msg, blockNumber)
	if err != nil {
		// Try to extract revert reason from error
		return extractRevertReason(err.Error())
	}

	return "", nil
}

// Helper function to extract revert reason from error string
func extractRevertReason(errStr string) (string, error) {
	// Look for revert reason in various formats
	revertReasonRegex := regexp.MustCompile(`execution reverted: (.*?)(?:\n|$)`)
	matches := revertReasonRegex.FindStringSubmatch(errStr)
	if len(matches) > 1 {
		return matches[1], nil
	}

	// Check for hex-encoded revert reason
	hexRegex := regexp.MustCompile(`0x[0-9a-fA-F]*`)
	hexMatch := hexRegex.FindString(errStr)
	if hexMatch != "" {
		decoded, err := hex.DecodeString(hexMatch[2:]) // Remove '0x' prefix
		if err == nil && len(decoded) > 4 {            // Skip function selector (4 bytes)
			return string(decoded[4:]), nil
		}
	}

	return "", fmt.Errorf("could not extract revert reason")
}

func (c *DefaultConfirmTx) isNativeToken(tokenKey common.Hash) bool {
	return c.isNativeEth(tokenKey) || c.isNativeBittensor(tokenKey)
}

func (c *DefaultConfirmTx) isNativeEth(tokenKey common.Hash) bool {
	const nativeEthKey = "0x414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e"
	return c.chainId == 1 && tokenKey == common.HexToHash(nativeEthKey)
}

func (c *DefaultConfirmTx) isNativeBittensor(tokenKey common.Hash) bool {
	const nativeBittensorKey = "0x3a636391d72d0aec588d3a7908f5b3950fa7ac843ef46bf86ed066ba010044de"
	return c.chainId == 945 && tokenKey == common.HexToHash(nativeBittensorKey)
}
