package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	InvalidQuery        = "invalid query"
	ResourceNotFound    = errors.New("resource not found")
	InternalServerError = errors.New("internal server error, please contact with administrator")
)

type Error struct {
	err    error
	status int
}

func NewError(err error, status int) *Error {
	return &Error{
		err:    err,
		status: status,
	}
}

func (e *Error) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	if e.status == 0 {
		return ""
	}
	// If there is only status, then return status text
	if text := http.StatusText(e.status); text != "" {
		return fmt.Sprintf("%d %s", e.status, text)
	}
	return fmt.Sprintf("%d", e.status)
}

func (e *Error) Status() int {
	return e.status
}

func (e *Error) Text() string {
	parts := make([]string, 0, 5)
	if e.status != 0 {
		parts = append(parts, fmt.Sprintf("%d", e.status))
		text := http.StatusText(e.status)
		if text != "" {
			parts = append(parts, " ")
			parts = append(parts, text)
		}
	}
	if e.err != nil {
		if len(parts) != 0 {
			parts = append(parts, ": ")
		}
		parts = append(parts, e.err.Error())
	}

	return strings.Join(parts, "")
}

func (e *Error) Unwrap() error {
	return e.err
}
