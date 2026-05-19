// Package logger provides structured logging for NUX CLI.
//
// This package includes:
//   - Options: Logger configuration options
//   - Init: Initialize global logger with options
//   - ConsoleHandler: Handler for console output with color support
//   - MultiHandler: Dispatch to multiple handlers
//   - Log levels: Debug, Info, Warn, Error
//
// The logger supports:
//   - Console output with levels (INFO, WARN, ERROR)
//   - File output in JSON format for detailed logging
//   - Multi-writer for both console and file
//   - Color support (configurable)
//
// Example usage:
//
//	logger.Init(logger.Options{
//	    Debug:    true,
//	    UseColor: true,
//	    LogFile:  "/var/log/nux.log",
//	})
//
//	logger.Info("Starting NUX CLI")
//	logger.Warn("Low disk space")
//	logger.Error("Failed to connect", "error", err)
package logger