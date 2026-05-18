package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

var jsonOutput bool

func SetFormat(json bool, yaml bool) {
	jsonOutput = json
}

type Output struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Total   int         `json:"total,omitempty"`
	Items   interface{} `json:"items,omitempty"`
}

func NewSuccess(data interface{}) *Output {
	return &Output{
		Status: "success",
		Data:   data,
	}
}

func NewInfo(data interface{}) *Output {
	return &Output{
		Status: "info",
		Data:   data,
	}
}

func NewError(err string, code string) *Output {
	return &Output{
		Status: "error",
		Error:  err,
		Code:   code,
	}
}

func NewList(items interface{}, total int) *Output {
	return &Output{
		Status: "success",
		Total:  total,
		Items:  items,
	}
}

func (o *Output) WithMessage(msg string) *Output {
	o.Message = msg
	return o
}

func (o *Output) Print() {
	if jsonOutput {
		data, _ := json.MarshalIndent(o, "", "  ")
		fmt.Println(string(data))
		return
	}

	if o.Status == "error" {
		fmt.Printf("✖ %s\n", o.Error)
		return
	}

	if o.Status == "info" {
		fmt.Printf("⚠ %s\n", o.Message)
		if o.Data != nil {
			printDataAsFormattedTable(o.Data)
		}
		return
	}

	// Success output with formatted table
	if o.Items != nil {
		printItemsAsFormattedTable(o.Items, o.Total, o.Message)
	} else if o.Data != nil {
		printDataAsFormattedTable(o.Data)
	}
}

// printItemsAsFormattedTable prints items in the exact format from output.md
func printItemsAsFormattedTable(items interface{}, total int, message string) {
	// Convert to []map[string]interface{}
	if slice, ok := items.([]map[string]interface{}); ok {
		if len(slice) == 0 {
			fmt.Println("No items found")
			return
		}

		// Get headers from first item
		headers := make([]string, 0)
		for k := range slice[0] {
			headers = append(headers, strings.ToUpper(k))
		}

		// Calculate column widths (minimum: header length + 2)
		widths := make([]int, len(headers))
		for i, h := range headers {
			widths[i] = len(h) + 2
		}

		// Adjust widths based on data
		for _, item := range slice {
			keys := getKeys(item)
			for j, k := range keys {
				val := fmt.Sprintf("%v", item[k])
				if len(val)+2 > widths[j] {
					widths[j] = len(val) + 2
				}
			}
		}

		// Print header with message
		if message != "" {
			fmt.Printf("\n%s\n", message)
		}

		// Print top border
		printTopBorder(widths)

		// Print header row
		printRow(headers, widths)

		// Print separator
		printSeparator(widths)

		// Print rows
		for _, item := range slice {
			keys := getKeys(item)
			values := make([]string, len(keys))
			for j, k := range keys {
				values[j] = fmt.Sprintf("%v", item[k])
			}
			printRow(values, widths)
		}

		// Print bottom border
		printBottomBorder(widths)

		fmt.Printf("\n%d items found\n", total)
	}
}

// printDataAsFormattedTable prints a map as a simple key-value table
func printDataAsFormattedTable(data interface{}) {
	if m, ok := data.(map[string]interface{}); ok {
		headers := []string{"KEY", "VALUE"}
		// Calculate widths based on content
		keyWidth := len(headers[0])
		valWidth := len(headers[1])
		for k, v := range m {
			keyLen := len(k)
			valLen := len(fmt.Sprintf("%v", v))
			if keyLen > keyWidth {
				keyWidth = keyLen
			}
			if valLen > valWidth {
				valWidth = valLen
			}
		}
		// Add padding
		keyWidth += 2
		valWidth += 2
		widths := []int{keyWidth, valWidth}

		printTopBorder(widths)
		printRow(headers, widths)
		printSeparator(widths)

		for k, v := range m {
			printRow([]string{k, fmt.Sprintf("%v", v)}, widths)
		}

		printBottomBorder(widths)
	} else {
		// Fallback to JSON
		d, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(d))
	}
}

func printTopBorder(widths []int) {
	fmt.Print("┌")
	for i, w := range widths {
		fmt.Print(strings.Repeat("─", w))
		if i < len(widths)-1 {
			fmt.Print("┬")
		}
	}
	fmt.Println("┐")
}

func printSeparator(widths []int) {
	fmt.Print("├")
	for i, w := range widths {
		fmt.Print(strings.Repeat("─", w))
		if i < len(widths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")
}

func printBottomBorder(widths []int) {
	fmt.Print("└")
	for i, w := range widths {
		fmt.Print(strings.Repeat("─", w))
		if i < len(widths)-1 {
			fmt.Print("┴")
		}
	}
	fmt.Println("┘")
}

func printRow(values []string, widths []int) {
	fmt.Print("│")
	for i, v := range values {
		fmt.Printf(" %-*s ", widths[i]-2, v)
		fmt.Print("│")
	}
	fmt.Println()
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// PrintTable is a simple helper for non-JSON table output
func PrintTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h) + 2
	}

	for _, row := range rows {
		for i, cell := range row {
			if len(cell)+2 > widths[i] {
				widths[i] = len(cell) + 2
			}
		}
	}

	printTopBorder(widths)
	printRow(headers, widths)
	printSeparator(widths)
	for _, row := range rows {
		printRow(row, widths)
	}
	printBottomBorder(widths)
}

// PrintCompactTable prints a compact table that fits content
func PrintCompactTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate minimal widths based on content (NO truncation)
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h) + 2
	}

	for _, row := range rows {
		for i, cell := range row {
			cleanCell := stripAnsi(cell)
			if len(cleanCell)+2 > widths[i] {
				widths[i] = len(cleanCell) + 2
			}
		}
	}

	// NO width capping - allow table to be as wide as needed
	printTopBorder(widths)
	printRow(headers, widths)
	printSeparator(widths)
	for _, row := range rows {
		printRow(row, widths)
	}
	printBottomBorder(widths)
}

// stripAnsi removes ANSI color codes from string
func stripAnsi(s string) string {
	result := ""
	inEscape := false
	for _, c := range s {
		if c == '\033' {
			inEscape = true
			continue
		}
		if inEscape && (c == 'm' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z') {
			if c == 'm' {
				inEscape = false
			}
			continue
		}
		if !inEscape {
			result += string(c)
		}
	}
	return result
}

// Helper functions for commands
func PrintSuccessMessage(msg string) {
	fmt.Printf("✔ %s\n", msg)
}

func PrintWarningMessage(msg string) {
	fmt.Printf("⚠ %s\n", msg)
}

func PrintErrorMessage(msg string) {
	fmt.Printf("✖ %s\n", msg)
}

func PrintInfoMessage(msg string) {
	fmt.Printf("⠋ %s\n", msg)
}
