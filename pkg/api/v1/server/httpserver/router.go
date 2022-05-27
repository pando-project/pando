package httpserver

import (
	"github.com/gin-gonic/gin"
	v1Admin "github.com/kenlabs/pando/pkg/api/v1/handler/http/admin"
	v1Graphql "github.com/kenlabs/pando/pkg/api/v1/handler/http/graphql"
	"github.com/kenlabs/pando/pkg/api/v1/handler/http/pando"
	"github.com/kenlabs/pando/pkg/option"

	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/api/middleware"
)

func NewAdminRouter(core *core.Core, opt *option.DaemonOptions) *gin.Engine {
	adminRouter := gin.New()
	adminRouter.Use(gin.Recovery())

	v1AdminAPI := v1Admin.NewV1AdminAPI(adminRouter, core, opt)
	v1AdminAPI.RegisterAPIs()

	return adminRouter
}

func NewHttpRouter(core *core.Core, opt *option.DaemonOptions) *gin.Engine {
	httpRouter := gin.New()
	httpRouter.Use(middleware.WithLoggerFormatter())
	httpRouter.Use(middleware.WithCorsAllowAllOrigin())
	httpRouter.Use(gin.Recovery())
	httpRouter.Use(middleware.WithAPIDoc())

	v1HttpAPI := pando.NewV1HttpAPI(httpRouter, core, opt)
	v1HttpAPI.RegisterAPIs()

	return httpRouter
}

func NewGraphqlRouter(core *core.Core) *gin.Engine {
	graphqlRouter := gin.New()
	graphqlRouter.Use(middleware.WithLoggerFormatter())
	graphqlRouter.Use(middleware.WithCorsAllowAllOrigin())
	graphqlRouter.Use(gin.Recovery())

	v1GraphAPI := v1Graphql.NewV1GraphqlAPI(graphqlRouter, core)
	v1GraphAPI.RegisterAPIs()

	return graphqlRouter
}
