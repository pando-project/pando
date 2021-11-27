package policy

import (
	"Pando/config"
	"Pando/internal/lotus"
	"Pando/internal/registry"
	"fmt"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
	"math"
	"testing"
)

var (
	bandwidth        = 100.0
	singleDAGSize    = 2.0
	baseTokenRate    = math.Ceil(0.8 * bandwidth / singleDAGSize)
	registryInstance *registry.Registry
	testLimiter      *Limiter
)

func init() {
	// init registryInstance
	dstore, err := leveldb.NewDatastore("/tmp", nil)
	if err != nil {
		panic(fmt.Errorf("create datastore failed, error: %v", err))
	}
	mutexDstore := sync.MutexWrap(dstore)
	lotusDiscoverer, err := lotus.NewDiscoverer("https://api.chain.love")
	if err != nil {
		panic(fmt.Errorf("create discovery failed, error: %v", err))
	}
	registryInstance, err = registry.NewRegistry(
		&config.Discovery{Policy: config.Policy{Allow: true}},
		&config.AccountLevel{Threshold: []int{1, 10}}, mutexDstore, lotusDiscoverer)
	if err != nil {
		panic(fmt.Errorf("new registry failed, error: %v", err))
	}

	// init testLimiter
	testLimiter, err = NewLimiter(LimiterConfig{
		TotalRate:  baseTokenRate,
		TotalBurst: int(baseTokenRate),
		Registry:   registryInstance,
	})
	if err != nil {
		panic(fmt.Errorf("new test limiter failed, error: %v", err))
	}
}

