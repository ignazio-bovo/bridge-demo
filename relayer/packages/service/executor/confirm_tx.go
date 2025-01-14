package executor

import (
	"context"
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
}

func NewDefaultConfirmTx(endChannel chan []domain_common.BridgeTxData, destinationClient *ethclient.Client, network observer_domain.Network, sourceClient *ethclient.Client) (*DefaultConfirmTx, error) {
	// Get the contract ABI
	contractABI, err := domain_common.GetContractABI()
	if err != nil {
		log.Error().AnErr("Error getting contract ABI", err)
		return nil, err
	}

	return &DefaultConfirmTx{
		endChannel:             endChannel,
		destinationClient:      destinationClient,
		contractAddress:        network.ContractAddress,
		sleepDuration:          network.BlockProductionTime,
		chainId:                network.ChainId,
		contractABI:            contractABI,
		sourceClient:           sourceClient,
		destinationNativeToken: common.Address{},
	}, nil
}

func (c *DefaultConfirmTx) Run() {
	for {
		select {
		case batch := <-c.endChannel:
			log.Debug().Msgf("Received batch of %d requests in chainId %d", len(batch), c.chainId)
			if err := c.ConfirmConfirmTransferRequest(batch); err != nil {
				log.Error().AnErr("Error confirming transfer request", err)
				continue
			}
		default:
			// Non-blocking default case
			time.Sleep(c.sleepDuration)

		}
	}
}

func (c *DefaultConfirmTx) ConfirmConfirmTransferRequest(batch []domain_common.BridgeTxData) error {
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("Error loading .env file")
	}

	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		return fmt.Errorf("private key is not set in .env file")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	fromAddress := common.HexToAddress(os.Getenv("ADDRESS"))
	if fromAddress == (common.Address{}) {
		return fmt.Errorf("from address is not set in .env file")
	}

	nonce, err := c.destinationClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Error().AnErr("Error getting pending nonce", err)
		return err
	}

	gasPrice, err := c.destinationClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error().AnErr("Error getting suggested gas price", err)
		return err
	}

	// Pack the function call with parameters
	data, err := c.contractABI.Pack("executeTransferRequests", batch)
	if err != nil || len(data) == 0 {
		log.Error().AnErr("Error packing transaction data", err)
		return err
	}

	// Estimate gas limit
	msg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &c.contractAddress,
		Gas:      0, // will be estimated
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     data,
	}
	gasLimit, err := c.destinationClient.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Error().AnErr("Error estimating gas limit", err)
		// Fallback to default gas limit if estimation fails
		gasLimit = uint64(300000)
		log.Warn().Msgf("Using fallback gas limit of %d", gasLimit)
	}

	// Add 10% buffer to estimated gas limit
	gasLimit = uint64(float64(gasLimit) * 1.1)

	// Calculate total value from batch
	totalValue := big.NewInt(0)
	for _, request := range batch {
		log.Debug().Msgf("💰 Amount for user %s: %d", request.To, request.Amount)

		if c.isNativeToken(request.TokenKey) {
			log.Debug().Msgf("💰 Amount for user %s: %d", request.To, request.Amount)
			totalValue.Add(totalValue, request.Amount)
		}
	}

	// Get signer's balance
	balance, err := c.destinationClient.BalanceAt(context.Background(), fromAddress, nil)
	if err != nil {
		log.Error().AnErr("Error getting signer balance", err)
		return err
	}

	// Calculate total transaction cost (gas + value)
	gasCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	totalCost := new(big.Int).Add(gasCost, totalValue)

	log.Info().
		Str("signer_balance", balance.String()).
		Str("total_cost", totalCost.String()).
		Str("gas_cost", gasCost.String()).
		Str("transfer_value", totalValue.String()).
		Msg("💰 Transaction cost breakdown")

	if balance.Cmp(totalCost) < 0 {
		return fmt.Errorf("insufficient balance for transaction: have %s, need %s", balance.String(), totalCost.String())
	}

	// Update transaction parameters
	toAddress := c.contractAddress
	tx := types.NewTransaction(
		nonce,
		toAddress,
		totalValue, // Updated: now using the sum of non-native transfer amounts
		gasLimit,
		gasPrice,
		data,
	)

	chainID, err := c.destinationClient.NetworkID(context.Background())
	if err != nil {
		log.Error().AnErr("Error getting chain ID", err)
		return err
	}
	log.Debug().Msgf("Chain ID: %d", chainID)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Error().AnErr("Error signing transaction", err)
		return err
	}

	err = c.sendAndMonitorTransaction(signedTx)
	if err != nil {
		log.Debug().Err(err).Msg("Error sending transaction")
		return err
	}

	log.Debug().Str("tx_hash", signedTx.Hash().Hex()).Msg("🎯 Transaction sent")

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

		log.Debug().Err(err).Msg("Error sending transaction")
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
