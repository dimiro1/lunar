package url

import gourl "net/url"

// ParsedURL represents a parsed URL with its components.
type ParsedURL struct {
	Scheme   string
	Host     string
	Path     string
	Fragment string
	Query    map[string][]string
	Username string
	Password string
}

// Parse parses a URL string into its components.
func Parse(rawURL string) (*ParsedURL, error) {
	parsedURL, err := gourl.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	result := &ParsedURL{
		Scheme:   parsedURL.Scheme,
		Host:     parsedURL.Host,
		Path:     parsedURL.Path,
		Fragment: parsedURL.Fragment,
		Query:    make(map[string][]string),
	}

	// Parse query parameters
	if parsedURL.RawQuery != "" {
		result.Query = parsedURL.Query()
	}

	// Add username and password if present
	if parsedURL.User != nil {
		result.Username = parsedURL.User.Username()
		if password, ok := parsedURL.User.Password(); ok {
			result.Password = password
		}
	}

	return result, nil
}

// Encode URL-encodes a string.
func Encode(s string) string {
	return gourl.QueryEscape(s)
}

// Decode URL-decodes a string.
// Returns the decoded string or an error.
func Decode(s string) (string, error) {
	return gourl.QueryUnescape(s)
}
