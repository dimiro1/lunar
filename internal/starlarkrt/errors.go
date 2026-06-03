package starlarkrt

import (
	"fmt"
	"regexp"
	"strings"
)

// lineColRe matches the "handler.star:LINE:COL:" prefix Starlark uses in both
// syntax and evaluation error messages.
var lineColRe = regexp.MustCompile(scriptName + `:(\d+):(\d+)`)

// EnhanceError transforms a raw Starlark error into a user-friendly message with
// code context and an actionable suggestion. It mirrors runner.EnhanceError so
// both runtimes surface errors in the same shape. Returns nil if err is nil.
func EnhanceError(err error, sourceCode string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	lineNum, colNum := extractLineCol(errMsg)

	var codeContext string
	if lineNum > 0 {
		codeContext = extractCodeContext(sourceCode, lineNum, colNum, 2)
	}

	suggestion := generateSuggestion(detectErrorPattern(errMsg))

	return formatEnhancedError(cleanErrorMessage(errMsg), lineNum, codeContext, suggestion)
}

// extractLineCol parses the line and column from a Starlark error message.
func extractLineCol(errMsg string) (line, col int) {
	matches := lineColRe.FindStringSubmatch(errMsg)
	if len(matches) < 3 {
		return 0, 0
	}
	_, _ = fmt.Sscanf(matches[1], "%d", &line)
	_, _ = fmt.Sscanf(matches[2], "%d", &col)
	return line, col
}

// extractCodeContext renders the lines around the error with the error line
// marked and (when known) a caret under the offending column.
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
		fmt.Fprintf(&context, "%s%3d | %s\n", prefix, i+1, lines[i])

		if i == lineNum-1 && colNum > 0 {
			arrowPos := 2 + 3 + 3 + colNum - 1
			context.WriteString(strings.Repeat(" ", arrowPos) + "^\n")
		}
	}

	return context.String()
}

// detectErrorPattern classifies common Starlark errors so we can attach a tip.
func detectErrorPattern(errMsg string) string {
	patterns := []struct {
		regex string
		name  string
	}{
		{`Starlark computation cancelled`, "timeout"},
		{`handler function not found`, "no_handler"},
		{`handler did not return`, "bad_return"},
		{`undefined: `, "undefined"},
		{`not callable`, "not_callable"},
		{`unhashable`, "unhashable"},
		{`has no .* field or method`, "no_field"},
		{`got .* want`, "syntax_error"},
		{`unexpected (indent|EOF|token)`, "syntax_error"},
		{`invalid syntax`, "syntax_error"},
		{`unsupported (binary|operand)`, "type_error"},
		{`index out of range`, "index_error"},
		{`key .* not in`, "key_error"},
	}

	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p.regex, errMsg); matched {
			return p.name
		}
	}
	return "unknown"
}

// generateSuggestion returns a tip for a classified error pattern.
func generateSuggestion(pattern string) string {
	suggestions := map[string]string{
		"timeout": `[TIP] Execution exceeded the time limit.
  • Avoid unbounded loops and long-running waits
  • Move heavy work behind cached results or external services`,

		"no_handler": `[TIP] Your function must define a 'handler' function.
  • Add: def handler(ctx, event): ...
  • The handler must be defined at the top level
  • Check the name is exactly 'handler'`,

		"bad_return": `[TIP] Handler must return a response dict.
  • Return format: {"statusCode": 200, "body": "..."}
  • statusCode is optional (defaults to 200)
  • body and headers are optional`,

		"undefined": `[TIP] Referenced a name that does not exist.
  • Check the spelling of the variable or function
  • Assign the variable before using it
  • Available modules: log, kv, env, http, json, crypto, time, strings, random, url, router, ai, email`,

		"not_callable": `[TIP] Tried to call something that is not a function.
  • Check you are calling a function, not indexing it
  • Module functions live under the module: http.get(...), kv.set(...)`,

		"no_field": `[TIP] Accessed a field or method that does not exist.
  • ctx exposes: executionId, functionId, version, requestId, baseUrl
  • event exposes: method, path, relativePath, body, headers, query
  • Use event.headers["Name"] for header values`,

		"syntax_error": `[TIP] Starlark syntax error.
  • Starlark is a Python dialect: use ':' and indentation for blocks
  • Use 'def handler(ctx, event):', not 'function handler(...)'
  • Strings need quotes; dict literals use {"key": value}`,

		"type_error": `[TIP] Operation applied to incompatible types.
  • Convert explicitly: str(x), int(x), float(x)
  • You cannot concatenate a string with a number directly`,

		"index_error": `[TIP] List index out of range.
  • Check the list length with len(x) before indexing
  • Remember indices are 0-based in Starlark`,

		"key_error": `[TIP] Missing dict key.
  • Use x.get("key") to read safely (returns None when absent)
  • Check the key exists: if "key" in x: ...`,

		"unhashable": `[TIP] Used an unhashable value as a dict key or set member.
  • Only strings, numbers, booleans and tuples can be keys`,
	}

	return suggestions[pattern]
}

// formatEnhancedError assembles the final, formatted error message.
func formatEnhancedError(errMsg string, lineNum int, codeContext, suggestion string) error {
	var msg strings.Builder

	if lineNum > 0 {
		fmt.Fprintf(&msg, "Error at line %d: %s\n", lineNum, errMsg)
	} else {
		fmt.Fprintf(&msg, "Error: %s\n", errMsg)
	}

	if codeContext != "" {
		msg.WriteString("\n[CODE]\n")
		msg.WriteString(codeContext)
		msg.WriteString("[/CODE]\n")
	}

	if suggestion != "" {
		msg.WriteString("\n")
		msg.WriteString(suggestion)
	}

	return fmt.Errorf("%s", msg.String())
}

// cleanErrorMessage strips the internal wrapping prefixes from the raw message.
func cleanErrorMessage(errMsg string) string {
	prefixes := []string{
		"failed to execute handler: ",
		"failed to load Starlark code: ",
	}
	for _, prefix := range prefixes {
		errMsg = strings.TrimPrefix(errMsg, prefix)
	}
	return errMsg
}
