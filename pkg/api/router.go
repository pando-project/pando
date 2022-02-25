package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kenlabs/pando/pkg/option"

	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/api/middleware"
	v1Graphql "github.com/kenlabs/pando/pkg/api/v1/graphql"
	v1Http "github.com/kenlabs/pando/pkg/api/v1/http"
)

func NewHttpRouter(core *core.Core, opt *option.Options) *gin.Engine {
	httpRouter := gin.New()
	httpRouter.Use(middleware.WithLoggerFormatter())
	httpRouter.Use(middleware.WithCorsAllowAllOrigin())
	httpRouter.Use(gin.Recovery())
	httpRouter.Use(middleware.WithAPIDoc())

	v1HttpAPI := v1Http.NewV1HttpAPI(httpRouter, core, opt)
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
