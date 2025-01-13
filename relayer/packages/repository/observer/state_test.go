package observer

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	observer_domain "github.com/Datura-ai/relayer/packages/domain/observer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockDB struct {
	data map[string][]byte
}

func NewMockDB() *MockDB {
	return &MockDB{
		data: make(map[string][]byte),
	}
}

func (m *MockDB) Open(path string) error {
	return nil
}

func (m *MockDB) Put(key []byte, value []byte) error {
	m.data[string(key)] = value
	return nil
}

func (m *MockDB) Get(key []byte) ([]byte, error) {
	value, ok := m.data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	return value, nil
}

func (m *MockDB) Delete(key []byte) error {
	delete(m.data, string(key))
	return nil
}

func (m *MockDB) Close() error {
	return nil
}

func TestObserverStateManager(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "observer_state_test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	var manager ObserverStateManager

	// Setup and teardown for each test
	setupTest := func() {
		unconfirmedStore := NewMockDB()
		confirmedStore := NewMockDB()
		manager = NewObserverStateManager(unconfirmedStore, confirmedStore)
	}

	t.Run("TestAddConfirmTransferRequest", func(t *testing.T) {
		t.Run("SucceedsWithNewEntry", func(t *testing.T) {
			setupTest()

			// Arrange
			transfer := createMockTransfer()

			// Act
			err := manager.AddConfirmTransferRequest(transfer)

			// Assert
			assert.NoError(t, err)

			// Verify transfer is in unconfirmed store
			unconfirmedManager := manager.(*observerStateManager)
			storedTransfer, err := unconfirmedManager.getUnconfirmedTransfer(transfer.TransferHash)
			assert.NoError(t, err)
			assert.Equal(t, transfer.TransferHash, storedTransfer.TransferHash)
			assert.Equal(t, transfer.SourceChainID, storedTransfer.SourceChainID)
			assert.Equal(t, transfer.From, storedTransfer.From)
			assert.Equal(t, transfer.DestinationChainID, storedTransfer.DestinationChainID)
			assert.Equal(t, transfer.To, storedTransfer.To)
			assert.Equal(t, transfer.Amount, storedTransfer.Amount)
			assert.Equal(t, transfer.Nonce, storedTransfer.Nonce)
			assert.WithinDuration(t, transfer.Timestamp, storedTransfer.Timestamp, time.Second)
		})

		t.Run("FailsIfTransferAlreadyExists", func(t *testing.T) {
			setupTest()

			// Arrange
			transfer := createMockTransfer()
			err := manager.AddConfirmTransferRequest(transfer)
			require.NoError(t, err)

			// Act
			err = manager.AddConfirmTransferRequest(transfer)

			// Assert
			assert.Error(t, err)
		})
	})

	t.Run("TestMarkTransferAsConfirmed", func(t *testing.T) {
		t.Run("SucceedsWithAddingEntryToConfirmedStore", func(t *testing.T) {
			setupTest()

			// Arrange
			transfer := createMockTransfer()
			err := manager.AddConfirmTransferRequest(transfer)
			require.NoError(t, err)

			// Act
			err = manager.MarkTransferAsConfirmed(transfer.TransferHash)

			// Assert
			// Verify it's in confirmed store
			assert.NoError(t, err)
			confirmedManager := manager.(*observerStateManager)
			confirmedTransfer, err := confirmedManager.getConfirmedTransfer(transfer.TransferHash)
			assert.NoError(t, err)
			assert.Equal(t, transfer.TransferHash, confirmedTransfer.TransferHash)
			assert.Equal(t, transfer.SourceChainID, confirmedTransfer.SourceChainID)
			assert.Equal(t, transfer.From, confirmedTransfer.From)
			assert.Equal(t, transfer.DestinationChainID, confirmedTransfer.DestinationChainID)
			assert.Equal(t, transfer.To, confirmedTransfer.To)
			assert.Equal(t, transfer.Amount, confirmedTransfer.Amount)
			assert.Equal(t, transfer.Nonce, confirmedTransfer.Nonce)
			assert.WithinDuration(t, transfer.Timestamp, confirmedTransfer.Timestamp, time.Second)
		})

		t.Run("SucceedsAndRemovesEntryFromUnconfirmedStore", func(t *testing.T) {
			setupTest()

			// Arrange
			transfer := createMockTransfer()
			err := manager.AddConfirmTransferRequest(transfer)
			require.NoError(t, err)

			// Act
			err = manager.MarkTransferAsConfirmed(transfer.TransferHash)

			// Assert
			assert.NoError(t, err)
			unconfirmedManager := manager.(*observerStateManager)
			// Verify it's no longer in unconfirmed store
			_, err = unconfirmedManager.getUnconfirmedTransfer(transfer.TransferHash)
			assert.Error(t, err)
		})

		t.Run("FailsIfTransferIsNotInUnconfirmedStore", func(t *testing.T) {
			setupTest()

			// Arrange
			transfer := createMockTransfer()

			// Act
			err = manager.MarkTransferAsConfirmed(transfer.TransferHash)

			// Assert
			assert.Error(t, err)
		})

		t.Run("FailsWhenTransferIsAlreadyInConfirmedStore", func(t *testing.T) {
			setupTest()

			// Arrange
			transfer := createMockTransfer()
			err := manager.AddConfirmTransferRequest(transfer)
			require.NoError(t, err)
			err = manager.MarkTransferAsConfirmed(transfer.TransferHash)
			require.NoError(t, err)

			// Act
			err = manager.MarkTransferAsConfirmed(transfer.TransferHash)

			// Assert
			assert.Error(t, err)
		})
	})
}

// Helper function to create a mock TransferData
func createMockTransfer() *observer_domain.TransferData {
	return &observer_domain.TransferData{
		TransferHash:       common.HexToHash("0x1234567890123456789012345678901234567890"),
		SourceChainID:      1,
		From:               common.HexToAddress("0x1234567890123456789012345678901234567890"),
		DestinationChainID: 2,
		To:                 common.HexToAddress("0x0987654321098765432109876543210987654321"),
		Amount:             big.NewInt(1000000000000000000), // 1 ETH in wei
		Nonce:              big.NewInt(1),
		Timestamp:          time.Now(),
	}
}
