package pando

import (
	"github.com/gin-gonic/gin"
	"github.com/kenlabs/pando/docs"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSwaggerSpecs(t *testing.T) {
	Convey("TestSwaggerSpecs", t, func() {
		responseRecorder := httptest.NewRecorder()
		testContext, _ := gin.CreateTestContext(responseRecorder)

		req, err := http.NewRequest("GET", "http://127.0.0.1", nil)
		testContext.Request = req

		mockAPI.swaggerSpecs(testContext)
		respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
		if err != nil {
			t.Error(err)
		}

		So(respBody, ShouldResemble, docs.SwaggerSpecs)
	})
}
