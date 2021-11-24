package policy

import (
	"Pando/config"
	"Pando/internal/lotus"
	"Pando/internal/registry"
	"fmt"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"math"
	"testing"
)

var (
	Bandwidth     = 100.0
	SingleDAGSize = 2.0
	BaseTokenRate = math.Ceil(0.8 * Bandwidth / SingleDAGSize)
	testLimiter   *Limiter
)

func newRegistry() *registry.Registry {
	dstore, err := leveldb.NewDatastore("/tmp", nil)
	lotusDiscoverer, err := lotus.NewDiscoverer("https://api.chain.love")
	registryInstance, err := registry.NewRegistry(
		&config.Discovery{Policy: config.Policy{Allow: true}},
		&config.AccountLevel{Threshold: []int{1, 10}}, dstore, lotusDiscoverer)
	if err != nil {
		panic(fmt.Errorf("new registry failed, error: %v", err))
	}
	return registryInstance
}

func TestNewLimiter(t *testing.T) {
	limiterConf := LimiterConfig{
		TotalRate:  BaseTokenRate,
		TotalBurst: int(BaseTokenRate),
	}
	limiter := NewLimiter(limiterConf)

	gateLimiterLimit := limiter.GateLimiter().Limit()
	assert.Equal(t, rate.Limit(BaseTokenRate), gateLimiterLimit)

	gateLimiterBurst := limiter.GateLimiter().Burst()
	assert.Equal(t, int(BaseTokenRate), gateLimiterBurst)

	testLimiter = limiter
}

func TestLimiter_UnregisteredLimiter(t *testing.T) {
	// m * baseRate = 0.1 * 0.8 * 100 / 2 = 4
	unregisteredPeerRate := rate.Limit(4.0)
	limiter := testLimiter.UnregisteredLimiter(BaseTokenRate)
	assert.Equal(t, unregisteredPeerRate, limiter.Limit())
}

func TestLimiter_WhitelistLimiter(t *testing.T) {
	// m * baseRate = 0.5 * 0.8 * 100 / 2 = 20
	whitelistPeerRate := rate.Limit(20.0)
	limiter := testLimiter.WhitelistLimiter(BaseTokenRate)
	assert.Equal(t, whitelistPeerRate, limiter.Limit())
}

func TestLimiter_RegisteredLimiter(t *testing.T) {
	const levelCount = 5

	// m * baseRate = 0.4 * weight * baseRate = 0.4 * 1 / 5 * 0.8 * 100 / 2 = 3.2
	// math.Ceil(3.2) = 4
	accountLevel1 := 1
	limiter := testLimiter.RegisteredLimiter(BaseTokenRate, accountLevel1, levelCount)
	assert.Equal(t, rate.Limit(4), limiter.Limit())
}
