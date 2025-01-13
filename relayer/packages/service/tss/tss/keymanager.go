package tss

type TssKeyManager interface {
	RemoteSign(msg []byte, poolPubKey string, msgId string, isLastTssInput bool) ([]byte, []byte, string, error)
}
