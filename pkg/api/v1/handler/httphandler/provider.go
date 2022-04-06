package httphandler

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
		handleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}

	err = a.handler.ProviderRegister(ctx, bodyBytes)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("register success", nil))
}

func writeProviderInfo(ctx *gin.Context, info []*registry.ProviderInfo) {
	res, err := model.GetProviderRes(info)
	if err != nil {
		logger.Errorf("failed to marshal provider info, err: %v", err)
		handleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", res))
}

func (a *API) listProviderInfo(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetRegisteredProviderInfoLatency)
	defer record()

	peerid, err := decodePeerid(ctx)
	if err != nil {
		handleError(ctx, v1.NewError(errors.New("invalid peerid"), http.StatusBadRequest))
		return
	}

	info, err := a.handler.ListProviderInfo(peerid)
	if err != nil {
		handleError(ctx, err)
		return
	}
	writeProviderInfo(ctx, info)

}

//func (a *API) listProviderInfo(ctx *gin.Context) {
//	record := metrics.APITimer(context.Background(), metrics.GetRegisteredProviderInfoLatency)
//	defer record()
//
//	peerid, err := decodePeerid(ctx)
//
//	if err != nil {
//		handleError(ctx, http.StatusBadRequest, fmt.Errorf("invalid peerid"))
//		return
//	} else {
//		info := a.core.Registry.ProviderInfo(peerid)
//		if info == nil {
//			handleError(ctx, http.StatusNotFound, fmt.Errorf("provider not found"))
//			return
//		}
//		writeProviderInfo(ctx, info)
//	}
//}

func (a *API) listProviderHead(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetProviderHeadLatency)
	defer record()

	peerid, err := decodePeerid(ctx)
	if err != nil {
		handleError(ctx, v1.NewError(errors.New("invalid peerid"), http.StatusBadRequest))
		return
	}
	headCid, err := a.handler.ListProviderHead(peerid)
	if err != nil {
		handleError(ctx, err)
		return
	}

	res := struct {
		Cid string
	}{}
	res.Cid = headCid.String()
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", res))
}

//func (a *API) listProviderHead(ctx *gin.Context) {
//	record := metrics.APITimer(context.Background(), metrics.GetProviderHeadLatency)
//	defer record()
//
//	failError := "failed to retrieve the head of the provider, err: %v"
//
//	peerid, err := decodePeerid(ctx)
//	if err != nil || peerid == "" {
//		handleError(ctx, http.StatusBadRequest, fmt.Errorf("invalid peerid"))
//		return
//	} else {
//		res := struct {
//			Cid string
//		}{}
//		var cidBytes []byte
//		cidBytes, err = a.core.StoreInstance.DataStore.Get(ctx, datastore.NewKey(legs.SyncPrefix+peerid.String()))
//		if err != nil {
//			if err == datastore.ErrNotFound {
//				handleError(ctx, http.StatusNotFound, errors.New("not found the head of this provider"))
//				return
//			}
//			logger.Errorf(failError, err)
//			handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
//			return
//		}
//		_, providerCid, err := cid.CidFromBytes(cidBytes)
//		if err != nil {
//			logger.Errorf(failError, err)
//			handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
//			return
//		}
//		res.Cid = providerCid.String()
//		ctx.JSON(http.StatusOK, types.NewOKResponse("OK", res))
//	}
//}

func decodePeerid(ctx *gin.Context) (peer.ID, error) {
	peeridStr := ctx.Query("peerid")
	if peeridStr == "" {
		return "", nil
	} else {
		return peer.Decode(peeridStr)
	}
}
