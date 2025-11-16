// Package review_errors provides typed error handling for the review service.
package review_errors

import (
	"fmt"
	"net/http"
)

// InfrastructureError represents failures in external dependencies.
type InfrastructureError struct {
	HTTPStatus int    // 8 bytes
	Code       string // 16 bytes
	Message    string // 16 bytes
	Cause      error  // 16 bytes
}

func (e *InfrastructureError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// StatusCode returns the HTTP status code for this infrastructure error
func (e *InfrastructureError) StatusCode() int {
	if e.HTTPStatus > 0 {
		return e.HTTPStatus
	}
	return http.StatusServiceUnavailable
}

// BusinessError represents violations of business rules.
type BusinessError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// StatusCode returns the HTTP status code for this business error
func (e *BusinessError) StatusCode() int {
	if e.HTTPStatus > 0 {
		return e.HTTPStatus
	}
	return http.StatusBadRequest
}
