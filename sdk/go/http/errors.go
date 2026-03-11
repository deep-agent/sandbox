package http

import (
	"fmt"
	"os"

	"github.com/deep-agent/sandbox/types/consts"
)

// APIError represents an error returned by the sandbox API.
type APIError struct {
	Code    int
	Message string
}

func newAPIError(code int, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (code %d): %s", e.Code, e.Message)
}

func (e *APIError) Unwrap() error {
	if e.Code == consts.CodeNotFound {
		return os.ErrNotExist
	}
	return nil
}
