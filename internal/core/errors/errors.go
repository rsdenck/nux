package errors

import (
	"errors"
	"fmt"
)

var (
	// Standard error codes
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrTimeout           = errors.New("operation timed out")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternal          = errors.New("internal error")
	ErrNotImplemented    = errors.New("not implemented")
	ErrUnsupported       = errors.New("unsupported operation")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrConnectionRefused = errors.New("connection refused")
	ErrNetwork           = errors.New("network error")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// ErrorCode represents a structured error code
type ErrorCode string

const (
	// General errors
	ErrCodeGeneral        ErrorCode = "GENERAL_ERROR"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeTimeout        ErrorCode = "TIMEOUT"
	ErrCodeInvalidInput   ErrorCode = "INVALID_INPUT"
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodePermission     ErrorCode = "PERMISSION_DENIED"
	ErrCodeAlreadyExists  ErrorCode = "ALREADY_EXISTS"
	ErrCodeNetwork        ErrorCode = "NETWORK_ERROR"

	// Service errors
	ErrCodeServiceStart   ErrorCode = "SERVICE_START_FAILED"
	ErrCodeServiceStop    ErrorCode = "SERVICE_STOP_FAILED"
	ErrCodeServiceRestart ErrorCode = "SERVICE_RESTART_FAILED"

	// Package errors
	ErrCodePkgInstall ErrorCode = "PKG_INSTALL_FAILED"
	ErrCodePkgRemove  ErrorCode = "PKG_REMOVE_FAILED"
	ErrCodePkgUpdate  ErrorCode = "PKG_UPDATE_FAILED"

	// Network errors
	ErrCodeNetworkConfig ErrorCode = "NETWORK_CONFIG_FAILED"
	ErrCodeFirewall      ErrorCode = "FIREWALL_ERROR"

	// Disk errors
	ErrCodeDiskRead  ErrorCode = "DISK_READ_FAILED"
	ErrCodeDiskWrite ErrorCode = "DISK_WRITE_FAILED"
	ErrCodeLVM       ErrorCode = "LVM_ERROR"

	// Security errors
	ErrCodeAudit   ErrorCode = "AUDIT_ERROR"
	ErrCodeSecurityScan ErrorCode = "SECURITY_SCAN_FAILED"

	// Container errors
	ErrCodeContainerStart ErrorCode = "CONTAINER_START_FAILED"
	ErrCodeContainerStop  ErrorCode = "CONTAINER_STOP_FAILED"

	// SSH errors
	ErrCodeSSHConnect ErrorCode = "SSH_CONNECT_FAILED"
	ErrCodeSSHAuth    ErrorCode = "SSH_AUTH_FAILED"
)

// NuxError represents a structured NUX error
type NuxError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *NuxError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *NuxError) Unwrap() error {
	return e.Err
}

// New creates a new NuxError
func New(code ErrorCode, message string) *NuxError {
	return &NuxError{
		Code:    code,
		Message: message,
	}
}

// Wrap creates a new NuxError wrapping an existing error
func Wrap(err error, code ErrorCode, message string) *NuxError {
	return &NuxError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf creates a new NuxError with formatted message
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *NuxError {
	return &NuxError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// Is checks if the error matches the given error
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Helper functions for common errors

func NewNotFoundError(resource string) *NuxError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

func NewUnauthorizedError(message string) *NuxError {
	return New(ErrCodeUnauthorized, message)
}

func NewTimeoutError(operation string) *NuxError {
	return New(ErrCodeTimeout, fmt.Sprintf("%s timed out", operation))
}

func NewInvalidInputError(field string, message string) *NuxError {
	return New(ErrCodeInvalidInput, fmt.Sprintf("invalid %s: %s", field, message))
}

func NewPermissionError(resource string) *NuxError {
	return New(ErrCodePermission, fmt.Sprintf("permission denied for %s", resource))
}

func NewServiceError(code ErrorCode, service string, operation string) *NuxError {
	return New(code, fmt.Sprintf("failed to %s service %s", operation, service))
}

func NewNetworkError(message string) *NuxError {
	return New(ErrCodeNetwork, message)
}