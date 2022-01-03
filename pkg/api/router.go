package api

import (
	"github.com/gin-gonic/gin"

	"pando/pkg/api/core"
	"pando/pkg/api/middleware"
	v1Graphql "pando/pkg/api/v1/graphql"
	v1Http "pando/pkg/api/v1/http"
)

func NewHttpRouter(core *core.Core) *gin.Engine {
	httpRouter := gin.New()
	httpRouter.Use(middleware.WithLoggerFormatter())
	httpRouter.Use(middleware.WithCorsAllowAllOrigin())
	httpRouter.Use(gin.Recovery())

	v1HttpAPI := v1Http.NewV1HttpAPI(httpRouter, core)
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