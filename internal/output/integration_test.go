package output

import (
	"testing"
)

func TestOutputJSONFormat(t *testing.T) {
	SetFormat(true, false)
	defer SetFormat(false, false)

	result := NewSuccess(map[string]interface{}{
		"key":   "value",
		"count": 42,
	})

	// This should not panic
	result.Print()
}

func TestTableFormatting(t *testing.T) {
	headers := []string{"NAME", "STATUS", "COUNT"}
	rows := [][]string{
		{"item1", "active", "10"},
		{"item2", "inactive", "20"},
		{"item3", "active", "30"},
	}

	// This should not panic
	PrintTable(headers, rows)
}

func TestCompactTableFormatting(t *testing.T) {
	headers := []string{"KEY", "VALUE"}
	rows := [][]string{
		{"status", "ok"},
		{"count", "100"},
		{"message", "success"},
	}

	// This should not panic
	PrintCompactTable(headers, rows)
}

func TestOutputWithLargeData(t *testing.T) {
	items := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]interface{}{
			"id":    i,
			"name":  "item" + string(rune(i)),
			"value": i * 10,
		}
	}

	result := NewList(items, len(items))
	result.WithMessage("Large list test")

	// This should not panic
	result.Print()
}

func TestErrorOutput(t *testing.T) {
	err := NewError("Test error message", "TEST_ERROR_CODE")
	err.WithMessage("Additional context")

	// This should not panic
	err.Print()
}

func TestInfoOutput(t *testing.T) {
	info := NewInfo(map[string]interface{}{
		"status":  "running",
		"pid":     1234,
		"uptime":  "1h 23m",
	})

	// This should not panic
	info.Print()
}

func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			input:  "\033[32mgreen text\033[0m",
			output: "green text",
		},
		{
			input:  "plain text",
			output: "plain text",
		},
		{
			input:  "\033[1mbold\033[22m normal",
			output: "bold normal",
		},
	}

	for _, tt := range tests {
		result := stripAnsi(tt.input)
		if result != tt.output {
			t.Errorf("stripAnsi(%q) = %q, want %q", tt.input, result, tt.output)
		}
	}
}

func TestGetKeys(t *testing.T) {
	m := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	keys := getKeys(m)
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
}

func TestPrintSuccessMessage(t *testing.T) {
	// This should not panic
	PrintSuccessMessage("Operation completed successfully")
}

func TestPrintWarningMessage(t *testing.T) {
	// This should not panic
	PrintWarningMessage("Warning: low disk space")
}

func TestPrintErrorMessage(t *testing.T) {
	// This should not panic
	PrintErrorMessage("Error: connection failed")
}

func TestPrintInfoMessage(t *testing.T) {
	// This should not panic
	PrintInfoMessage("Processing...")
}