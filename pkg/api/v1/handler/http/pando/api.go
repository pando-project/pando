package pando

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/pando-project/pando/pkg/api/core"
	v1 "github.com/pando-project/pando/pkg/api/v1"
	"github.com/pando-project/pando/pkg/api/v1/controller"
	"github.com/pando-project/pando/pkg/option"
	"github.com/pando-project/pando/pkg/util/log"
	"net/http"

	"github.com/pando-project/pando/pkg/api/types"
)

var logger = log.NewSubsystemLogger()

type API struct {
	router     *gin.Engine
	controller *controller.Controller
}

type ErrorTemplate map[string]string

func NewV1HttpAPI(router *gin.Engine, core *core.Core, opt *option.DaemonOptions) *API {
	return &API{
		router:     router,
		controller: controller.New(core, opt),
	}
}

func (a *API) RegisterAPIs() {
	a.registerMetadata()
	a.registerProvider()
	a.registerPando()
	a.registerSwagger()
}

func HandleError(ctx *gin.Context, err error) {
	var apiErr *v1.Error
	var code = http.StatusBadRequest
	if errors.As(err, &apiErr) {
		ctx.AbortWithStatusJSON(apiErr.Status(), types.NewErrorResponse(apiErr.Status(), apiErr.Error()))
		return
	}

	ctx.AbortWithStatusJSON(code, types.NewErrorResponse(code, err.Error()))
}
