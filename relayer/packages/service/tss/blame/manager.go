package blame

import (
	"sync"

	btss "github.com/bnb-chain/tss-lib/tss"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	logger            zerolog.Logger
	blame             *Blame
	lastUnicastPeer   map[string][]peer.ID
	shareMgr          *ShareMgr
	roundMgr          *RoundMgr
	partyInfo         *PartyInfo
	partyIDtoP2PID    map[string]peer.ID
	newPartyIDtoP2PID map[string]peer.ID
	lastMsgLocker     *sync.RWMutex
	lastMsg           btss.Message
	acceptedShares    map[RoundInfo][]string
	acceptShareLocker *sync.Mutex
	localPartyID      string
	newLocalPartyID   string
}

func NewBlameManager() *Manager {
	blame := NewBlame("", nil)
	return &Manager{
		logger:            log.With().Str("module", "blame_manager").Logger(),
		partyInfo:         nil,
		partyIDtoP2PID:    make(map[string]peer.ID),
		newPartyIDtoP2PID: make(map[string]peer.ID),
		lastUnicastPeer:   make(map[string][]peer.ID),
		shareMgr:          NewTssShareMgr(),
		roundMgr:          NewTssRoundMgr(),
		blame:             &blame,
		lastMsgLocker:     &sync.RWMutex{},
		acceptedShares:    make(map[RoundInfo][]string),
		acceptShareLocker: &sync.Mutex{},
	}
}

func (m *Manager) GetBlame() *Blame {
	return m.blame
}

func (m *Manager) GetShareMgr() *ShareMgr {
	return m.shareMgr
}

func (m *Manager) GetRoundMgr() *RoundMgr {
	return m.roundMgr
}

func (m *Manager) UpdateAcceptShare(round RoundInfo, id string) {
	m.acceptShareLocker.Lock()
	defer m.acceptShareLocker.Unlock()
	partyList, ok := m.acceptedShares[round]
	if !ok {
		partyList := []string{id}
		m.acceptedShares[round] = partyList
		return
	}
	partyList = append(partyList, id)
	m.acceptedShares[round] = partyList
}

func (m *Manager) CheckMsgDuplication(round RoundInfo, id string) bool {
	m.acceptShareLocker.Lock()
	defer m.acceptShareLocker.Unlock()
	partyList, ok := m.acceptedShares[round]
	if ok {
		for _, el := range partyList {
			if el == id {
				return true
			}
		}
	}
	return false
}

func (m *Manager) SetLastMsg(lastMsg btss.Message) {
	m.lastMsgLocker.Lock()
	defer m.lastMsgLocker.Unlock()
	m.lastMsg = lastMsg
}

func (m *Manager) GetLastMsg() btss.Message {
	m.lastMsgLocker.RLock()
	defer m.lastMsgLocker.RUnlock()
	return m.lastMsg
}

func (m *Manager) SetPartyInfo(partyMap, newPartyMap *sync.Map, partyIDMap map[string]*btss.PartyID, isEcdsa bool) {
	partyInfo := &PartyInfo{
		PartyMap:    partyMap,
		NewPartyMap: newPartyMap,
		PartyIDMap:  partyIDMap,
	}
	m.partyInfo = partyInfo

	if !isEcdsa {
		if !isMapEmpty(m.partyInfo.PartyMap) {
			var localParty btss.Party
			m.partyInfo.PartyMap.Range(func(key, value interface{}) bool {
				localParty = value.(btss.Party)
				return false
			})
			m.localPartyID = localParty.PartyID().Id
		}

		if !isMapEmpty(m.partyInfo.NewPartyMap) {
			var newLocalPartyID btss.Party
			m.partyInfo.NewPartyMap.Range(func(key, value interface{}) bool {
				newLocalPartyID = value.(btss.Party)
				return false
			})
			m.newLocalPartyID = newLocalPartyID.PartyID().Id
		}
	}
}

func (m *Manager) SetLastUnicastPeer(peerID peer.ID, roundInfo string) {
	m.lastMsgLocker.Lock()
	defer m.lastMsgLocker.Unlock()
	l, ok := m.lastUnicastPeer[roundInfo]
	if !ok {
		peerList := []peer.ID{peerID}
		m.lastUnicastPeer[roundInfo] = peerList
	} else {
		l = append(l, peerID)
		m.lastUnicastPeer[roundInfo] = l
	}
}

func isMapEmpty(m *sync.Map) bool {
	empty := true
	if m == nil {
		return empty
	}

	m.Range(func(key, value interface{}) bool {
		empty = false
		return false
	})
	return empty
}
