package admin

import (
	"github.com/gin-gonic/gin"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/option"
)

var logger = logging.Logger("v1AdminAPI")

type API struct {
	router  *gin.Engine
	core    *core.Core
	options *option.Options
}

func NewV1AdminAPI(router *gin.Engine, core *core.Core, opt *option.Options) *API {
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
