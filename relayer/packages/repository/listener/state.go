package listener

import (
	"fmt"

	"github.com/Datura-ai/relayer/packages/domain/common"
	db "go.mills.io/bitcask/v2"
)

type ListenerStateManager interface {
	SaveBroadcastMsg(msg common.BridgeTxData)
	GetBroadcastMsg(msgId string) (common.BridgeTxData, error)
	Close()
}

type listenerStateManager struct {
	store *db.Bitcask
}

func NewListenerStateManager(home string) ListenerStateManager {
	option := []db.Option{
		db.WithMaxKeySize(1024),
		db.WithMaxValueSize(1024 * 1024),
		db.WithAutoRecovery(true),
		db.WithDirMode(0755),
		db.WithFileMode(0644),
	}
	store, _ := db.Open(fmt.Sprintf("%s%s", home, "/listener_db"), option...)
	//defer store.Close()
	return &listenerStateManager{
		store: store,
	}
}

func (l *listenerStateManager) SaveBroadcastMsg(msg common.BridgeTxData) {
	msgBytes := msg.ToBytes()
	msgId := msg.GetMsgId()
	l.store.Put(db.Key(msgId), msgBytes)
}

func (l *listenerStateManager) GetBroadcastMsg(msgId string) (common.BridgeTxData, error) {
	msgBytes, err := l.store.Get(db.Key(msgId))
	if err != nil {
		return common.BridgeTxData{}, err
	}
	bridgeTxData := common.BridgeTxData{}
	err = bridgeTxData.FromBytes(msgBytes)
	if err != nil {
		return common.BridgeTxData{}, err
	}
	return bridgeTxData, nil
}
func (l *listenerStateManager) Close() {
	l.store.Close()
}
