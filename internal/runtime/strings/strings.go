package strings

import gostrings "strings"

// Trim removes leading and trailing whitespace from a string.
func Trim(s string) string {
	return gostrings.TrimSpace(s)
}

// TrimLeft removes leading whitespace from a string.
func TrimLeft(s string) string {
	return gostrings.TrimLeft(s, " \t\n\r")
}

// TrimRight removes trailing whitespace from a string.
func TrimRight(s string) string {
	return gostrings.TrimRight(s, " \t\n\r")
}

// Split splits a string by a separator and returns a slice of parts.
func Split(s, sep string) []string {
	return gostrings.Split(s, sep)
}

// Join joins a slice of strings with a separator.
func Join(parts []string, sep string) string {
	return gostrings.Join(parts, sep)
}

// HasPrefix checks if a string has the given prefix.
func HasPrefix(s, prefix string) bool {
	return gostrings.HasPrefix(s, prefix)
}

// HasSuffix checks if a string has the given suffix.
func HasSuffix(s, suffix string) bool {
	return gostrings.HasSuffix(s, suffix)
}

// Replace replaces occurrences of old with new in s.
// n is the number of replacements: -1 means replace all.
func Replace(s, old, new string, n int) string {
	return gostrings.Replace(s, old, new, n)
}

// ToLower converts a string to lowercase.
func ToLower(s string) string {
	return gostrings.ToLower(s)
}

// ToUpper converts a string to uppercase.
func ToUpper(s string) string {
	return gostrings.ToUpper(s)
}

// Contains checks if a string contains a substring.
func Contains(s, substr string) bool {
	return gostrings.Contains(s, substr)
}

// Repeat repeats a string n times.
func Repeat(s string, count int) string {
	return gostrings.Repeat(s, count)
}
