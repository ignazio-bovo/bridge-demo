package common

import (
	"time"
)

type TssConfig struct {
	// Party Timeout defines how long do we wait for the party to form
	PartyTimeout time.Duration
	// KeyresharingTimeout defines how long do we wait the key resharing parties to pass messages along
	EcdsaTimeout time.Duration
	// KeygenTimeout defines how long do we wait the keygen parties to pass messages along
	KeygenTimeout time.Duration
	// KeyresharingTimeout defines how long do we wait the key resharing parties to pass messages along
	KeyresharingTimeout time.Duration
	// KeysignTimeout defines how long do we wait keysign
	KeysignTimeout time.Duration
	// Pre-parameter define the pre-parameter generations timeout
	PreParamTimeout time.Duration
	// enable the tss monitor
	EnableMonitor bool
}

const KEYGEN_TSS_HANDLE_ID = uint64(0)
const KEYGEN_MAX_COUNT = 2
