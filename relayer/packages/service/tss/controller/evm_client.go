package controller

import (
	"context"
	"time"

	"github.com/Datura-ai/relayer/packages/domain/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EVMClient interface {
	GetClient() *ethclient.Client
	BroadcastTx(rawTx []byte) (string, error)
}

type client struct {
	ethClient *ethclient.Client
}

func NewEVMClient(hostUrl string, tx *common.BridgeTxData) (EVMClient, error) {
	ethClient, err := ethclient.Dial(hostUrl)
	if err != nil {
		return nil, err
	}
	return &client{
		ethClient: ethClient,
	}, nil
}

func (c *client) GetClient() *ethclient.Client {
	return c.ethClient
}

func (c *client) BroadcastTx(rawTx []byte) (string, error) {
	tx := &etypes.Transaction{}
	if err := tx.UnmarshalJSON(rawTx); err != nil {
		return "", err
	}
	ctx, cancel := c.getContext()
	defer cancel()

	err := c.ethClient.SendTransaction(ctx, tx)
	if err != nil {
		return "", err
	}
	txID := tx.Hash().String()
	return txID, nil
}

func (c *client) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
