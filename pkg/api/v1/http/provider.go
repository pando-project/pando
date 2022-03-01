package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/kenlabs/pando/pkg/api/types"
	"github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/pkg/metrics"
	"github.com/kenlabs/pando/pkg/register"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"net/http"
)

func (a *API) registerProvider() {
	provider := a.router.Group("/provider")
	{
		provider.POST("/register", a.providerRegister)
		provider.GET("/info", a.listProviderInfo)
		provider.GET("/head", a.listProviderHead)
	}
}

func (a *API) providerRegister(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.PostProviderRegisterLatency)
	defer record()

	bodyBytes, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.Errorf("read register body failed: %v\n", err)
		handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
		return
	}

	registerRequest, err := register.ReadRegisterRequest(bodyBytes)
	if err != nil {
		logger.Errorf("read register info failed: %v\n", err)
		handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
		return
	}

	if len(registerRequest.PeerID) == 0 {
		handleError(ctx, http.StatusBadRequest, "missing account id")
		return
	}

	if err = a.core.Registry.CheckSequence(registerRequest.PeerID, registerRequest.Seq); err != nil {
		handleError(ctx, http.StatusBadRequest, fmt.Sprintf("bad sequence: %v", err.Error()))
		return
	}

	providerMultiAddr := make([]multiaddr.Multiaddr, len(registerRequest.Addrs))
	for i, m := range registerRequest.Addrs {
		var err error
		providerMultiAddr[i], err = multiaddr.NewMultiaddr(m)
		if err != nil {
			handleError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid address: %s", providerMultiAddr[i]))
			return
		}
	}

	providerInfo := &registry.ProviderInfo{
		AddrInfo: peer.AddrInfo{
			ID:    registerRequest.PeerID,
			Addrs: providerMultiAddr,
		},
		DiscoveryAddr: registerRequest.MinerAccount,
	}

	err = a.core.Registry.Register(ctx, providerInfo)
	if err != nil {
		handleError(ctx, http.StatusInternalServerError, fmt.Sprintf("register failed: %v", err))
		return
	}

	logger.Debugf("pando register success: %s", providerInfo.AddrInfo.ID)

	ctx.JSON(http.StatusOK, types.NewOKResponse("register success", nil))
}

type providerInfoRes map[string]struct {
	MultiAddr []string
	MinerAddr string
}

func writeProviderInfo(ctx *gin.Context, info []*registry.ProviderInfo, unRegInfo []*registry.ProviderInfo) {
	res := make(map[string]providerInfoRes)
	res["registeredProviders"] = make(providerInfoRes)
	res["unregisteredProviders"] = make(providerInfoRes)
	provInfos := res["registeredProviders"]
	unprovInfos := res["unregisteredProviders"]
	for _, provider := range info {
		peeridStr := provider.AddrInfo.ID.String()
		addrs := make([]string, 0)
		for _, addr := range provider.AddrInfo.Addrs {
			addrs = append(addrs, addr.String())
		}
		provInfos[peeridStr] = struct {
			MultiAddr []string
			MinerAddr string
		}{MultiAddr: addrs, MinerAddr: provider.DiscoveryAddr}
	}
	for _, provider := range unRegInfo {
		peeridStr := provider.AddrInfo.ID.String()
		addrs := make([]string, 0)
		for _, addr := range provider.AddrInfo.Addrs {
			addrs = append(addrs, addr.String())
		}
		unprovInfos[peeridStr] = struct {
			MultiAddr []string
			MinerAddr string
		}{MultiAddr: addrs, MinerAddr: provider.DiscoveryAddr}
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", res))
}

func (a *API) listProviderInfo(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetRegisteredProviderInfoLatency)
	defer record()

	peerid, err := decodePeerid(ctx)

	if err != nil {
		handleError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid peerid"))
		return
	} else {
		info := a.core.Registry.ProviderInfo(peerid)
		unregInfo := a.core.Registry.UnregProviderInfo(peerid)
		if info == nil && unregInfo == nil {
			handleError(ctx, http.StatusNotFound, fmt.Sprintf("provider not found"))
			return
		}
		writeProviderInfo(ctx, info, unregInfo)
	}
}

func (a *API) listProviderHead(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetProviderHeadLatency)
	defer record()

	failError := "failed to retrieve the head of the provider, err: %v"

	peerid, err := decodePeerid(ctx)
	if err != nil || peerid == "" {
		handleError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid peerid"))
		return
	} else {
		res := struct {
			Cid string
		}{}
		cidBytes, err := a.core.StoreInstance.DataStore.Get(ctx, datastore.NewKey(legs.SyncPrefix+peerid.String()))
		if err != nil && err != datastore.ErrNotFound {
			logger.Errorf(failError, err)
			handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
			return
		}
		_, providerCid, err := cid.CidFromBytes(cidBytes)
		if err != nil {
			logger.Errorf(failError, err)
			handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
			return
		}
		res.Cid = providerCid.String()
		ctx.JSON(http.StatusOK, types.NewOKResponse("OK", res))
	}
}

func decodePeerid(ctx *gin.Context) (peer.ID, error) {
	peeridStr := ctx.Query("peerid")
	if peeridStr == "" {
		return "", nil
	} else {
		return peer.Decode(peeridStr)
	}
}
