package observer

import (
	"context"
	"fmt"
	"math/big"

	domain_common "github.com/Datura-ai/relayer/packages/domain/common"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth_comm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

type EventParser interface {
	ParseEvent([]geth_comm.Hash, []byte) (domain_common.RoutingMessage, error)
}

type DefaultEventParser struct {
	abi           *abi.ABI
	client        *ethclient.Client
	metadataCache map[common.Hash]domain_common.TokenMetadata
	address       common.Address
}

func NewDefaultEventParser(client *ethclient.Client, address common.Address) *DefaultEventParser {
	abi, err := domain_common.GetContractABI()
	if err != nil {
		log.Fatal().AnErr("Error getting contract ABI", err)
	}
	return &DefaultEventParser{abi: abi, client: client, address: address, metadataCache: make(map[common.Hash]domain_common.TokenMetadata)}
}

func (p *DefaultEventParser) ParseEvent(topics []geth_comm.Hash, data []byte) (domain_common.RoutingMessage, error) {
	var transfer domain_common.BridgeTxData

	// Define struct matching the TransferRequest tuple structure
	type TransferRequest struct {
		Request struct {
			Nonce       *big.Int
			From        common.Address
			To          common.Address
			TokenKey    [32]byte
			Amount      *big.Int
			SrcChainId  uint64
			DestChainId uint64
		}
	}

	var request TransferRequest
	err := p.abi.UnpackIntoInterface(&request, "TransferRequested", data)
	if err != nil {
		return domain_common.RoutingMessage{}, fmt.Errorf("failed to unpack TransferRequested event: %w", err)
	}

	// Fill transfer from the nested Request struct
	transfer.Nonce = request.Request.Nonce
	transfer.To = request.Request.To
	transfer.TokenKey = common.BytesToHash(request.Request.TokenKey[:])
	transfer.Amount = request.Request.Amount
	transfer.SrcChainId = request.Request.SrcChainId

	// Get metadata
	metadata, err := p.GetTokenMetadata(p.client, p.address, transfer.TokenKey)
	if err != nil {
		return domain_common.RoutingMessage{}, fmt.Errorf("failed to get token metadata: %w", err)
	}
	transfer.TokenMetadata = metadata

	return domain_common.RoutingMessage{
		TxData:     transfer,
		DstChainId: request.Request.DestChainId,
	}, nil
}

func (p *DefaultEventParser) GetTokenMetadata(client *ethclient.Client, bridgeAddr common.Address, tokenKey common.Hash) (domain_common.TokenMetadata, error) {
	if metadata, ok := p.metadataCache[tokenKey]; ok {
		return metadata, nil
	}

	abi, err := domain_common.GetContractABI()
	if err != nil {
		return domain_common.TokenMetadata{}, fmt.Errorf("failed to get contract ABI: %w", err)
	}

	// Pack the call data
	data, err := abi.Pack("metadataForToken", tokenKey)
	if err != nil {
		return domain_common.TokenMetadata{}, fmt.Errorf("failed to pack metadataForToken call: %w", err)
	}

	// Make the call
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &bridgeAddr,
		Data: data,
	}, nil)
	if err != nil {
		return domain_common.TokenMetadata{}, fmt.Errorf("failed to call metadataForToken: %w", err)
	}

	// Unpack the result - note we use "metadataForToken" here, not "tokenMetadata"
	var unpacked struct {
		Name     string
		Symbol   string
		Decimals uint8
	}
	err = abi.UnpackIntoInterface(&unpacked, "metadataForToken", result)
	if err != nil {
		return domain_common.TokenMetadata{}, fmt.Errorf("failed to unpack metadata: %w", err)
	}

	return domain_common.TokenMetadata{
		Name:     unpacked.Name,
		Symbol:   unpacked.Symbol,
		Decimals: unpacked.Decimals,
	}, nil
}
