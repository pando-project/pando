package httpadminserver

import (
	"Pando/api/v0/admin/model"
	"Pando/config"
	"Pando/internal/lotus"
	"Pando/internal/registry"
	"bytes"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testProtocolID = 0x300000

var ident = config.Identity{
	PeerID:  "12D3KooWPw6bfQbJHfKa2o5XpusChoq67iZoqgfnhecygjKsQRmG",
	PrivKey: "CAESQEQliDSXbU/zR4hrGNgAM0crtmxcZ49F3OwjmptYEFuU0b0TwLTJz/OlSBBuK7QDV2PiyGOCjDkyxSXymuqLu18=",
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

	disco, err := lotus.NewDiscoverer(discoveryCfg.LotusGateway)
	if err != nil {
		panic(err)
	}
	ds := datastore.NewMapDatastore()
	reg, err = registry.NewRegistry(discoveryCfg, ds, disco)
	if err != nil {
		panic(err)
	}

	hnd = newHandler(reg)

	providerID, err = peer.Decode(ident.PeerID)
	if err != nil {
		panic("Could not decode peer ID")
	}
}

func TestRegisterProvider(t *testing.T) {
	peerID, privKey, err := ident.Decode()
	if err != nil {
		t.Fatal(err)
	}

	addrs := []string{"/ip4/127.0.0.1/tcp/9999"}
	account := "t01000"
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
}