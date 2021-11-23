package registry

import "errors"

var (
	ErrInProgress  = errors.New("discovery already in progress")
	ErrNotAllowed  = errors.New("provider not allowed by policy")
	ErrNoDiscovery = errors.New("discovery not available")
	ErrNotTrusted  = errors.New("provider not trusted to register without on-chain verification")
	ErrWrongWeight = errors.New("provider should not have weight before evaluating")
	ErrNotVerified = errors.New("provider cannot be verified")
	ErrTooSoon     = errors.New("not enough time since previous discovery")
)
