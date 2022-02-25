package graphql

import (
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/api/core"
)

var logger = logging.Logger("v1GraphqlAPI")

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
