package defs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrKVNotFound = errors.New("not found")

func ErrNotFound() *APIError   { return &APIError{code: http.StatusNotFound} }
func ErrBadRequest() *APIError { return &APIError{code: http.StatusBadRequest} }
func ErrForbidden() *APIError  { return &APIError{code: http.StatusForbidden} }
func ErrInternal() *APIError   { return &APIError{code: http.StatusInternalServerError} }

type APIError struct {
	wrapped error
	code    int
	detail  string
}

func (e *APIError) Error() string {
	statusText := http.StatusText(e.code)

	if e.wrapped != nil {
		return fmt.Sprintf("%s: %v", statusText, e.wrapped)
	}

	return statusText
}

func (e *APIError) Wrap(err error) *APIError {
	e.wrapped = err
	return e
}

func (e *APIError) APIError() (int, string) {
	return e.code, e.detail
}

func (e *APIError) Code() int {
	return e.code
}

func (e *APIError) JSON() []byte {
	detail := e.detail
	if detail == "" && e.wrapped != nil {
		detail = e.wrapped.Error()
	}
	data, _ := json.Marshal(struct {
		Code   int
		Detail string
	}{
		Code:   e.code,
		Detail: detail,
	})

	return data
}

func (e *APIError) WithDetail(detail string) *APIError {
	e.detail = detail
	return e
}

func (e *APIError) Unwrap() error {
	return e.wrapped
}

func (e *APIError) Is(err error) bool {
	if _, ok := err.(*APIError); ok {
		return true
	}
	return false
}

func (e *APIError) As(err any) bool {
	if _, ok := err.(*APIError); ok {
		return true
	}
	return false
}
