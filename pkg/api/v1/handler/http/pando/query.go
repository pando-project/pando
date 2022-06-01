package pando

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/kenlabs/pando/pkg/api/types"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/metrics"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"io/ioutil"
	"net/http"
)

func (a *API) registerQuery() {
	a.router.POST("/query", a.query)
}

type queryRequest struct {
	DBName   string `json:"DBName"`
	QueryStr string `json:"QueryStr"`
}

func (a *API) query(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.PostQueryLatency)
	defer record()

	bodyBytes, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.Errorf("read query body failed: %v\n", err)
		HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}

	queryReq := queryRequest{}
	err = json.Unmarshal(bodyBytes, &queryReq)
	if err != nil {
		logger.Errorf("query request is invalid: %v", err)
		HandleError(ctx, v1.NewError(v1.InvalidQuery, http.StatusBadRequest))
		return
	}

	respData, err := a.controller.Query(queryReq.DBName, queryReq.QueryStr)
	if err != nil {
		logger.Errorf("query failed: %v", err)
		if err == bsonrw.ErrInvalidJSON {
			HandleError(ctx, v1.NewError(v1.InvalidQuery, http.StatusBadRequest))
		}
		HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", respData))
}
