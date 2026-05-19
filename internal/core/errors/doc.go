// Package errors provides structured error handling for NUX CLI.
//
// This package includes:
//   - NuxError: Structured error with code, message, and cause
//   - ErrorCode: Type for error codes
//   - Standard error codes: ErrNotFound, ErrUnauthorized, ErrTimeout, etc.
//   - Helper functions: New, Wrap, Wrapf, Is, As
//
// Error codes are organized by category:
//   - General: ErrCodeGeneral, ErrCodeNotFound, ErrCodeUnauthorized
//   - Service: ErrCodeServiceStart, ErrCodeServiceStop
//   - Package: ErrCodePkgInstall, ErrCodePkgRemove
//   - Network: ErrCodeNetworkConfig, ErrCodeFirewall
//   - Disk: ErrCodeDiskRead, ErrCodeDiskWrite, ErrCodeLVM
//   - Security: ErrCodeAudit, ErrCodeSecurityScan
//
// Example usage:
//
//	err := errors.New(errors.ErrCodeNotFound, "user not found")
//	if errors.Is(err, errors.ErrNotFound) {
//	    // Handle not found error
//	}
//
//	err := errors.Wrap(originalErr, errors.ErrCodeInternal, "failed to process request")
package errors