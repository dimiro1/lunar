package time

import (
	"context"
	gostrings "strings"
	"testing"
	gotime "time"
)

func TestNow(t *testing.T) {
	before := gotime.Now().Unix()
	result := Now()
	after := gotime.Now().Unix()

	if result < before || result > after {
		t.Errorf("Now() = %d, expected between %d and %d", result, before, after)
	}
}

func TestFormat(t *testing.T) {
	// Test that Format produces valid output (timezone-agnostic)
	timestamp := int64(1609459200) // 2021-01-01 00:00:00 UTC

	t.Run("date format", func(t *testing.T) {
		result := Format(timestamp, "2006-01-02")
		// Just check that it produces a valid date format
		if len(result) != 10 {
			t.Errorf("Format(%d, date) = %q, expected YYYY-MM-DD format", timestamp, result)
		}
	})

	t.Run("custom format", func(t *testing.T) {
		result := Format(timestamp, "Jan 2, 2006")
		// Should contain "2021"
		if !gostrings.Contains(result, "2021") {
			t.Errorf("Format(%d, custom) = %q, expected to contain 2021", timestamp, result)
		}
	})

	t.Run("time format", func(t *testing.T) {
		result := Format(timestamp, "15:04:05")
		// Should be a valid time format HH:MM:SS
		if len(result) != 8 {
			t.Errorf("Format(%d, time) = %q, expected HH:MM:SS format", timestamp, result)
		}
	})
}

func TestParse(t *testing.T) {
	t.Run("valid date", func(t *testing.T) {
		result, err := Parse("2021-01-01", "2006-01-02")
		if err != nil {
			t.Errorf("Parse returned error: %v", err)
			return
		}
		// Check it's a reasonable timestamp (around Jan 2021)
		if result < 1609400000 || result > 1609600000 {
			t.Errorf("Parse returned unexpected timestamp: %d", result)
		}
	})

	t.Run("custom format", func(t *testing.T) {
		result, err := Parse("Jan 1, 2021", "Jan 2, 2006")
		if err != nil {
			t.Errorf("Parse returned error: %v", err)
			return
		}
		// Check it's a reasonable timestamp
		if result < 1609400000 || result > 1609600000 {
			t.Errorf("Parse returned unexpected timestamp: %d", result)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := Parse("not-a-date", "2006-01-02")
		if err == nil {
			t.Error("Parse expected error, got nil")
		}
	})

	t.Run("mismatched layout", func(t *testing.T) {
		_, err := Parse("2021-01-01", "Jan 2, 2006")
		if err == nil {
			t.Error("Parse expected error, got nil")
		}
	})
}

func TestSleep(t *testing.T) {
	t.Run("completes successfully", func(t *testing.T) {
		ctx := context.Background()
		start := gotime.Now()
		completed := Sleep(ctx, 50)
		elapsed := gotime.Since(start)

		if !completed {
			t.Error("Sleep returned false, expected true")
		}
		if elapsed < 50*gotime.Millisecond {
			t.Errorf("Sleep returned too quickly: %v", elapsed)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		completed := Sleep(ctx, 1000)
		if completed {
			t.Error("Sleep returned true on cancelled context, expected false")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*gotime.Millisecond)
		defer cancel()

		start := gotime.Now()
		completed := Sleep(ctx, 1000) // Try to sleep for 1 second
		elapsed := gotime.Since(start)

		if completed {
			t.Error("Sleep returned true, expected false due to timeout")
		}
		if elapsed > 100*gotime.Millisecond {
			t.Errorf("Sleep took too long to respond to timeout: %v", elapsed)
		}
	})
}

func TestRoundTrip(t *testing.T) {
	// Test that formatting and parsing are consistent
	// Use date-only layout to avoid timezone issues
	layout := "2006-01-02"
	timestamp := int64(1609459200) // 2021-01-01 00:00:00 UTC

	formatted := Format(timestamp, layout)
	parsed, err := Parse(formatted, layout)
	if err != nil {
		t.Fatalf("Round trip failed: %v", err)
	}

	// Parse and format again - should get the same string
	reformatted := Format(parsed, layout)
	if formatted != reformatted {
		t.Errorf("Round trip mismatch: %q -> %d -> %q", formatted, parsed, reformatted)
	}
}
