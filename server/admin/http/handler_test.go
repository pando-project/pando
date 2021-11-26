package httpadminserver

import (
	"Pando/api/v0/admin/model"
	"Pando/config"
	"Pando/internal/registry"
	"Pando/internal/registry/discovery"
	"bytes"
	"context"
	"errors"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testProtocolID = 0x300000

var ident = config.Identity{
	PeerID:  "12D3KooWPw6bfQbJHfKa2o5XpusChoq67iZoqgfnhecygjKsQRmG",
	PrivKey: "CAESQEQliDSXbU/zR4hrGNgAM0crtmxcZ49F3OwjmptYEFuU0b0TwLTJz/OlSBBuK7QDV2PiyGOCjDkyxSXymuqLu18=",
}

type mockDiscoverer struct {
	discoverRsp *discovery.Discovered
}

func newMockDiscoverer(providerID string) *mockDiscoverer {
	peerID, err := peer.Decode(providerID)
	if err != nil {
		panic(err)
	}

	return &mockDiscoverer{
		discoverRsp: &discovery.Discovered{
			AddrInfo: peer.AddrInfo{
				ID: peerID,
			},
			Type:    discovery.MinerType,
			Balance: new(big.Int).Mul(registry.FIL, big.NewInt(10)),
		},
	}
}

func (m *mockDiscoverer) Discover(ctx context.Context, peerID peer.ID, filecoinAddr string) (*discovery.Discovered, error) {
	if filecoinAddr == "bad1234" {
		return nil, errors.New("unknown miner")
	}

	return m.discoverRsp, nil
}

var providerID peer.ID

var hnd *httpHandler
var reg *registry.Registry

func init() {
	var discoveryCfg = config.Discovery{
		Policy: config.Policy{
			Allow:       false,
			Except:      []string{ident.PeerID},
			Trust:       false,
			TrustExcept: []string{ident.PeerID},
		},
		LotusGateway: "api.chain.love",
	}
	var aclCfg = config.AccountLevel{Threshold: []int{1, 10, 99}}
	var err error
	providerID, err = peer.Decode(ident.PeerID)
	if err != nil {
		panic("Could not decode account ID")
	}

	disco := newMockDiscoverer(providerID.String())

	ds := datastore.NewMapDatastore()
	reg, err = registry.NewRegistry(&discoveryCfg, &aclCfg, ds, disco)
	if err != nil {
		panic(err)
	}

	hnd = newHandler(reg)

}

func TestRegisterProvider(t *testing.T) {
	peerID, privKey, err := ident.Decode()
	if err != nil {
		t.Fatal(err)
	}

	addrs := []string{"/ip4/127.0.0.1/tcp/9999"}
	account := "t01234"
	data, err := model.MakeRegisterRequest(peerID, privKey, addrs, account)
	if err != nil {
		t.Fatal(err)
	}
	reqBody := bytes.NewBuffer(data)

	req := httptest.NewRequest("POST", "http://example.com/providers", reqBody)
	w := httptest.NewRecorder()
	hnd.RegisterProvider(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatal("expected response to be", http.StatusOK)
	}

	pinfo := reg.ProviderInfo(peerID)
	if pinfo == nil {
		t.Fatal("provider was not registered")
	}
	level, err := reg.ProviderAccountLevel(peerID)
	assert.NoError(t, err)
	assert.Equal(t, level, 3, "not get weight rightly")
}

func TestRegisterProviderBadRequest(t *testing.T) {
	// error data, should be envelop
	reqBody := bytes.NewBuffer([]byte("bad request data"))

	req := httptest.NewRequest("POST", "http://example.com/providers", reqBody)
	w := httptest.NewRecorder()
	hnd.RegisterProvider(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("expected response to be", http.StatusBadRequest)
	}

}
