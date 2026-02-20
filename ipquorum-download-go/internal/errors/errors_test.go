// Package errors_test provides comprehensive unit tests for the errors package,
// covering all error types and their behaviors.
package errors

import (
	"errors"
	"testing"
)

func TestIPQuorumError(t *testing.T) {
	t.Run("error without wrapped error", func(t *testing.T) {
		err := &IPQuorumError{
			Message: "test error",
		}
		expected := "test error"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("error with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped error")
		err := &IPQuorumError{
			Message: "test error",
			Err:     wrappedErr,
		}
		expected := "test error: wrapped error"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped error")
		err := &IPQuorumError{
			Message: "test error",
			Err:     wrappedErr,
		}
		if err.Unwrap() != wrappedErr {
			t.Errorf("Unwrap() did not return the wrapped error")
		}
	})

	t.Run("unwrap returns nil when no wrapped error", func(t *testing.T) {
		err := &IPQuorumError{
			Message: "test error",
		}
		if err.Unwrap() != nil {
			t.Errorf("Unwrap() should return nil when no wrapped error")
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Run("create validation error", func(t *testing.T) {
		err := NewValidationError("invalid configuration")
		expected := "invalid configuration"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("validation error contains IPQuorumError", func(t *testing.T) {
		err := NewValidationError("test")
		// ValidationError embeds IPQuorumError, so we can access it directly
		if err.Message != "test" {
			t.Errorf("ValidationError.Message = %q, want %q", err.Message, "test")
		}
	})
}

func TestAuthenticationError(t *testing.T) {
	t.Run("create authentication error", func(t *testing.T) {
		err := NewAuthenticationError("authentication failed", 401)
		expected := "authentication failed"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
		if err.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", err.StatusCode)
		}
	})

	t.Run("authentication error contains IPQuorumError", func(t *testing.T) {
		err := NewAuthenticationError("test", 401)
		// AuthenticationError embeds IPQuorumError, so we can access it directly
		if err.Message != "test" {
			t.Errorf("AuthenticationError.Message = %q, want %q", err.Message, "test")
		}
	})
}

func TestNetworkError(t *testing.T) {
	t.Run("create network error", func(t *testing.T) {
		wrappedErr := errors.New("connection refused")
		err := NewNetworkError("network failure", wrappedErr)
		expected := "network failure: connection refused"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("network error unwraps correctly", func(t *testing.T) {
		wrappedErr := errors.New("connection refused")
		err := NewNetworkError("network failure", wrappedErr)
		if err.Unwrap() != wrappedErr {
			t.Error("NetworkError should unwrap to the original error")
		}
	})

	t.Run("network error contains IPQuorumError", func(t *testing.T) {
		err := NewNetworkError("test", nil)
		// NetworkError embeds IPQuorumError, so we can access it directly
		if err.Message != "test" {
			t.Errorf("NetworkError.Message = %q, want %q", err.Message, "test")
		}
	})
}

func TestAPIError(t *testing.T) {
	t.Run("create API error", func(t *testing.T) {
		err := NewAPIError("API call failed", 500, "/api/v1/endpoint")
		expected := "API call failed"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
		if err.StatusCode != 500 {
			t.Errorf("StatusCode = %d, want 500", err.StatusCode)
		}
		if err.Endpoint != "/api/v1/endpoint" {
			t.Errorf("Endpoint = %q, want %q", err.Endpoint, "/api/v1/endpoint")
		}
	})

	t.Run("API error contains IPQuorumError", func(t *testing.T) {
		err := NewAPIError("test", 500, "/test")
		// APIError embeds IPQuorumError, so we can access it directly
		if err.Message != "test" {
			t.Errorf("APIError.Message = %q, want %q", err.Message, "test")
		}
	})
}

// Test error type assertions
func TestErrorTypeAssertions(t *testing.T) {
	t.Run("can distinguish error types", func(t *testing.T) {
		validationErr := NewValidationError("validation failed")
		authErr := NewAuthenticationError("auth failed", 401)
		networkErr := NewNetworkError("network failed", nil)
		apiErr := NewAPIError("api failed", 500, "/test")

		// Test ValidationError
		var ve *ValidationError
		if !errors.As(validationErr, &ve) {
			t.Error("Should be able to assert ValidationError")
		}

		// Test AuthenticationError
		var ae *AuthenticationError
		if !errors.As(authErr, &ae) {
			t.Error("Should be able to assert AuthenticationError")
		}

		// Test NetworkError
		var ne *NetworkError
		if !errors.As(networkErr, &ne) {
			t.Error("Should be able to assert NetworkError")
		}

		// Test APIError
		var apie *APIError
		if !errors.As(apiErr, &apie) {
			t.Error("Should be able to assert APIError")
		}
	})
}

// Code generated with assistance from IBM Bob
