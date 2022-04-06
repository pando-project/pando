package httphandler

import (
	"errors"
	"github.com/gin-gonic/gin"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/api/core"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/api/v1/handler"
	"github.com/kenlabs/pando/pkg/option"
	"net/http"

	"github.com/kenlabs/pando/pkg/api/types"
)

var logger = logging.Logger("v1HttpAPI")

type API struct {
	router  *gin.Engine
	handler *handler.ServerHandler
}

type ErrorTemplate map[string]string

func NewV1HttpAPI(router *gin.Engine, core *core.Core, opt *option.Options) *API {
	return &API{
		router:  router,
		handler: handler.New(core, opt),
	}
}

func (a *API) RegisterAPIs() {
	a.registerMetadata()
	a.registerProvider()
	a.registerPando()
	a.registerSwagger()
}

func handleError(ctx *gin.Context, err error) {
	var apiErr *v1.Error
	var code = http.StatusBadRequest
	if errors.As(err, &apiErr) {
		ctx.AbortWithStatusJSON(apiErr.Status(), types.NewErrorResponse(apiErr.Status(), apiErr.Error()))
		return
	}

	ctx.AbortWithStatusJSON(code, types.NewErrorResponse(code, err.Error()))
}
