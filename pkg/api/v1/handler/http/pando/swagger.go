package pando

import (
	"github.com/gin-gonic/gin"
	"github.com/pando-project/pando/docs"
	"net/http"
)

func (a *API) registerSwagger() {
	a.router.GET("/swagger/specs", a.swaggerSpecs)
}

func (a *API) swaggerSpecs(ctx *gin.Context) {
	ctx.Data(http.StatusOK, gin.MIMEPlain, docs.SwaggerSpecs)
}
