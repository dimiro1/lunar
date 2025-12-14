package router

import "strings"

// MatchResult contains the result of a path match operation.
type MatchResult struct {
	Matched bool
	Params  map[string]string
}

// Match checks if a path matches a pattern and returns match result with extracted parameters.
// Pattern syntax:
//   - :name captures a path segment into params["name"]
//   - * at the end matches any remaining path segments
func Match(path, pattern string) MatchResult {
	params := make(map[string]string)

	path = strings.TrimSuffix(path, "/")
	pattern = strings.TrimSuffix(pattern, "/")
	if path == "" {
		path = "/"
	}
	if pattern == "" {
		pattern = "/"
	}

	pathSegments := SplitPath(path)
	patternSegments := SplitPath(pattern)
	hasWildcard := len(patternSegments) > 0 && patternSegments[len(patternSegments)-1] == "*"

	if hasWildcard {
		patternSegments = patternSegments[:len(patternSegments)-1]
		if len(pathSegments) <= len(patternSegments) {
			return MatchResult{Matched: false, Params: nil}
		}
	} else if len(pathSegments) != len(patternSegments) {
		return MatchResult{Matched: false, Params: nil}
	}

	for i, patternSeg := range patternSegments {
		pathSeg := pathSegments[i]
		if strings.HasPrefix(patternSeg, ":") {
			params[patternSeg[1:]] = pathSeg
		} else if pathSeg != patternSeg {
			return MatchResult{Matched: false, Params: nil}
		}
	}

	return MatchResult{Matched: true, Params: params}
}

// SplitPath splits a path into non-empty segments.
func SplitPath(path string) []string {
	parts := strings.Split(path, "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

// BuildPath substitutes parameter placeholders in a pattern with values from params.
// Example: BuildPath("/users/:id", map[string]string{"id": "42"}) returns "/users/42"
func BuildPath(pattern string, params map[string]string) string {
	if len(params) == 0 {
		return pattern
	}
	result := pattern
	for key, value := range params {
		result = strings.ReplaceAll(result, ":"+key, value)
	}
	return result
}

// FunctionPath builds a full path for a function with the given pattern and parameters.
// Returns "/fn/{functionID}{path}"
func FunctionPath(functionID, pattern string, params map[string]string) string {
	return "/fn/" + functionID + BuildPath(pattern, params)
}

// FunctionURL builds a full URL for a function with the given pattern and parameters.
// Returns "{baseURL}/fn/{functionID}{path}"
func FunctionURL(baseURL, functionID, pattern string, params map[string]string) string {
	return strings.TrimSuffix(baseURL, "/") + "/fn/" + functionID + BuildPath(pattern, params)
}
