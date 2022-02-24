package lotus

import (
	"context"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/kenlabs/pando/pkg/registry/discovery"
	"math/big"
	"net/url"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Discoverer struct {
	gatewayURL string
}

type ExpTipSet struct {
	Cids   []cid.Cid
	Blocks []interface{}
	Height int64
}

// NewDiscoverer creates a new lotus Discoverer
func NewDiscoverer(gateway string) (*Discoverer, error) {
	u, err := url.Parse(gateway)
	if err != nil {
		return nil, err
	}
	u.Scheme = "https"
	u.Path = "/rpc/v1"

	return &Discoverer{
		gatewayURL: u.String(),
	}, nil
}

//func (d *Discoverer) getMinerPeerAddr(minerInfo miner.MinerInfo) (peer.AddrInfo, error) {
//	multiaddrs := make([]multiaddr.Multiaddr, 0, len(minerInfo.Multiaddrs))
//	for _, a := range minerInfo.Multiaddrs {
//		maddr, err := multiaddr.NewMultiaddrBytes(a)
//		if err != nil {
//			continue
//		}
//		multiaddrs = append(multiaddrs, maddr)
//	}
//
//	return peer.AddrInfo{
//		ID:    *minerInfo.PeerId,
//		Addrs: multiaddrs,
//	}, nil
//}

//func (d *Discoverer) _Discover(ctx context.Context, peerID peer.ID, minerAddr string) (*discovery.Discovered, error) {
//	// todo fill
//	authToken := "<value found in ~/.lotus/token>"
//	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
//
//	var api lotusapi.FullNodeStruct
//
//	closer, err := jsonrpc.NewMergeClient(context.Background(),
//		"https://api.chain.love/rpc/v1",
//		"Filecoin",
//		[]interface{}{&api.Internal, &api.CommonStruct.Internal},
//		headers)
//	if err != nil {
//		log.Fatalf("connecting with lotus failed: %s", err)
//	}
//	defer closer()
//
//	// Get miner info from lotus
//	minerAddress, err := address.NewFromString(minerAddr)
//	if err != nil {
//		return nil, fmt.Errorf("invalid provider filecoin address: %s", err)
//	}
//
//	balance, err := api.WalletBalance(ctx, minerAddress)
//	if err != nil {
//		return nil, err
//	}
//
//	tsp, err := api.ChainHead(ctx)
//	if err != nil {
//		return nil, err
//	}
//	info, err := api.StateMinerInfo(ctx, minerAddress, tsp.Key())
//	if err != nil {
//		return nil, err
//	}
//
//	if *info.PeerId != peerID {
//		return nil, errors.New("provider id mismatch")
//	}
//
//	addrInfo, err := d.getMinerPeerAddr(info)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get account addrinfo from minerinfo: %s", err.Error())
//	}
//
//	return &discovery.Discovered{
//		AddrInfo: addrInfo,
//		Type:     discovery.MinerType,
//		Balance:  balance.Int,
//	}, nil
//}

func (d *Discoverer) Discover(ctx context.Context, peerID peer.ID, minerAddr string) (*discovery.Discovered, error) {
	return &discovery.Discovered{
		AddrInfo: peer.AddrInfo{ID: peerID},
		Type:     discovery.MinerType,
		Balance:  big.NewInt(0).Mul(registry.FIL, big.NewInt(5)),
	}, nil
}
