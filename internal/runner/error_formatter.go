package runner

import (
	"fmt"
	"regexp"
	"strings"
)

// EnhanceError transforms a raw Lua error into a user-friendly error message
// with code context and actionable suggestions.
//
// The enhanced error includes:
//   - The original error message with line number
//   - Code snippet showing surrounding lines with the error line highlighted
//   - A column indicator (^) when column information is available
//   - Pattern-matched suggestions for common Lua errors
//
// Returns nil if err is nil. If line number cannot be extracted from the error,
// only the error message and suggestion (if applicable) are included.
func EnhanceError(err error, sourceCode string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Extract line number and column from error message
	lineNum := extractLineNumber(errMsg)
	colNum := extractColumnNumber(errMsg)

	// Extract code context if we have a line number
	var codeContext string
	if lineNum > 0 {
		codeContext = extractCodeContext(sourceCode, lineNum, colNum, 2)
	}

	// Detect error pattern and generate suggestion
	pattern := detectErrorPattern(errMsg)
	suggestion := generateSuggestion(pattern)

	// Format the enhanced error
	return formatEnhancedError(errMsg, lineNum, codeContext, suggestion)
}

// extractLineNumber parses the line number from Lua error messages
// Example: "<string>:7:" -> 7
// Example: "<string> line:6(column:33)" -> 6
func extractLineNumber(errMsg string) int {
	// Try format: <string>:7:
	re := regexp.MustCompile(`<string>:(\d+):`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) > 1 {
		var lineNum int
		if _, err := fmt.Sscanf(matches[1], "%d", &lineNum); err == nil {
			return lineNum
		}
	}

	// Try format: <string> line:6(column:33)
	re = regexp.MustCompile(`<string> line:(\d+)`)
	matches = re.FindStringSubmatch(errMsg)
	if len(matches) > 1 {
		var lineNum int
		if _, err := fmt.Sscanf(matches[1], "%d", &lineNum); err == nil {
			return lineNum
		}
	}

	return 0
}

// extractColumnNumber parses the column number from Lua error messages
// Example: "<string> line:6(column:33)" -> 33
func extractColumnNumber(errMsg string) int {
	re := regexp.MustCompile(`column:(\d+)`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) > 1 {
		var colNum int
		if _, err := fmt.Sscanf(matches[1], "%d", &colNum); err == nil {
			return colNum
		}
	}
	return 0
}

// extractCodeContext extracts lines around the error location
func extractCodeContext(sourceCode string, lineNum, colNum, contextLines int) string {
	if lineNum <= 0 {
		return ""
	}

	lines := strings.Split(sourceCode, "\n")
	if lineNum > len(lines) {
		return ""
	}

	start := max(0, lineNum-contextLines-1)
	end := min(len(lines), lineNum+contextLines)

	var context strings.Builder
	for i := start; i < end; i++ {
		prefix := "  "
		if i == lineNum-1 {
			prefix = "> "
		}
		context.WriteString(fmt.Sprintf("%s%3d | %s\n", prefix, i+1, lines[i]))

		// Add arrow pointing to column if this is the error line and we have a column number
		if i == lineNum-1 && colNum > 0 {
			// Calculate arrow position: prefix (2) + line number (3) + " | " (3) + column (0-indexed)
			arrowPos := 2 + 3 + 3 + colNum - 1
			arrow := strings.Repeat(" ", arrowPos) + "^"
			context.WriteString(arrow + "\n")
		}
	}

	return context.String()
}

// detectErrorPattern identifies the type of error
func detectErrorPattern(errMsg string) string {
	patterns := map[string]string{
		`attempt to index.*nil`:              "nil_index",
		`attempt to index a non-table`:       "non_table_index",
		`attempt to call.*nil`:               "nil_call",
		`bad argument.*expected.*got`:        "bad_argument",
		`unexpected symbol`:                  "syntax_error",
		`syntax error`:                       "syntax_error",
		`'end' expected`:                     "missing_end",
		`attempt to perform arithmetic`:      "arithmetic_error",
		`attempt to concatenate`:             "concat_error",
		`attempt to compare`:                 "compare_error",
		`handler function not found`:         "no_handler",
		`handler did not return a table`:     "bad_return",
	}

	for pattern, name := range patterns {
		matched, _ := regexp.MatchString(pattern, errMsg)
		if matched {
			return name
		}
	}
	return "unknown"
}

