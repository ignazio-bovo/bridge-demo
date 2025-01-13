package observer

import (
	"context"
	"fmt"
	"testing"

	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

// First, define your Go struct to match the Solidity struct
type TokenMetadata struct {
	Name     string
	Symbol   string
	Decimals uint8
}

func metadataCall(client *ethclient.Client, bridgeAddr common.Address, tokenKey common.Hash) (TokenMetadata, error) {
	abi, err := domain_common.GetContractABI()
	if err != nil {
		return TokenMetadata{}, fmt.Errorf("failed to get contract ABI: %w", err)
	}

	// Pack the call data
	data, err := abi.Pack("metadataForToken", tokenKey)
	if err != nil {
		return TokenMetadata{}, fmt.Errorf("failed to pack metadataForToken call: %w", err)
	}

	// Make the call
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &bridgeAddr,
		Data: data,
	}, nil)
	if err != nil {
		return TokenMetadata{}, fmt.Errorf("failed to call metadataForToken: %w", err)
	}

	// For debugging
	// fmt.Printf("result: %x\n", result)

	// Unpack the result - note we use "metadataForToken" here, not "tokenMetadata"
	var unpacked struct {
		Name     string
		Symbol   string
		Decimals uint8
	}
	err = abi.UnpackIntoInterface(&unpacked, "metadataForToken", result)
	if err != nil {
		return TokenMetadata{}, fmt.Errorf("failed to unpack metadata: %w", err)
	}

	return TokenMetadata{
		Name:     unpacked.Name,
		Symbol:   unpacked.Symbol,
		Decimals: unpacked.Decimals,
	}, nil
}

func TestFetchTokenMetadata(t *testing.T) {
	// Common test setup
	setupTest := func(t *testing.T) TokenMetadata {
		client, err := ethclient.Dial("http://localhost:8545")
		if err != nil {
			t.Fatalf("Failed to connect to Ethereum client: %v", err)
		}
		bridgeAddr := common.HexToAddress("0x71C95911E9a5D330f4D621842EC243EE1343292e")
		tokenKey := common.HexToHash("0x414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e")
		metadata, err := metadataCall(client, bridgeAddr, tokenKey)
		if err != nil {
			t.Fatalf("Failed to get metadata: %v", err)
		}
		print(fmt.Sprintf("metadata: %x", metadata))
		return metadata
	}

	t.Run("Should return metadata matching params", func(t *testing.T) {
		metadata := setupTest(t)
		assert.Equal(t, metadata.Name, "Ether", "Name should be Ether")
		assert.Equal(t, metadata.Symbol, "ETH", "Symbol should be ETH")
		assert.Equal(t, metadata.Decimals, uint8(18), "Decimals should be 18")
	})

}
