// Package output provides structured output formatting for NUX CLI.
//
// This package includes:
//   - Output: Structured output with status, data, error fields
//   - NewSuccess: Create success output
//   - NewError: Create error output with error code
//   - NewInfo: Create info output
//   - NewList: Create list output with pagination
//   - PrintCompactTable: Print formatted tables
//
// Example usage:
//
//	output.NewSuccess(map[string]interface{}{
//	    "status": "ok",
//	    "count":  10,
//	}).Print()
//
//	Output in JSON mode:
//	{
//	  "status": "success",
//	  "data": {
//	    "status": "ok",
//	    "count": 10
//	  }
//	}
package output