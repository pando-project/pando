package policy

import (
	"Pando/internal/registry"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/time/rate"
	"sync"
)

type LimiterConfig struct {
	Registry   *registry.Registry
	TotalRate  float64
	TotalBurst int

	PeerRate  float64
	PeerBurst int

	BaseTokenRate float64
}

type Limiter struct {
	control *rate.Limiter

	peers map[peer.ID]*rate.Limiter
	mu    *sync.RWMutex

	config LimiterConfig
}

func NewLimiter(c LimiterConfig) *Limiter {
	return &Limiter{
		control: rate.NewLimiter(rate.Limit(c.TotalRate), c.TotalBurst),
		mu:      &sync.RWMutex{},
		peers:   make(map[peer.ID]*rate.Limiter),

		config: c,
	}
}

func (i *Limiter) Allow() bool {
	return i.control.Allow()
}

func (i *Limiter) AddPeerLimiter(peerID peer.ID, peerRate float64, peerBurst int) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(rate.Limit(peerRate), peerBurst)

	i.peers[peerID] = limiter

	return limiter
}

// PeerLimiter return a rate-limiter for specified account if exists, or return nil
func (i *Limiter) PeerLimiter(peerID peer.ID) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.peers[peerID]

	if !exists {
		i.mu.Unlock()
		return nil
	}

	i.mu.Unlock()

	return limiter
}

func (i *Limiter) Config() LimiterConfig {
	return i.config
}
