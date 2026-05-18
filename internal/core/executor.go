package core

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const DefaultTimeout = 30 * time.Second

// Executor interface defines execution methods
type Executor interface {
	Run(name string, args ...string) (string, error)
	RunSilent(name string, args ...string) error
	CombinedOutput(name string, args ...string) (string, error)
	RunWithContext(ctx context.Context, name string, args ...string) (string, error)
}

// RealExecutor executes real system commands
type RealExecutor struct{}

func (r *RealExecutor) Run(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	return r.RunWithContext(ctx, name, args...)
}

func (r *RealExecutor) RunSilent(name string, args ...string) error {
	_, err := r.Run(name, args...)
	return err
}

func (r *RealExecutor) CombinedOutput(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	
	var cmd *exec.Cmd
	if len(args) > 0 {
		cmd = exec.CommandContext(ctx, name, args...)
	} else {
		cmd = exec.CommandContext(ctx, name)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command %s timed out after %v", name, DefaultTimeout)
		}
		return "", fmt.Errorf("command %s failed: %w - %s", name, err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *RealExecutor) RunWithContext(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command %s timed out after %v", name, DefaultTimeout)
		}
		return "", fmt.Errorf("command %s failed: %w", name, err)
	}
	return strings.TrimSpace(string(output)), nil
}



// SanitizeInput prevents shell injection attacks
func SanitizeInput(input string) string {
	dangerous := []string{";", "&&", "||", "`", "$(", "${", "|", ">", "<", "\n", "\r"}
	result := input
	for _, d := range dangerous {
		result = strings.ReplaceAll(result, d, "")
	}
	return strings.TrimSpace(result)
}

// ValidatePath ensures path is safe
func ValidatePath(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}
	if strings.Contains(path, "..") {
		return false
	}
	return true
}

// ValidateCommand ensures command contains no dangerous characters
func ValidateCommand(cmd string) bool {
	dangerous := []string{";", "&&", "||", "`", "$(", "${", "|", ">", "<", "\n", "\r", "'", "\""}
	for _, d := range dangerous {
		if strings.Contains(cmd, d) {
			return false
		}
	}
	return true
}
