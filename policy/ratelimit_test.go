package policy

import (
	"Pando/config"
	"Pando/internal/lotus"
	"Pando/internal/registry"
	"fmt"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"math"
	"testing"
)

var (
	Bandwidth        = 100.0
	SingleDAGSize    = 2.0
	BaseTokenRate    = math.Ceil(0.8 * Bandwidth / SingleDAGSize)
	RegistryInstance = newRegistry()

	testLimiter  *Limiter
	resetLimiter = func() {
		testLimiter.unregisteredLimiter = nil
		testLimiter.whitelistLimiter = nil
		testLimiter.registeredLimiter = nil
	}
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
	nilLimiter, err := NewLimiter(LimiterConfig{
		TotalRate:  0,
		TotalBurst: 0,
	})
	assert.Nil(t, nilLimiter)
	assert.NotNil(t, err)

	limiter, err := NewLimiter(LimiterConfig{
		TotalRate:  BaseTokenRate,
		TotalBurst: int(BaseTokenRate),
		Registry:   RegistryInstance,
	})
	if err != nil {
		t.Error(err)
	}

	gateLimiterLimit := limiter.GateLimiter().Limit()
	assert.Equal(t, rate.Limit(BaseTokenRate), gateLimiterLimit)

	gateLimiterBurst := limiter.GateLimiter().Burst()
	assert.Equal(t, int(BaseTokenRate), gateLimiterBurst)

	testLimiter = limiter
}

func TestLimiter_UnregisteredLimiter(t *testing.T) {
	// m * baseRate = 0.1 * 0.8 * 100 / 2 = 4
	unregisteredPeerRate := rate.Limit(4.0)
	limiter, err := testLimiter.UnregisteredLimiter(BaseTokenRate)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, unregisteredPeerRate, limiter.Limit())

	limiterExactlyExists, err := testLimiter.UnregisteredLimiter(BaseTokenRate)
	if err != nil {
		t.Error(err)
	}
	assert.Exactly(t, limiter, limiterExactlyExists)

	resetLimiter()
	tokenRateZeroLimiter, err := testLimiter.UnregisteredLimiter(0)
	assert.Nil(t, tokenRateZeroLimiter)
	assert.NotNil(t, err)
}

func TestLimiter_WhitelistLimiter(t *testing.T) {
	// m * baseRate = 0.5 * 0.8 * 100 / 2 = 20
	whitelistPeerRate := rate.Limit(20.0)
	limiter, err := testLimiter.WhitelistLimiter(BaseTokenRate)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, whitelistPeerRate, limiter.Limit())

	limiterExactlyExists, err := testLimiter.WhitelistLimiter(BaseTokenRate)
	if err != nil {
		t.Error(err)
	}
	assert.Exactly(t, limiter, limiterExactlyExists)

	testLimiter.whitelistLimiter = nil
	tokenRateZeroLimiter, err := testLimiter.WhitelistLimiter(0)
	assert.Nil(t, tokenRateZeroLimiter)
	assert.NotNil(t, err)
}

func TestLimiter_RegisteredLimiter(t *testing.T) {
	const levelCount = 5

	// m * baseRate = 0.4 * weight * baseRate = 0.4 * 1 / 5 * 0.8 * 100 / 2 = 3.2
	// math.Ceil(3.2) = 4
	accountLevel := 1
	limiter, err := testLimiter.RegisteredLimiter(BaseTokenRate, accountLevel, levelCount)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, rate.Limit(4), limiter.Limit())

	nilLimiter, err := testLimiter.RegisteredLimiter(BaseTokenRate, 0, levelCount)
	assert.Nil(t, nilLimiter)
	assert.NotNil(t, err)

	resetLimiter()
	nilLimiter, err = testLimiter.RegisteredLimiter(BaseTokenRate, accountLevel, 0)
	assert.Nil(t, nilLimiter)
	assert.NotNil(t, err)

	resetLimiter()
	nilLimiter, err = testLimiter.RegisteredLimiter(BaseTokenRate, 6, 5)
	assert.Nil(t, nilLimiter)
	assert.NotNil(t, err)

	resetLimiter()
	tokenRateZeroLimiter, err := testLimiter.RegisteredLimiter(0, accountLevel, levelCount)
	assert.Nil(t, tokenRateZeroLimiter)
	assert.NotNil(t, err)
}

func TestLimiter_Allow(t *testing.T) {
	limitBackup := testLimiter.GateLimiter().Limit()
	burstBackup := testLimiter.GateLimiter().Burst()

	assert.Equal(t, true, testLimiter.Allow())

	testLimiter.GateLimiter().SetLimit(0)
	testLimiter.GateLimiter().SetBurst(0)
	assert.Equal(t, false, testLimiter.Allow())

	testLimiter.GateLimiter().SetLimit(limitBackup)
	testLimiter.GateLimiter().SetBurst(burstBackup)
}

func TestLimiter_AddPeerLimiter(t *testing.T) {
	const testPeerID = "12D3KooWJfFoQ2D1nukmG84DEh6gGEEE49yG6rPCdHoCqhF7YyL1"

	peerID, err := peer.Decode(testPeerID)
	if err != nil {
		t.Errorf("decode peer id failed, error: %v", err)
	}
	limiter := rate.NewLimiter(1, 1)
	addedLimiter := testLimiter.AddPeerLimiter(peerID, limiter)
	assert.Equal(t, limiter, addedLimiter)
}

func TestLimiter_PeerLimiter(t *testing.T) {
	peerID := peer.NewPeerRecord().PeerID
	limiter := rate.NewLimiter(1, 1)

	addedLimiter := testLimiter.AddPeerLimiter(peerID, limiter)
	peerLimiter := testLimiter.PeerLimiter(peerID)
	assert.Equal(t, addedLimiter, peerLimiter)

	assert.Nil(t, testLimiter.PeerLimiter("呵呵"))
}

func TestLimiter_Config(t *testing.T) {
	limiterConf := testLimiter.Config()
	assert.Equal(t, BaseTokenRate, limiterConf.TotalRate)
	assert.Equal(t, int(BaseTokenRate), limiterConf.TotalBurst)
	assert.Equal(t, RegistryInstance, limiterConf.Registry)
}
