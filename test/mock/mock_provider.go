package mock

import (
	"context"
	"fmt"
	goLegs "github.com/filecoin-project/go-legs"
	"github.com/filecoin-project/go-legs/dtsync"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/kenlabs/pando/pkg/types/schema"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"math/rand"
	"time"
)

type ProviderMock struct {
	ID           peer.ID
	pk           crypto.PrivKey
	LegsProvider goLegs.Publisher
	lsys         *linking.LinkSystem
	DagService   format.DAGService
	prevMetaLink datamodel.Link
}

func getDagNodes() []format.Node {
	a := merkledag.NewRawNode([]byte("aaaaa" + string(rune(rand.Intn(100000)))))
	b := merkledag.NewRawNode([]byte("bbbbb" + string(rune(rand.Intn(100000)))))
	c := merkledag.NewRawNode([]byte("ccccc" + string(rune(rand.Intn(100000)))))

	nd1 := &merkledag.ProtoNode{}
	err := nd1.AddNodeLink("cat", a)
	if err != nil {
		return nil
	}

	nd2 := &merkledag.ProtoNode{}
	if err = nd2.AddNodeLink("first", nd1); err != nil {
		return nil
	}
	if err := nd2.AddNodeLink("dog", b); err != nil {
		return nil
	}

	nd3 := &merkledag.ProtoNode{}
	if err := nd3.AddNodeLink("second", nd2); err != nil {
		return nil
	}
	if err := nd3.AddNodeLink("bear", c); err != nil {
		return nil
	}

	return []format.Node{nd3, nd2, nd1, c, b, a}
}

func (p *ProviderMock) genMetaWithBytesPayload(link datamodel.Link) (*schema.Metadata, error) {
	data := make([]byte, 256)
	rand.Read(data)
	var meta *schema.Metadata
	var err error
	if link == nil {
		meta, err = schema.NewMetaWithBytesPayload(data, p.ID, p.pk)
		if err != nil {
			return nil, fmt.Errorf("failed to create meta: %s", err.Error())
		}
	} else {
		meta, err = schema.NewMetadataWithLink(data, p.ID, p.pk, link)
		if err != nil {
			return nil, fmt.Errorf("failed to create meta with link: %s", err.Error())
		}
	}
	return meta, nil
}

func (p *ProviderMock) genMetaWithLinkPayload(prev datamodel.Link) (*schema.Metadata, cid.Cid, error) {
	data := make([]byte, 256)
	rand.Read(data)
	var meta *schema.Metadata
	var err error
	pnode := basicnode.NewBytes(data)
	lnk, err := p.lsys.Store(ipld.LinkContext{}, schema.LinkProto, pnode)
	if err != nil {
		return nil, cid.Undef, err
	}
	meta, err = schema.NewMetaWithPayloadNode(basicnode.NewLink(lnk), p.ID, p.pk, prev)
	if err != nil {
		return nil, cid.Undef, fmt.Errorf("failed to create meta: %s", err.Error())
	}
	return meta, lnk.(cidlink.Link).Cid, nil
}

func NewMockProvider(p *PandoMock) (*ProviderMock, error) {
	rand.Seed(time.Now().UnixNano())
	// test provider legs
	srcHost, err := libp2p.New()
	if err != nil {
		return nil, err
	}
	pk := srcHost.Peerstore().PrivKey(srcHost.ID())
	srcDatastore := dssync.MutexWrap(datastore.NewMapDatastore())
	srcBlockstore := blockstore.NewBlockstore(srcDatastore)
	//srcPandoStore, err := store.NewStoreFromDatastore(context.Background(), srcDatastore, &config.StoreConfig{SnapShotInterval: "60m"})
	srcLinkSystem := basicLinkSystem(srcBlockstore)
	dags := merkledag.NewDAGService(blockservice.New(srcBlockstore, offline.Exchange(srcBlockstore)))
	legsPublisher, err := dtsync.NewPublisher(srcHost, srcDatastore, srcLinkSystem, topic)
	if err != nil {
		return nil, err
	}

	multiAddress := p.Host.Addrs()[0].String() + "/ipfs/" + p.Host.ID().String()
	peerInfo, err := peer.AddrInfoFromString(multiAddress)
	if err != nil {
		return nil, err
	}

	if err = srcHost.Connect(context.Background(), *peerInfo); err != nil {
		return nil, err
	}

	return &ProviderMock{
		ID:           srcHost.ID(),
		LegsProvider: legsPublisher,
		lsys:         &srcLinkSystem,
		DagService:   dags,
		pk:           pk,
	}, nil
}

func (p *ProviderMock) SendDag() ([]cid.Cid, error) {
	cidlist := make([]cid.Cid, 0)

	// store test dag
	dagNodes := getDagNodes()
	for i := 0; i < len(dagNodes); i++ {
		err := p.DagService.Add(context.Background(), dagNodes[i])
		if err != nil {
			return nil, err
		}
		cidlist = append(cidlist, dagNodes[i].Cid())
	}

	err := p.LegsProvider.UpdateRoot(context.Background(), dagNodes[0].Cid())
	if err != nil {
		return nil, err
	}

	return cidlist, nil
}

func (p *ProviderMock) SendMeta(update bool) (cid.Cid, error) {
	meta, err := p.genMetaWithBytesPayload(p.prevMetaLink)
	if err != nil {
		return cid.Undef, err
	}
	mnode, err := meta.ToNode()
	if err != nil {
		return cid.Undef, err
	}
	lnk, err := p.lsys.Store(ipld.LinkContext{}, schema.LinkProto, mnode)
	if err != nil {
		return cid.Undef, err
	}
	if update {
		err = p.LegsProvider.UpdateRoot(context.Background(), lnk.(cidlink.Link).Cid)
		if err != nil {
			return cid.Undef, err
		}
	}
	p.prevMetaLink = lnk
	return lnk.(cidlink.Link).Cid, nil
}

// the payload is a ipld link
func (p *ProviderMock) SendMetaWithDataLink(update bool) (cid.Cid, cid.Cid, error) {
	meta, c, err := p.genMetaWithLinkPayload(p.prevMetaLink)
	if err != nil {
		return cid.Undef, cid.Undef, err
	}
	mnode, err := meta.ToNode()
	if err != nil {
		return cid.Undef, cid.Undef, err
	}
	lnk, err := p.lsys.Store(ipld.LinkContext{}, schema.LinkProto, mnode)
	if err != nil {
		return cid.Undef, cid.Undef, err
	}
	if update {
		err = p.LegsProvider.UpdateRoot(context.Background(), lnk.(cidlink.Link).Cid)
		if err != nil {
			return cid.Undef, cid.Undef, err
		}
	}
	p.prevMetaLink = lnk
	return lnk.(cidlink.Link).Cid, c, nil
}

func (p *ProviderMock) Close() error {
	return p.LegsProvider.Close()
}
