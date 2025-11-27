package e2e

import (
	"testing"
	"time"
)

// TestCreateFunctionWithHTTPTemplate verifies creating a function with default HTTP template
func TestCreateFunctionWithHTTPTemplate(t *testing.T) {
	bt := newBrowserTest(t)
	functionName := "http-func-" + time.Now().Format("150405")

	bt.Login("#!/functions/new").
		WaitVisible(`#function-name`).
		Type(`#function-name`, functionName).
		Click(`.create-function-actions button`).
		Sleep(1 * time.Second).
		AssertURL("#!/functions").
		AssertURLNot("/new").
		AssertFunctionCode(functionName,
			"-- HTTP Handler",
			"function handler(ctx, event)",
			`message = "Hello from Lua!"`,
			"event.method",
		)
}

// TestCreateFunctionWithAPITemplate verifies creating a function with API template
func TestCreateFunctionWithAPITemplate(t *testing.T) {
	bt := newBrowserTest(t)
	functionName := "api-func-" + time.Now().Format("150405")

	bt.Login("#!/functions/new").
		WaitVisible(`#function-name`).
		Type(`#function-name`, functionName).
		Click(`.template-card:nth-child(2)`). // Select API template
		Sleep(200 * time.Millisecond).
		Click(`.create-function-actions button`).
		Sleep(1 * time.Second).
		AssertURL("#!/functions").
		AssertURLNot("/new").
		AssertFunctionCode(functionName,
			"-- REST API Endpoint",
			`method == "GET"`,
			`method == "POST"`,
			"crypto.uuid()",
		).
		AssertFunctionCodeNot(functionName, "-- HTTP Handler")
}

// TestCreateFunctionValidation verifies validation error for empty name
func TestCreateFunctionValidation(t *testing.T) {
	bt := newBrowserTest(t)

	bt.Login("#!/functions/new").
		WaitVisible(`#function-name`).
		Click(`.create-function-actions button`). // Submit without name
		Sleep(500 * time.Millisecond).
		AssertURL("/new") // Should stay on create page

	// Should show error state
	hasError := bt.ElementExists(`.form-input--error`) ||
		bt.ElementExists(`.form-help--error`) ||
		bt.ElementExists(`.toast--error`)

	if !hasError {
		t.Error("expected validation error for empty name")
	}
}

// TestCreateFunctionNavigation verifies navigation to and from the create page
func TestCreateFunctionNavigation(t *testing.T) {
	bt := newBrowserTest(t)

	// Navigate from list to create page
	bt.Login("#!/functions").
		WaitVisible(`a[href="#!/functions/new"]`).
		Click(`a[href="#!/functions/new"]`).
		Sleep(500 * time.Millisecond).
		AssertURL("/new").
		WaitVisible(`h1.create-function-title`)

	// Cancel button returns to list
	bt.Click(`a[href="#!/functions"]`).
		Sleep(500 * time.Millisecond).
		AssertURL("#!/functions").
		AssertURLNot("/new")

	// Navigate back to create, use back button
	bt.Click(`a[href="#!/functions/new"]`).
		Sleep(500 * time.Millisecond).
		WaitVisible(`.create-function-back a`).
		Click(`.create-function-back a`).
		Sleep(500 * time.Millisecond).
		AssertURL("#!/functions").
		AssertURLNot("/new")
}