// generateSuggestion provides helpful advice based on error type
func generateSuggestion(pattern string) string {
	suggestions := map[string]string{
		"nil_index": `[TIP] A variable is nil. Common causes:
  • Function returned nil due to an error (check error returns)
  • Variable not initialized or assigned
  • http.get/http.post returned nil (network error)
  • kv.get returned nil (key doesn't exist)

  Always check before accessing:
    local value, err = someFunction()
    if err or not value then
      -- handle error
    end`,

		"non_table_index": `[TIP] Trying to access a property on a non-table value.
  • Check if the value is actually a table
  • The value might be a string, number, or nil
  • Use type() to check: if type(value) == "table" then ... end`,

		"nil_call": `[TIP] Attempting to call a function that doesn't exist.
  • Check the function name spelling
  • Make sure the function is defined before calling it
  • Verify the API is available (log, http, kv, json, etc.)`,

		"bad_argument": `[TIP] Wrong type passed to a function.
  • Check function expects (number, string, table, etc.)
  • Use type conversion: tostring(), tonumber()
  • Validate input types before passing them`,

		"syntax_error": `[TIP] Lua syntax error.
  • Check for missing commas, quotes, or parentheses
  • Ensure all strings are properly quoted
  • Verify table syntax: { key = value }`,

		"missing_end": `[TIP] Missing 'end' keyword.
  • Every 'function', 'if', 'for', 'while' needs an 'end'
  • Check that all blocks are properly closed
  • Count your 'end' keywords to match opening statements`,

		"arithmetic_error": `[TIP] Cannot perform math on non-number values.
  • Ensure both operands are numbers
  • Use tonumber() to convert strings to numbers
  • Check for nil values before arithmetic`,

		"concat_error": `[TIP] Cannot concatenate nil values with '..'.
  • Check that all values exist before concatenating
  • Use tostring() to safely convert values
  • Example: "Value: " .. tostring(myVar)`,

		"compare_error": `[TIP] Cannot compare different types.
  • Ensure both values are the same type
  • Convert types explicitly before comparing
  • Use type() to check types first`,

		"no_handler": `[TIP] Your function must export a 'handler' function.
  • Add: function handler(ctx, event) ... end
  • The handler function must be at the top level
  • Check the function name is exactly 'handler'`,

		"bad_return": `[TIP] Handler must return a table with statusCode.
  • Return format: { statusCode = 200, body = "..." }
  • statusCode is required (number)
  • body is optional (string)
  • headers is optional (table)`,
	}

	if suggestion, ok := suggestions[pattern]; ok {
		return suggestion
	}
	return ""
}

// formatEnhancedError creates a well-formatted error message
func formatEnhancedError(errMsg string, lineNum int, codeContext, suggestion string) error {
	var msg strings.Builder

	// Error message
	if lineNum > 0 {
		msg.WriteString(fmt.Sprintf("Error at line %d: %s\n", lineNum, cleanErrorMessage(errMsg)))
	} else {
		msg.WriteString(fmt.Sprintf("Error: %s\n", cleanErrorMessage(errMsg)))
	}

	// Code context with markers for frontend parsing
	if codeContext != "" {
		msg.WriteString("\n[CODE]\n")
		msg.WriteString(codeContext)
		msg.WriteString("[/CODE]\n")
	}

	// Suggestion
	if suggestion != "" {
		msg.WriteString("\n")
		msg.WriteString(suggestion)
	}

	return fmt.Errorf("%s", msg.String())
}

// cleanErrorMessage removes redundant prefixes from error messages
func cleanErrorMessage(errMsg string) string {
	// Remove common prefixes
	prefixes := []string{
		"failed to execute handler: ",
		"failed to load Lua code: ",
	}

	for _, prefix := range prefixes {
		errMsg = strings.TrimPrefix(errMsg, prefix)
	}

	return errMsg
}
