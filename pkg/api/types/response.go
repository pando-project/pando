package types

import (
	"encoding/json"
	"net/http"
)

type ResponseJson struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"Data"`
}

func NewErrorResponse(code int, message string) *ResponseJson {
	return &ResponseJson{
		Code:    code,
		Message: message,
	}
}

func NewOKResponse(message string, data interface{}) *ResponseJson {
	var res interface{}
	byteData, ok := data.([]byte)
	if ok {
		err := json.Unmarshal(byteData, &res)
		if err != nil {
			// todo failed unmarshal the []byte result(json)
		} else {
			return &ResponseJson{
				Code:    http.StatusOK,
				Message: message,
				Data:    res,
			}
		}
	}
	return &ResponseJson{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	}
}
