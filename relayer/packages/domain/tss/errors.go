package tss

import "errors"

var (
	ErrChannelClosed error = errors.New("error channel closed fail to start local party")
	ErrExitSignal    error = errors.New("received exit signal")
)
