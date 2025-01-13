package observer

import (
	"encoding/json"
	"fmt"

	observer_domain "github.com/Datura-ai/relayer/packages/domain/observer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

type ObserverStateManager interface {
	AddConfirmTransferRequest(transfer *observer_domain.TransferData) error
	MarkTransferAsConfirmed(transferHash common.Hash) error
}

type DBInterface interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
}

type observerStateManager struct {
	unconfirmedStore DBInterface
	confirmedStore   DBInterface
}

func NewObserverStateManager(unconfirmedStore, confirmedStore DBInterface) ObserverStateManager {
	return &observerStateManager{
		unconfirmedStore: unconfirmedStore,
		confirmedStore:   confirmedStore,
	}
}

func (s *observerStateManager) AddConfirmTransferRequest(transfer *observer_domain.TransferData) error {
	existingTransfer, err := s.getUnconfirmedTransfer(transfer.TransferHash)
	if err == nil && existingTransfer != nil {
		log.Error().Str("transferHash", transfer.TransferHash.Hex()).Msg("❌ Transfer already exists")
		return fmt.Errorf("transfer with hash %s already exists", transfer.TransferHash.Hex())
	}
	err = s.saveUnconfirmedTransfer(transfer)
	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to save unconfirmed transfer")
	} else {
		log.Info().Str("transferHash", transfer.TransferHash.Hex()).Msg("✅ Successfully saved unconfirmed transfer")
	}
	return err
}

func (s *observerStateManager) MarkTransferAsConfirmed(transferHash common.Hash) error {
	transfer, err := s.getUnconfirmedTransfer(transferHash)
	if err != nil {
		log.Error().Err(err).Str("transferHash", transferHash.Hex()).Msg("❌ Failed to get unconfirmed transfer")
		return err
	}

	err = s.saveConfirmedTransfer(transfer)
	if err != nil {
		log.Error().Err(err).Str("transferHash", transferHash.Hex()).Msg("❌ Failed to save confirmed transfer")
		return err
	}

	err = s.unconfirmedStore.Delete(transferHash.Bytes())
	if err != nil {
		log.Error().Err(err).Str("transferHash", transferHash.Hex()).Msg("❌ Failed to delete unconfirmed transfer")
		return err
	}

	log.Info().Str("transferHash", transferHash.Hex()).Msg("✅ Successfully marked transfer as confirmed")
	return nil
}

func (s *observerStateManager) saveUnconfirmedTransfer(transfer *observer_domain.TransferData) error {
	return s.saveTransfer(s.unconfirmedStore, transfer)
}

func (s *observerStateManager) saveConfirmedTransfer(transfer *observer_domain.TransferData) error {
	return s.saveTransfer(s.confirmedStore, transfer)
}

func (s *observerStateManager) saveTransfer(store DBInterface, transfer *observer_domain.TransferData) error {
	key := transfer.TransferHash.Bytes()
	value, err := json.Marshal(transfer)
	if err != nil {
		return err
	}
	return store.Put(key, value)
}

func (s *observerStateManager) getUnconfirmedTransfer(transferHash common.Hash) (*observer_domain.TransferData, error) {
	return s.getTransfer(s.unconfirmedStore, transferHash)
}

func (s *observerStateManager) getConfirmedTransfer(transferHash common.Hash) (*observer_domain.TransferData, error) {
	return s.getTransfer(s.confirmedStore, transferHash)
}

func (s *observerStateManager) getTransfer(store DBInterface, transferHash common.Hash) (*observer_domain.TransferData, error) {
	key := transferHash.Bytes()
	value, err := store.Get(key)
	if err != nil {
		return nil, err
	}

	var transfer observer_domain.TransferData
	err = json.Unmarshal(value, &transfer)
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}
