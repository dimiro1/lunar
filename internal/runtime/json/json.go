package json

import gojson "encoding/json"

// Encode converts a Go value to a JSON string.
func Encode(v any) (string, error) {
	jsonBytes, err := gojson.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Decode converts a JSON string to a Go value.
// Returns the decoded value which will be one of:
// - nil for JSON null
// - bool for JSON boolean
// - float64 for JSON numbers
// - string for JSON strings
// - []any for JSON arrays
// - map[string]any for JSON objects
func Decode(jsonStr string) (any, error) {
	var v any
	if err := gojson.Unmarshal([]byte(jsonStr), &v); err != nil {
		return nil, err
	}
	return v, nil
}
