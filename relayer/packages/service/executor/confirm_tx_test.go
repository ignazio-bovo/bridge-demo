package executor

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfirmTx(t *testing.T) (*ethclient.Client, []byte) {
	// Setup test environment
	destinationClient, err := ethclient.Dial("http://localhost:9944")
	require.NoError(t, err)
	// Test data representing a transfer request
	data := common.FromHex("0x640c89ee00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000000000000000000000000000000000000000000000414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000008457468657265756d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000034554480000000000000000000000000000000000000000000000000000000000")
	return destinationClient, data
}
func TestGasPriceEstimation(t *testing.T) {
	destinationClient, _ := setupTestConfirmTx(t)

	// returns min gas price
	gasPrice, err := destinationClient.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	assert.NotNil(t, gasPrice)
	assert.GreaterOrEqual(t, gasPrice.Uint64(), uint64(10000000000))
}

func TestGasLimitEstimation(t *testing.T) {
	destinationClient, data := setupTestConfirmTx(t)
	fromAddress := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	gasPrice, _ := destinationClient.SuggestGasPrice(context.Background())
	contractAddress := common.HexToAddress("0x057ef64E23666F000b34aE31332854aCBd1c8544")

	// Estimate gas limit
	msg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &contractAddress,
		Gas:      0, // will be estimated
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     data,
	}
	gasLimit, err := destinationClient.EstimateGas(context.Background(), msg)
	require.NoError(t, err)

	assert.NotNil(t, gasLimit)
	assert.Equal(t, gasLimit, uint64(71393))
}

func TestAbiPacking(t *testing.T) {
	contractABI, _ := domain_common.GetContractABI()
	request := domain_common.BridgeTxData{
		Amount:   big.NewInt(100000000000000000), // 0.1 ETH in Wei
		Nonce:    big.NewInt(0),
		To:       common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
		TokenKey: common.HexToHash("0x414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e"),
		TokenMetadata: domain_common.TokenMetadata{
			Symbol:   "ETH",
			Name:     "Ethereum",
			Decimals: 18,
		},
		SrcChainId: 1,
	}
	data, err := contractABI.Pack("executeTransferRequests", []domain_common.BridgeTxData{request})
	require.NoError(t, err)

	assert.NotNil(t, data)
	expectedData := "0x640c89ee00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000000000000000000000000000000000000000000000414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000008457468657265756d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000034554480000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expectedData, "0x"+hex.EncodeToString(data))
}

func TestBuildAndSignTransaction(t *testing.T) {
	destinationClient, data := setupTestConfirmTx(t)

	// Setup test data
	nonce := uint64(1)
	gasLimit := uint64(71393)
	value := big.NewInt(0)
	chainId := uint64(0)
	contractAddress := common.HexToAddress("0x057ef64E23666F000b34aE31332854aCBd1c8544")

	// Create private key for testing
	privateKey, err := crypto.HexToECDSA("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	require.NoError(t, err)
	gasTipCap, err := destinationClient.SuggestGasTipCap(context.Background())
	require.NoError(t, err)
	header, err := destinationClient.HeaderByNumber(context.Background(), nil)

	require.NoError(t, err)

	baseFee := header.BaseFee
	maxFeePerGas := new(big.Int).Mul(baseFee, big.NewInt(2))
	maxFeePerGas.Add(maxFeePerGas, gasTipCap)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(chainId)),
		Nonce:     nonce,
		To:        &contractAddress,
		Value:     value,
		Gas:       gasLimit,
		GasFeeCap: maxFeePerGas, // Max fee per gas (baseFee + priority fee)
		GasTipCap: gasTipCap,    // Max priority fee per gas (tip)
		Data:      data,
	})
	fmt.Println("tx", tx)
	assert.Equal(t, uint8(types.DynamicFeeTxType), tx.Type())

	// Build and sign transaction
	signer := types.NewLondonSigner(big.NewInt(int64(chainId)))
	hash := signer.Hash(tx)
	fmt.Println("hash", hash)
	assert.NotNil(t, hash)

	signature, _ := crypto.Sign(hash[:], privateKey)
	assert.NotNil(t, signature)

	signedTx, err := tx.WithSignature(signer, signature)
	require.NoError(t, err)
	assert.NotNil(t, signedTx)
	tx = signedTx
	v, r, s := tx.RawSignatureValues()
	assert.NotNil(t, v)
	assert.NotNil(t, r)
	assert.NotNil(t, s)

}
