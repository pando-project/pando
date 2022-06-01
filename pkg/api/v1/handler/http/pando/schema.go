package pando

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kenlabs/pando/pkg/api/types"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/metrics"
	"io/ioutil"
	"net/http"
)

func (a *API) registerSchema() {
	schema := a.router.Group("/schema")
	{
		schema.POST("/register", a.schemaRegister)
	}
}

func (a *API) schemaRegister(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.PostSchemaRegisterLatency)
	defer record()

	bodyBytes, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.Errorf("read register body failed: %v\n", err)
		HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}

	err = a.controller.SchemaRegister(bodyBytes)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("register success", nil))
}
