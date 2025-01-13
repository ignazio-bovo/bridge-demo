package discovery

import (
	"fmt"
	"sort"
	"time"

	repository "github.com/Datura-ai/relayer/packages/repository/tss"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerDiscovery struct {
	host       host.Host
	stateMgr   repository.TssStateManager
	keygenChan *chan []peer.ID
}

func NewPeerDiscovery(host host.Host, stateMgr repository.TssStateManager) *PeerDiscovery {
	return &PeerDiscovery{
		host:     host,
		stateMgr: stateMgr,
	}
}

func (pd *PeerDiscovery) Start() error {
	for {
		peers := pd.host.Network().Peers()
		peers = append(peers, pd.host.ID())
		sort.Slice(peers, func(i, j int) bool {
			return peers[i].String() < peers[j].String()
		})
		*pd.keygenChan <- peers
		time.Sleep(50 * time.Second)
	}
}

func (pd *PeerDiscovery) GetNewPeerChannel() (chan []peer.ID, error) {
	if pd.keygenChan == nil {
		return nil, fmt.Errorf("newPeerChannel is not initialized")
	}
	return *pd.keygenChan, nil
}

func (pd *PeerDiscovery) SetNewKeygenChannel(keygenChan *chan []peer.ID) {
	pd.keygenChan = keygenChan
}

func slicesEqual(a, b []peer.ID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
