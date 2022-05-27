package types

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

func TestNewErrorResponse(t *testing.T) {
	Convey("Test NewErrorResponse", t, func() {
		testErrorResp := NewErrorResponse(http.StatusBadRequest, "bad request")
		ShouldResemble(testErrorResp, &ResponseJson{Code: http.StatusBadRequest, Message: "bad request"})
	})
}

func TestNewOKResponse(t *testing.T) {
	Convey("Test NewOKResponse", t, func() {
		testOKResp := NewOKResponse("ok", nil)
		ShouldResemble(testOKResp, &ResponseJson{Code: http.StatusOK, Message: "ok", Data: nil})
	})
}
