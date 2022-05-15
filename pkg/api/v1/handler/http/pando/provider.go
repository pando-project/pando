package pando

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kenlabs/pando/pkg/api/types"
	"github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/api/v1/model"
	"github.com/kenlabs/pando/pkg/metrics"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/libp2p/go-libp2p-core/peer"
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
		HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}

	err = a.controller.ProviderRegister(ctx, bodyBytes)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("register success", nil))
}

func writeProviderInfo(ctx *gin.Context, info []*registry.ProviderInfo) {
	res, err := model.GetProviderRes(info)
	if err != nil {
		logger.Errorf("failed to marshal provider info, err: %v", err)
		HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", res))
}

func (a *API) listProviderInfo(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetRegisteredProviderInfoLatency)
	defer record()

	peerid, err := decodePeerid(ctx)
	if err != nil {
		HandleError(ctx, v1.NewError(errors.New("invalid peerid"), http.StatusBadRequest))
		return
	}

	info, err := a.controller.ListProviderInfo(peerid)
	if err != nil {
		HandleError(ctx, err)
		return
	}
	writeProviderInfo(ctx, info)

}

func (a *API) listProviderHead(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetProviderHeadLatency)
	defer record()

	peerid, err := decodePeerid(ctx)
	if err != nil {
		HandleError(ctx, v1.NewError(errors.New("invalid peerid"), http.StatusBadRequest))
		return
	}
	headCid, err := a.controller.ListProviderHead(peerid)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	res := struct {
		Cid string
	}{}
	res.Cid = headCid.String()
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", res))
}

func decodePeerid(ctx *gin.Context) (peer.ID, error) {
	peeridStr := ctx.Query("peerid")
	if peeridStr == "" {
		return "", nil
	} else {
		return peer.Decode(peeridStr)
	}
}
