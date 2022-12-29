package types

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewErrorResponse(t *testing.T) {
	t.Run("Test NewErrorResponse", func(t *testing.T) {
		testErrorResp := NewErrorResponse(http.StatusBadRequest, "bad request")
		assert.Equal(t, &ResponseJson{Code: http.StatusBadRequest, Message: "bad request"}, testErrorResp)
	})
}

func TestNewOKResponse(t *testing.T) {
	t.Run("Test NewOKResponse", func(t *testing.T) {
		testOKResp := NewOKResponse("ok", nil)
		assert.Equal(t, &ResponseJson{Code: http.StatusOK, Message: "ok", Data: nil}, testOKResp)
	})
}
