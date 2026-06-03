package model

import (
	"encoding/json"
	"fmt"
	"io"
)

// StringMap is the Go type backing the GraphQL `Map` scalar: an arbitrary
// string→string mapping serialized as a JSON object. It is used for environment
// variables and key/value store entries, matching the JSON shape the REST API
// exposed today.
//
// This file is hand-written; gqlgen binds the `Map` scalar to it via gqlgen.yml
// and never regenerates it.
type StringMap map[string]string

// MarshalGQL writes the map as a compact JSON object. encoding/json sorts string
// keys, so the output is deterministic.
func (m StringMap) MarshalGQL(w io.Writer) {
	if m == nil {
		_, _ = io.WriteString(w, "{}")
		return
	}
	b, err := json.Marshal(map[string]string(m))
	if err != nil {
		_, _ = io.WriteString(w, "{}")
		return
	}
	_, _ = w.Write(b)
}

// UnmarshalGQL accepts a JSON object with string values (the shape produced when
// a `Map` is supplied as a GraphQL input) and populates the map.
func (m *StringMap) UnmarshalGQL(v any) error {
	switch val := v.(type) {
	case map[string]string:
		*m = val
	case map[string]any:
		out := make(StringMap, len(val))
		for k, raw := range val {
			s, ok := raw.(string)
			if !ok {
				return fmt.Errorf("invalid Map value for key %q: must be a string, got %T", k, raw)
			}
			out[k] = s
		}
		*m = out
	default:
		return fmt.Errorf("invalid Map: must be an object, got %T", v)
	}
	return nil
}
