package errors

import "fmt"

// IPQuorumError is the base error type for all IPQuorum operations
type IPQuorumError struct {
	Message string
	Err     error
}

func (e *IPQuorumError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *IPQuorumError) Unwrap() error {
	return e.Err
}

// ValidationError represents configuration validation failures
type ValidationError struct {
	IPQuorumError
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		IPQuorumError: IPQuorumError{Message: message},
	}
}

// AuthenticationError represents authentication failures
type AuthenticationError struct {
	IPQuorumError
	StatusCode int
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string, statusCode int) *AuthenticationError {
	return &AuthenticationError{
		IPQuorumError: IPQuorumError{Message: message},
		StatusCode:    statusCode,
	}
}

// NetworkError represents network connectivity issues
type NetworkError struct {
	IPQuorumError
}

// NewNetworkError creates a new network error
func NewNetworkError(message string, err error) *NetworkError {
	return &NetworkError{
		IPQuorumError: IPQuorumError{Message: message, Err: err},
	}
}

// APIError represents API call failures
type APIError struct {
	IPQuorumError
	StatusCode int
	Endpoint   string
}

// NewAPIError creates a new API error
func NewAPIError(message string, statusCode int, endpoint string) *APIError {
	return &APIError{
		IPQuorumError: IPQuorumError{Message: message},
		StatusCode:    statusCode,
		Endpoint:      endpoint,
	}
}

//
