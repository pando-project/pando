package http

import (
	"github.com/gin-gonic/gin"
	logging "github.com/ipfs/go-log/v2"
	"pando/pkg/api/core"
	"pando/pkg/option"

	"pando/pkg/api/types"
)

var logger = logging.Logger("v1HttpAPI")

type API struct {
	router  *gin.Engine
	core    *core.Core
	options *option.Options
}

type ErrorTemplate map[string]string

func NewV1HttpAPI(router *gin.Engine, core *core.Core, opt *option.Options) *API {
	return &API{
		router:  router,
		core:    core,
		options: opt,
	}
}

func (a *API) RegisterAPIs() {
	a.registerMetadata()
	a.registerProvider()
	a.registerPando()
	a.registerSwagger()
}

func handleError(ctx *gin.Context, code int, errStr string) {
	ctx.AbortWithStatusJSON(code, types.NewErrorResponse(code, errStr))
}
