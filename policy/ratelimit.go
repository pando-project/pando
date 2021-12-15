package policy

import (
	"Pando/internal/registry"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/time/rate"
	"math"
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
	gateLimiter         *rate.Limiter
	whitelistLimiter    *rate.Limiter
	unregisteredLimiter *rate.Limiter

	registeredLimiter map[int]*rate.Limiter

	peers map[peer.ID]*rate.Limiter
	mu    *sync.RWMutex

	config LimiterConfig
}

var tokenRateZeroError = fmt.Errorf("token rate is zero")

func NewLimiter(c LimiterConfig) (*Limiter, error) {
	if !rateIsValid(c.TotalRate) || !rateIsValid(float64(c.TotalBurst)) {
		return nil, fmt.Errorf("total rate or total burst is zero")
	}
	return &Limiter{
		gateLimiter: rate.NewLimiter(rate.Limit(c.TotalRate), c.TotalBurst),
		mu:          &sync.RWMutex{},
		peers:       make(map[peer.ID]*rate.Limiter),
		config:      c,
	}, nil
}

func (i *Limiter) Allow() bool {
	return i.gateLimiter.Allow()
}

func (i *Limiter) GateLimiter() *rate.Limiter {
	return i.gateLimiter
}

func (i *Limiter) UnregisteredLimiter(baseTokenRate float64) (*rate.Limiter, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.unregisteredLimiter != nil {
		return i.unregisteredLimiter, nil
	}

	tokenRate := math.Ceil(0.1 * baseTokenRate)
	if !rateIsValid(tokenRate) {
		return nil, tokenRateZeroError
	}
	i.unregisteredLimiter = rate.NewLimiter(rate.Limit(tokenRate), int(tokenRate))

	return i.unregisteredLimiter, nil
}

func (i *Limiter) WhitelistLimiter(baseTokenRate float64) (*rate.Limiter, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.whitelistLimiter != nil {
		return i.whitelistLimiter, nil
	}

	tokenRate := math.Ceil(0.6 * baseTokenRate)
	if !rateIsValid(tokenRate) {
		return nil, tokenRateZeroError
	}
	i.whitelistLimiter = rate.NewLimiter(rate.Limit(tokenRate), int(tokenRate))

	return i.whitelistLimiter, nil
}

func (i *Limiter) RegisteredLimiter(baseTokenRate float64, accountLevel int, levelCount int) (*rate.Limiter, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.registeredLimiter == nil {
		i.registeredLimiter = make(map[int]*rate.Limiter)
	}

	limiter, exists := i.registeredLimiter[accountLevel]
	if !exists {
		if accountLevel < 1 || levelCount < 1 || accountLevel > levelCount {
			return nil, fmt.Errorf("accountLevel or levelCount is invalid, given: accountLevel=%v, levelCount=%v",
				accountLevel, levelCount)
		}
		weight := float64(accountLevel) / float64(levelCount)
		tokenRate := math.Ceil(0.4 * weight * baseTokenRate)
		if !rateIsValid(tokenRate) {
			return nil, tokenRateZeroError
		}
		limiter = rate.NewLimiter(rate.Limit(tokenRate), int(tokenRate))
		i.registeredLimiter[accountLevel] = limiter
	}

	return limiter, nil
}

// AddPeerLimiter append a new rate-limiter for a peer into the peers array of Limiter
func (i *Limiter) AddPeerLimiter(peerID peer.ID, limiter *rate.Limiter) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.peers[peerID] = limiter

	return limiter
}

// PeerLimiter return a rate-limiter for specified peer if exists, or return nil
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

func rateIsValid(tokenRate float64) bool {
	return tokenRate > 0
}
