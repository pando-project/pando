package types

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
	return &ResponseJson{
		Message: message,
		Data:    data,
	}
}
