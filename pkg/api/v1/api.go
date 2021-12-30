package v1

import (
	"github.com/gin-gonic/gin"
	logging "github.com/ipfs/go-log/v2"
	core2 "pando/pkg/api/core"

	"pando/pkg/api/types"
)

var logger = logging.Logger("v1API")

type API struct {
	router *gin.Engine
	core   *core2.Core
}

type ErrorTemplate map[string]string

func NewV1API(router *gin.Engine, core *core2.Core) *API {
	return &API{
		router: router,
		core:   core,
	}
}

func (a *API) RegisterAPIs() {
	a.registerMetadata()
	a.registerProvider()
	a.registerPando()
}

func handleError(ctx *gin.Context, code int, errStr string) {
	ctx.AbortWithStatusJSON(code, types.NewErrorResponse(code, errStr))
}
