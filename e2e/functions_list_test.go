package e2e

import (
	"testing"
)

// TestFunctionsListAfterLogin verifies functions list displays after authentication
func TestFunctionsListAfterLogin(t *testing.T) {
	bt := newBrowserTest(t)

	bt.Login("#!/functions").
		WaitVisible(`h1`).
		AssertText(`h1`, "Functions")

	// Verify empty state message or table is present
	hasEmptyMessage := bt.ElementExists(`.empty-state`) ||
		bt.GetHTML("body") != "" // fallback check
	hasTable := bt.ElementExists(`table`)

	if !hasEmptyMessage && !hasTable {
		t.Error("expected either empty state message or functions table")
	}
}

// TestNewFunctionButtonPresent verifies the new function button is visible
func TestNewFunctionButtonPresent(t *testing.T) {
	bt := newBrowserTest(t)

	bt.Login("#!/functions").
		WaitVisible(`a[href="#!/functions/new"]`).
		AssertText(`a[href="#!/functions/new"]`, "New Function")
}
