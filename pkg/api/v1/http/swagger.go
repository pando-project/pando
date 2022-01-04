package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pando/docs"
)

func (a *API) registerSwagger() {
	a.router.GET("/swagger/specs", a.swaggerSpecs)
}

func (a *API) swaggerSpecs(ctx *gin.Context) {
	ctx.Data(http.StatusOK, gin.MIMEPlain, docs.SwaggerSpecs)
}