func TestNewLimiter(t *testing.T) {
	Convey("TestNewLimiter", t, func() {
		Convey("return nil when TotalRate is 0", func() {
			limiter, err := NewLimiter(LimiterConfig{
				TotalRate:  0,
				TotalBurst: 1,
			})
			So(limiter, ShouldBeNil)
			So(err, ShouldResemble, fmt.Errorf("total rate or total burst is zero"))
		})

		Convey("return nil when TotalBurst is 0", func() {
			limiter, err := NewLimiter(LimiterConfig{
				TotalRate:  1,
				TotalBurst: 0,
			})
			So(limiter, ShouldBeNil)
			So(err, ShouldResemble, fmt.Errorf("total rate or total burst is zero"))
		})

		Convey("create limiter", func() {
			limiter, err := NewLimiter(LimiterConfig{
				TotalRate:  baseTokenRate,
				TotalBurst: int(baseTokenRate),
				Registry:   registryInstance,
			})
			So(limiter, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("limit of gateLimiter should be baseTokenRate", func() {
				So(limiter.GateLimiter().Limit(), ShouldEqual, rate.Limit(baseTokenRate))
			})

			Convey("burst of gateLimiter should be baseTokenRate", func() {
				So(limiter.GateLimiter().Burst(), ShouldEqual, int(baseTokenRate))
			})
		})
	})
}

func TestLimiter_UnregisteredLimiter(t *testing.T) {
	// m * baseRate = 0.1 * 0.8 * 100 / 2 = 4
	Convey("TestUnregisteredLimiter", t, func() {
		testLimiterCopy := *testLimiter
		stubTestLimiter := ApplyGlobalVar(&testLimiter, &testLimiterCopy)

		Convey("return a new unregistered limiter if not exists", func() {
			unregisteredLimiter, err := testLimiter.UnregisteredLimiter(baseTokenRate)
			So(unregisteredLimiter, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("limit of unregistered limiter = 4.0 when base rate = 40", func() {
				So(unregisteredLimiter.Limit(), ShouldEqual, rate.Limit(4.0))
			})

			Convey("return a limiter from *Limiter.unregisteredLimiter when it's exist,"+
				"in this case, it should be exactly the limiter created above", func() {
				limiterExactlyExists, err := testLimiter.UnregisteredLimiter(baseTokenRate)
				So(err, ShouldBeNil)
				So(limiterExactlyExists, ShouldEqual, unregisteredLimiter)
			})

			stubTestLimiter.Reset()
		})

		Convey("return nil and error when base rate is 0", func() {
			tokenRateZeroLimiter, err := testLimiter.UnregisteredLimiter(0)
			So(tokenRateZeroLimiter, ShouldBeNil)
			So(err, ShouldEqual, tokenRateZeroError)
		})
	})
}

func TestLimiter_WhitelistLimiter(t *testing.T) {
	// m * baseRate = 0.5 * 0.8 * 100 / 2 = 20
	Convey("TestWhiteLimiter", t, func() {
		testLimiterCopy := *testLimiter
		stubTestLimiter := ApplyGlobalVar(&testLimiter, &testLimiterCopy)

		Convey("return a new whitelist limiter if not exist", func() {
			whitelistLimiter, err := testLimiter.WhitelistLimiter(baseTokenRate)
			So(whitelistLimiter, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("limit of whitelist limiter = 20 when base rate = 40", func() {
				So(whitelistLimiter.Limit(), ShouldEqual, rate.Limit(20))
			})

			Convey("return a limiter from *Limiter.whitelistLimiter when it's exist,"+
				"in this case, it should be exactly the limiter created above", func() {
				limiterExactlyExists, err := testLimiter.WhitelistLimiter(baseTokenRate)
				So(err, ShouldBeNil)
				So(limiterExactlyExists, ShouldEqual, whitelistLimiter)
			})

			stubTestLimiter.Reset()
		})

		Convey("return nil and error when base rate is 0", func() {
			tokenRateZeroLimiter, err := testLimiter.WhitelistLimiter(0)
			So(tokenRateZeroLimiter, ShouldBeNil)
			So(err, ShouldEqual, tokenRateZeroError)
		})
	})
}

func TestLimiter_RegisteredLimiter(t *testing.T) {
	//m * baseRate = 0.4 * weight * baseRate = 0.4 * 1 / 5 * 0.8 * 100 / 2 = 3.2
	//math.Ceil(3.2) = 4

	Convey("TestRegisteredLimiter", t, func() {
		testLimiterCopy := *testLimiter
		stubTestLimiter := ApplyGlobalVar(&testLimiter, &testLimiterCopy)

		const levelCount = 5
		accountLevel := 1

		Convey("return a new registered limiter for a corresponding account level", func() {
			limiter, err := testLimiter.RegisteredLimiter(baseTokenRate, accountLevel, levelCount)
			So(limiter, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("limit of registered limiter with account level is [1/5] should be 4", func() {
				So(limiter.Limit(), ShouldEqual, rate.Limit(4))
			})

			stubTestLimiter.Reset()
		})

		// table-driven test
		type testArg struct {
			accountLevel int
			levelCount   int
			err          error
		}
		testData := func(accountLevel int, levelCount int) *testArg {
			return &testArg{
				accountLevel: accountLevel,
				levelCount:   levelCount,
				err: fmt.Errorf("accountLevel or levelCount is invalid, given: "+
					"accountLevel=%v, levelCount=%v", accountLevel, levelCount),
			}
		}
		testTable := map[string]*testArg{
			"return nil when accountLevel = 0":          testData(0, 5),
			"return nil when levelCount = 0":            testData(5, 0),
			"return nil when accountLevel < levelCount": testData(6, 5),
		}

		for testDescription, testData := range testTable {
			Convey(testDescription, func() {
				limiter, err := testLimiter.RegisteredLimiter(baseTokenRate,
					testData.accountLevel, testData.levelCount)
				So(limiter, ShouldBeNil)
				So(err, ShouldResemble, testData.err)

				stubTestLimiter.Reset()
			})
		}

		Convey("return nil when baseTokenRate = 0", func() {
			limiter, err := testLimiter.RegisteredLimiter(0,
				5, 5)
			So(limiter, ShouldBeNil)
			So(err, ShouldResemble, tokenRateZeroError)
		})

		stubTestLimiter.Reset()
	})
}

func TestLimiter_Allow(t *testing.T) {
	Convey("TestAllow", t, func() {
		testLimiterCopy := *testLimiter
		stubTestLimiter := ApplyGlobalVar(&testLimiter, &testLimiterCopy)

		type testArg struct {
			limit  rate.Limit
			burst  int
			expect Assertion
		}
		testData := func(limit float64, burst int, expect Assertion) *testArg {
			return &testArg{
				limit:  rate.Limit(limit),
				burst:  burst,
				expect: expect,
			}
		}
		testTable := map[string]*testArg{
			"request should be allowed " +
				"when limit = burst": testData(1, 1, ShouldBeTrue),
			"request should be allowed when limit > burst":     testData(0, 1, ShouldBeTrue),
			"request should not be allowed when limit < burst": testData(1, 0, ShouldBeFalse),
		}

		for testDescription, testData := range testTable {
			Convey(testDescription, func() {
				testLimiter.GateLimiter().SetLimit(testData.limit)
				testLimiter.GateLimiter().SetBurst(testData.burst)
				So(testLimiter.Allow(), testData.expect)

				stubTestLimiter.Reset()
			})
		}
	})
}

func TestLimiter_AddPeerLimiter(t *testing.T) {
	Convey("Test AddPeerLimiter and PeerLimiter", t, func() {
		const peerIDStr = "12D3KooWJfFoQ2D1nukmG84DEh6gGEEE49yG6rPCdHoCqhF7YyL1"
		peerID, err := peer.Decode(peerIDStr)
		if err != nil {
			t.Errorf("decode peer id failed, error: %v", err)
		}
		limiter := rate.NewLimiter(1, 1)

		Convey("limiter returned should be exactly the one passed to", func() {
			addedLimiter := testLimiter.AddPeerLimiter(peerID, limiter)
			So(addedLimiter, ShouldEqual, limiter)
		})

		Convey("limiter fetched should be exactly the one added above", func() {
			fetchedLimiter := testLimiter.PeerLimiter(peerID)
			So(fetchedLimiter, ShouldResemble, limiter)
		})

		Convey("return nil when peerID cannot find in Limiter.peers map", func() {
			fetchedLimiter := testLimiter.PeerLimiter("呵呵")
			So(fetchedLimiter, ShouldBeNil)
		})
	})
}
