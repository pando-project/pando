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

	ctx.JSON(http.StatusOK, types.NewOKResponse("register success", nil))
}

type providerInfoRes map[string]struct {
	MultiAddr []string
	MinerAddr string
}

func writeProviderInfo(ctx *gin.Context, info []*registry.ProviderInfo) {
	res := make(map[string]providerInfoRes)
	res["registeredProviders"] = make(providerInfoRes)
	provInfos := res["registeredProviders"]
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
	ctx.JSON(http.StatusOK, types.NewOKResponse("find registerd provider successfully", res))
}

func (a *API) listProviderInfo(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetRegisteredProviderInfoLatency)
	defer record()

	peerid, err := getPeerid(ctx)
	if err != nil && err != v1.ErrorNotFound {
		handleError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid peerid: %s, err: %s",
			ctx.Query("peerid"), err.Error()))
		return
	} else if err == v1.ErrorNotFound {
		// list all registered providers' info
		infos := a.core.Registry.AllProviderInfo()
		writeProviderInfo(ctx, infos)
	} else {
		info := a.core.Registry.ProviderInfo(peerid)
		if info == nil {
			handleError(ctx, http.StatusNotFound, fmt.Sprintf("not found registerd provider: %s ", peerid.String()))
			return
		}
		writeProviderInfo(ctx, []*registry.ProviderInfo{info})
	}
}

func (a *API) listProviderHead(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetProviderHeadLatency)
	defer record()

	peerid, err := getPeerid(ctx)
	if err != nil {
		handleError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid peerid: %s, err: %s",
			ctx.Query("peerid"), err.Error()))
		return
	} else {
		res := struct {
			Cid string
		}{}
		cidBytes, err := a.core.StoreInstance.DataStore.Get(ctx, datastore.NewKey(legs.SyncPrefix+peerid.String()))
		if err != nil && err != datastore.ErrNotFound {
			logger.Errorf("failed to get provider head: %s", err.Error())
			handleError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get provider head: %s", err.Error()))
			return
		}
		_, pcid, err := cid.CidFromBytes(cidBytes)
		if err != nil {
			logger.Errorf("failed to get provider head: %s", err.Error())
			handleError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get provider head: %s", err.Error()))
			return
		}
		res.Cid = pcid.String()
		ctx.JSON(http.StatusOK, types.NewOKResponse("find provider head successfully", res))
	}
}

func getPeerid(ctx *gin.Context) (peer.ID, error) {
	peeridStr := ctx.Query("peerid")
	if peeridStr == "" {
		return "", v1.ErrorNotFound
	} else {
		peerid, err := peer.Decode(peeridStr)
		if err != nil {
			return "", err
		}
		return peerid, nil
	}
}
