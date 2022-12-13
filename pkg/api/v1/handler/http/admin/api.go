package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/pando-project/pando/pkg/api/core"
	"github.com/pando-project/pando/pkg/option"
	"github.com/pando-project/pando/pkg/util/log"
)

var logger = log.NewSubsystemLogger()

type API struct {
	router  *gin.Engine
	core    *core.Core
	options *option.DaemonOptions
}

func NewV1AdminAPI(router *gin.Engine, core *core.Core, opt *option.DaemonOptions) *API {
	return &API{
		router:  router,
		core:    core,
		options: opt,
	}
}

func (a *API) RegisterAPIs() {
	a.registerBackup()
}

//func handleError(ctx *gin.Context, code int, errStr string) {
//	ctx.AbortWithStatusJSON(code, types.NewErrorResponse(code, errStr))
//}
