// Package core provides core functionality for NUX CLI.
//
// This package includes:
//   - Executor: Interface for executing system commands
//   - RealExecutor: Implementation that executes real system commands
//   - SanitizeInput: Sanitize user input to prevent shell injection
//   - ValidatePath: Validate file paths for security
//
// Example usage:
//
//	executor := &core.RealExecutor{}
//	result, err := executor.Run("ls", "-la")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result)
package core