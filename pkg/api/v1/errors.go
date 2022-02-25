package v1

import "fmt"

const (
	InvalidQuery        = "invalid query"
	ResourceNotFound    = "resource not found"
	InternalServerError = "internal server error, please contact with administrator"
)

var (
	ErrorNotFound = fmt.Errorf("resource not found")
)
