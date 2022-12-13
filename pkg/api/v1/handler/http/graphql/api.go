package graphql

import (
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/pando-project/pando/pkg/api/core"
	"github.com/pando-project/pando/pkg/util/log"
)

var logger = log.NewSubsystemLogger()

type API struct {
	router *gin.Engine
	core   *core.Core
	schema graphql.Schema
}

type ErrorTemplate map[string]string

func NewV1GraphqlAPI(router *gin.Engine, core *core.Core) *API {
	return &API{
		router: router,
		core:   core,
	}
}

func (a *API) RegisterAPIs() {
	a.registerGraphql()
}
