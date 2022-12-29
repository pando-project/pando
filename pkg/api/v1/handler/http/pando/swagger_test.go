package pando

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pando-project/pando/docs"
	"github.com/stretchr/testify/assert"
)

func TestSwaggerSpecs(t *testing.T) {
	t.Run("TestSwaggerSpecs", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		testContext, _ := gin.CreateTestContext(responseRecorder)

		req, err := http.NewRequest("GET", "http://127.0.0.1", nil)
		testContext.Request = req

		mockAPI.swaggerSpecs(testContext)
		respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, docs.SwaggerSpecs, respBody)
	})
}
