package tss

func (t *TssServer) StartPeerDiscovery() error {
	return t.discovery.Start()
}
