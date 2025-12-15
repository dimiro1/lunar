package time

import (
	"context"
	gotime "time"
)

// Now returns the current Unix timestamp in seconds.
func Now() int64 {
	return gotime.Now().Unix()
}

// Format formats a Unix timestamp to a string using the given layout.
// Uses Go's time format layout (e.g., "2006-01-02 15:04:05").
func Format(timestamp int64, layout string) string {
	t := gotime.Unix(timestamp, 0)
	return t.Format(layout)
}

// Parse parses a time string according to a layout.
// Returns Unix timestamp or an error.
func Parse(timeStr, layout string) (int64, error) {
	t, err := gotime.Parse(layout, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// Sleep sleeps for the specified number of milliseconds.
// Respects the provided context for cancellation.
// Returns true if sleep completed, false if cancelled.
func Sleep(ctx context.Context, milliseconds int64) bool {
	duration := gotime.Duration(milliseconds) * gotime.Millisecond

	select {
	case <-ctx.Done():
		return false
	case <-gotime.After(duration):
		return true
	}
}
