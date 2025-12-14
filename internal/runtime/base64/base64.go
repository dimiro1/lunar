package base64

import gobase64 "encoding/base64"

// Encode encodes a string to base64.
func Encode(s string) string {
	return gobase64.StdEncoding.EncodeToString([]byte(s))
}

// Decode decodes a base64 string.
// Returns the decoded string or an error.
func Decode(s string) (string, error) {
	decoded, err := gobase64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
