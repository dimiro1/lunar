package runner

import (
	"fmt"
	"strings"
	"testing"
)

func TestEnhanceError_NilIndexAccess(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local response = http.get("https://example.com")

  if response.statusCode == 200 then
    return { statusCode = 200 }
  end
end`

	err := fmt.Errorf("failed to execute handler: <string>:4: attempt to index a non-table object(nil) with key 'statusCode'")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	// Check that it contains line number
	if !strings.Contains(result, "line 4") {
		t.Errorf("Expected line number in error, got: %s", result)
	}

	// Check that it contains code context
	if !strings.Contains(result, "if response.statusCode == 200 then") {
		t.Errorf("Expected code context in error, got: %s", result)
	}

	// Check that it contains suggestion
	if !strings.Contains(result, "[TIP]") {
		t.Errorf("Expected suggestion in error, got: %s", result)
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_NilFunctionCall(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local data = event.body

  local result = processData(data)

  return { statusCode = 200, body = result }
end`

	err := fmt.Errorf("<string>:4: attempt to call a nil value (global 'processData')")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "line 4") {
		t.Errorf("Expected line number in error")
	}

	if !strings.Contains(result, "local result = processData(data)") {
		t.Errorf("Expected code context in error")
	}

	if !strings.Contains(result, "function that doesn't exist") {
		t.Errorf("Expected nil_call suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_BadArgument(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local num = "not a number"

  local result = math.sqrt(num)

  return { statusCode = 200 }
end`

	err := fmt.Errorf("<string>:4: bad argument #1 to 'sqrt' (number expected, got string)")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "line 4") {
		t.Errorf("Expected line number in error")
	}

	if !strings.Contains(result, "Wrong type") {
		t.Errorf("Expected bad_argument suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_SyntaxError(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local data = { key = "value"

  return { statusCode = 200 }
end`

	err := fmt.Errorf("<string>:3: unexpected symbol near 'return'")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "syntax error") || !strings.Contains(result, "Lua syntax") {
		t.Errorf("Expected syntax_error suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_MissingEnd(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  if event.method == "GET" then
    return { statusCode = 200 }
  -- missing end here
end`

	err := fmt.Errorf("<string>:5: 'end' expected (to close 'if' at line 2) near '<eof>'")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "Missing 'end'") {
		t.Errorf("Expected missing_end suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_ArithmeticOnNil(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local count = kv.get("counter")

  local newCount = count + 1

  return { statusCode = 200 }
end`

	err := fmt.Errorf("<string>:4: attempt to perform arithmetic on a nil value (local 'count')")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "line 4") {
		t.Errorf("Expected line number in error")
	}

	if !strings.Contains(result, "newCount = count + 1") {
		t.Errorf("Expected code context")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_ConcatenateNil(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local name = nil

  local message = "Hello " .. name

  return { statusCode = 200, body = message }
end`

	err := fmt.Errorf("<string>:4: attempt to concatenate a nil value (local 'name')")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "concatenate nil") {
		t.Errorf("Expected concat_error suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_NonTableIndex(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local value = "just a string"

  local prop = value.someProperty

  return { statusCode = 200 }
end`

	err := fmt.Errorf("<string>:4: attempt to index a non-table value (a string value)")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "non-table") {
		t.Errorf("Expected non_table_index suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_NoHandlerFunction(t *testing.T) {
	sourceCode := `function myHandler(ctx, event)
  return { statusCode = 200 }
end`

	err := fmt.Errorf("handler function not found in Lua code")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "export a 'handler' function") {
		t.Errorf("Expected no_handler suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_BadReturnType(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  return "just a string"
end`

	err := fmt.Errorf("handler did not return a table")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "must return a table") {
		t.Errorf("Expected bad_return suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_UndefinedVariable(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  log.info("Starting")

  local result = undefinedVariable + 10

  return { statusCode = 200 }
end`

	err := fmt.Errorf("<string>:4: attempt to perform arithmetic on a nil value (global 'undefinedVariable')")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "line 4") {
		t.Errorf("Expected line number")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_InvalidJSONAccess(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local data = json.decode(event.body)

  local value = data.nested.property.value

  return { statusCode = 200 }
end`

	err := fmt.Errorf("<string>:4: attempt to index a nil value (field 'nested')")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "line 4") {
		t.Errorf("Expected line number")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_TableAccessOnString(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local config = "string not table"

  local timeout = config.timeout

  return { statusCode = 200 }
end`

	err := fmt.Errorf("<string>:4: attempt to index a non-table object(a string value) with key 'timeout'")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "non-table") {
		t.Errorf("Expected non_table_index pattern")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_CompareError(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local a = "string"
  local b = 123

  if a < b then
    return { statusCode = 200 }
  end

  return { statusCode = 400 }
end`

	err := fmt.Errorf("<string>:5: attempt to compare string with number")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "compare") {
		t.Errorf("Expected compare_error suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_DivisionByNil(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local x = 100
  local y = nil

  local result = x / y

  return { statusCode = 200 }
end`

	err := fmt.Errorf("<string>:5: attempt to perform arithmetic on a nil value (local 'y')")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "line 5") {
		t.Errorf("Expected line number")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_NestedFunctionError(t *testing.T) {
	sourceCode := `function helper(data)
  return data.value
end

function handler(ctx, event)
  local result = helper(nil)

  return { statusCode = 200, body = result }
end`

	err := fmt.Errorf("<string>:2: attempt to index a nil value (local 'data')")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "line 2") {
		t.Errorf("Expected line number")
	}

	if !strings.Contains(result, "return data.value") {
		t.Errorf("Expected code context")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestEnhanceError_AlternateSyntaxErrorFormat(t *testing.T) {
	sourceCode := `function handler(ctx, event)
  local x = 1
  local y = 2
  local z = 3

  if x = y then
    return { statusCode = 200 }
  end
end`

	err := fmt.Errorf("<string> line:6(column:9) near '=': syntax error")
	enhanced := EnhanceError(err, sourceCode)

	result := enhanced.Error()

	if !strings.Contains(result, "line 6") {
		t.Errorf("Expected line number in error")
	}

	if !strings.Contains(result, "if x = y then") {
		t.Errorf("Expected code context")
	}

	if !strings.Contains(result, "syntax error") || !strings.Contains(result, "[TIP]") {
		t.Errorf("Expected syntax_error suggestion")
	}

	t.Logf("Enhanced error:\n%s", result)
}

func TestExtractLineNumber(t *testing.T) {
	tests := []struct {
		errMsg   string
		expected int
	}{
		{"<string>:7: some error", 7},
		{"<string>:123: another error", 123},
		{"<string>:1: first line error", 1},
		{"<string> line:6(column:33) near '=': syntax error", 6},
		{"<string> line:42(column:10) near 'end': syntax error", 42},
		{"error without line number", 0},
		{"", 0},
	}

	for _, tt := range tests {
		result := extractLineNumber(tt.errMsg)
		if result != tt.expected {
			t.Errorf("extractLineNumber(%q) = %d, expected %d", tt.errMsg, result, tt.expected)
		}
	}
}

func TestExtractCodeContext(t *testing.T) {
	sourceCode := `line 1
line 2
line 3
line 4
line 5
line 6
line 7`

	// Test middle line
	context := extractCodeContext(sourceCode, 4, 0, 1)
	if !strings.Contains(context, "> ") {
		t.Errorf("Expected error line marker '>' in context")
	}
	if !strings.Contains(context, "line 3") || !strings.Contains(context, "line 4") || !strings.Contains(context, "line 5") {
		t.Errorf("Expected surrounding lines in context, got: %s", context)
	}

	// Test first line
	context = extractCodeContext(sourceCode, 1, 0, 2)
	if !strings.Contains(context, "line 1") {
		t.Errorf("Expected first line in context")
	}

	// Test last line
	context = extractCodeContext(sourceCode, 7, 0, 2)
	if !strings.Contains(context, "line 7") {
		t.Errorf("Expected last line in context")
	}

	// Test invalid line number
	context = extractCodeContext(sourceCode, 0, 0, 1)
	if context != "" {
		t.Errorf("Expected empty context for invalid line number")
	}

	context = extractCodeContext(sourceCode, 100, 0, 1)
	if context != "" {
		t.Errorf("Expected empty context for out of range line number")
	}
}

func TestDetectErrorPattern(t *testing.T) {
	tests := []struct {
		errMsg   string
		expected string
	}{
		{"attempt to index a non-table object(nil)", "nil_index"},
		{"attempt to index a non-table value", "non_table_index"},
		{"attempt to call a nil value", "nil_call"},
		{"bad argument #1 expected string got number", "bad_argument"},
		{"unexpected symbol near '='", "syntax_error"},
		{"'end' expected to close 'if'", "missing_end"},
		{"attempt to perform arithmetic on nil", "arithmetic_error"},
		{"attempt to concatenate a nil value", "concat_error"},
		{"attempt to compare string with number", "compare_error"},
		{"handler function not found in Lua code", "no_handler"},
		{"handler did not return a table", "bad_return"},
		{"some unknown error", "unknown"},
	}

	for _, tt := range tests {
		result := detectErrorPattern(tt.errMsg)
		if result != tt.expected {
			t.Errorf("detectErrorPattern(%q) = %q, expected %q", tt.errMsg, result, tt.expected)
		}
	}
}

func TestGenerateSuggestion(t *testing.T) {
	patterns := []string{
		"nil_index",
		"nil_call",
		"bad_argument",
		"syntax_error",
		"missing_end",
		"arithmetic_error",
		"concat_error",
		"compare_error",
		"no_handler",
		"bad_return",
	}

	for _, pattern := range patterns {
		suggestion := generateSuggestion(pattern)
		if suggestion == "" {
			t.Errorf("Expected non-empty suggestion for pattern %q", pattern)
		}
		if !strings.Contains(suggestion, "[TIP]") {
			t.Errorf("Expected suggestion to contain [TIP] for pattern %q", pattern)
		}
	}

	// Test unknown pattern
	suggestion := generateSuggestion("unknown")
	if suggestion != "" {
		t.Errorf("Expected empty suggestion for unknown pattern")
	}
}
