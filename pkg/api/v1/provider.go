package v1

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"net/http"
	"pando/pkg/api/types"
	"pando/pkg/api/v1/model"
	"pando/pkg/metrics"
	"pando/pkg/registry"
)

func (a *API) registerProvider() {
	metadata := a.router.Group("/provider")
	{
		metadata.POST("/register", a.providerRegister)
	}
}

func (a *API) providerRegister(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.PostProviderRegisterLatency)
	defer record()

	bodyBytes, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.Errorf("read register body failed: %v\n", err)
		handleError(ctx, http.StatusInternalServerError, InternalServerError)
		return
	}

	registerRequest, err := model.ReadRegisterRequest(bodyBytes)
	if err != nil {
		logger.Errorf("read register info failed: %v\n", err)
		handleError(ctx, http.StatusInternalServerError, InternalServerError)
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

	err = a.core.Registry.Register(providerInfo)
	if err != nil {
		handleError(ctx, http.StatusInternalServerError, fmt.Sprintf("register failed: %v", err))
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("register success", nil))
}
