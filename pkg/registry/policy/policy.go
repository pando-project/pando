package policy

import (
	"errors"
	"fmt"
	"github.com/pando-project/pando/pkg/option"
	"sync"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Policy struct {
	allow       bool
	except      map[peer.ID]struct{}
	trust       bool
	trustExcept map[peer.ID]struct{}
	rwmutex     sync.RWMutex
}

func New(cfg option.Policy) (*Policy, error) {
	policy := &Policy{
		allow: cfg.Allow,
		trust: cfg.Trust,
	}

	var err error
	policy.except, err = getExceptPeerIDs(cfg.Except)
	if err != nil {
		return nil, fmt.Errorf("cannot read except list: %s", err)
	}

	// Error if no peers are allowed
	if !policy.allow && len(policy.except) == 0 {
		return nil, errors.New("policy does not allow any providers")
	}

	policy.trustExcept, err = getExceptPeerIDs(cfg.TrustExcept)
	if err != nil {
		return nil, fmt.Errorf("cannot read trust except list: %s", err)
	}

	return policy, nil
}

func getExceptPeerIDs(excepts []string) (map[peer.ID]struct{}, error) {
	if len(excepts) == 0 {
		return nil, nil
	}

	exceptIDs := make(map[peer.ID]struct{}, len(excepts))
	for _, except := range excepts {
		excPeerID, err := peer.Decode(except)
		if err != nil {
			return nil, fmt.Errorf("error decoding account id %q: %s", except, err)
		}
		exceptIDs[excPeerID] = struct{}{}
	}
	return exceptIDs, nil
}

// Allowed returns true if the policy allows the peer to index content.  This
// check does not check whether the peer is trusted. An allowed peer must still
// be verified.
func (p *Policy) Allowed(peerID peer.ID) bool {
	p.rwmutex.RLock()
	defer p.rwmutex.RUnlock()
	return p.allowed(peerID)
}

func (p *Policy) allowed(peerID peer.ID) bool {
	_, ok := p.except[peerID]
	if p.allow {
		return !ok
	}
	return ok
}

// Trusted returns true if the peer is explicitly trusted.  A trusted peer is
// allowed to register without requiring verification.
func (p *Policy) Trusted(peerID peer.ID) bool {
	p.rwmutex.RLock()
	defer p.rwmutex.RUnlock()
	return p.trusted(peerID)
}

func (p *Policy) trusted(peerID peer.ID) bool {
	_, ok := p.trustExcept[peerID]
	if p.trust {
		return !ok
	}
	return ok
}

// Check returns whether the two bool values.  The fisrt is true if the peer is
// allowed.  The second is true if the peer is allowed and is trusted (does not
// require verification).
func (p *Policy) Check(peerID peer.ID) (bool, bool) {
	p.rwmutex.RLock()
	defer p.rwmutex.RUnlock()

	if !p.allowed(peerID) {
		return false, false
	}

	if !p.trusted(peerID) {
		return true, false
	}

	return true, true
}
