package blame

import "github.com/Datura-ai/relayer/packages/service/tss/conversion"

// GetBlamePubKeysInList returns the nodes public key who are in the peer list
func (m *Manager) getBlamePubKeysInList(peers []string) ([]string, error) {
	var partiesInList []string
	// we convert nodes (in the peers list) P2PID to public key
	for partyID, p2pID := range m.partyIDtoP2PID {
		for _, el := range peers {
			if el == p2pID.String() {
				partiesInList = append(partiesInList, partyID)
			}
		}
	}
	for partyID, p2pID := range m.newPartyIDtoP2PID {
		for _, el := range peers {
			if el == p2pID.String() {
				partiesInList = append(partiesInList, partyID)
			}
		}
	}

	localPartyInfo := m.partyInfo
	partyIDMap := localPartyInfo.PartyIDMap
	blamePubKeys, err := conversion.AccPubKeysFromPartyIDs(partiesInList, partyIDMap)
	if err != nil {
		return nil, err
	}

	return blamePubKeys, nil
}

func (m *Manager) getBlamePubKeysNotInList(peers []string) ([]string, error) {
	partiesNotInList := make(map[string]struct{})
	// we convert nodes (NOT in the peers list) P2PID to public key
	for partyID, p2pID := range m.partyIDtoP2PID {
		if m.localPartyID == partyID {
			continue
		}
		found := false
		for _, each := range peers {
			if p2pID.String() == each {
				found = true
				break
			}
		}
		if !found {
			partiesNotInList[partyID] = struct{}{}
		}
	}

	for partyID, p2pID := range m.newPartyIDtoP2PID {
		if m.localPartyID == partyID {
			continue
		}
		found := false
		for _, each := range peers {
			if p2pID.String() == each {
				found = true
				break
			}
		}
		_, ok := partiesNotInList[partyID]
		if found && ok {
			delete(partiesNotInList, partyID)
		} else if !found && !ok {
			partiesNotInList[partyID] = struct{}{}
		}
	}

	var partyNotInListArray []string
	for partyID := range partiesNotInList {
		partyNotInListArray = append(partyNotInListArray, partyID)
	}

	partyIDMap := m.partyInfo.PartyIDMap
	blamePubKeys, err := conversion.AccPubKeysFromPartyIDs(partyNotInListArray, partyIDMap)
	if err != nil {
		return nil, err
	}

	return blamePubKeys, nil
}

// GetBlamePubKeysNotInList returns the nodes public key who are not in the peer list
func (m *Manager) GetBlamePubKeysLists(peer []string) ([]string, []string, error) {
	inList, err := m.getBlamePubKeysInList(peer)
	if err != nil {
		return nil, nil, err
	}

	notInlist, err := m.getBlamePubKeysNotInList(peer)
	if err != nil {
		return nil, nil, err
	}

	return inList, notInlist, err
}
