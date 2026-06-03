package e2e

import (
	"context"
	"strings"
	"testing"
	"time"
)

// assertLangLabel checks the editor's language label case-insensitively (it is
// uppercased by CSS, so the rendered text is e.g. "LUA").
func assertLangLabel(t *testing.T, bt *browserTest, want string) {
	t.Helper()
	got := bt.GetText(`.code-editor-lang`)
	if !strings.EqualFold(strings.TrimSpace(got), want) {
		t.Errorf("editor language label = %q, want %q (any case)", got, want)
	}
}

// TestCodePage_Lua shows the Lua editor language and the Lua API reference.
func TestCodePage_Lua(t *testing.T) {
	bt := newBrowserTest(t)
	seedFunction(t, bt.env, "page_lua", "lua",
		"function handler(ctx, event) return { statusCode = 200 } end")

	bt.Login("#!/functions/page_lua").
		WaitVisible(`.function-details-title`).
		AssertElementExists(`.code-editor-container`).
		// The handler section (default) types event fields as "table" for Lua.
		AssertText(`.api-reference`, "table")
	assertLangLabel(t, bt, "lua")
}

// TestCodePage_Starlark shows the Starlark editor language and the Starlark API
// reference (event fields typed as "dict", not "table").
func TestCodePage_Starlark(t *testing.T) {
	bt := newBrowserTest(t)
	seedFunction(t, bt.env, "page_star", "starlark",
		"def handler(ctx, event):\n    return {\"statusCode\": 200}")

	bt.Login("#!/functions/page_star").
		WaitVisible(`.function-details-title`).
		AssertElementExists(`.code-editor-container`).
		AssertText(`.api-reference`, "dict")
	assertLangLabel(t, bt, "starlark")
}

// TestVersionsPage_ListsAllVersions verifies every version of a function is
// listed on the versions page.
func TestVersionsPage_ListsAllVersions(t *testing.T) {
	bt := newBrowserTest(t)
	ctx := context.Background()
	seedFunction(t, bt.env, "page_versions", "lua",
		"function handler(ctx, event) return { statusCode = 200, body = 'v1' } end")
	if _, err := bt.env.Store.CreateVersion(ctx, "page_versions",
		"function handler(ctx, event) return { statusCode = 200, body = 'v2' } end",
		"", nil); err != nil {
		t.Fatalf("CreateVersion v2: %v", err)
	}

	bt.Login("#!/functions/page_versions/versions").
		WaitVisible(`tbody tr`).
		AssertElementCount(`tbody tr`, 2)
}

// TestSettingsPage_Renders verifies the settings page renders with the function
// name and the general settings controls.
func TestSettingsPage_Renders(t *testing.T) {
	bt := newBrowserTest(t)
	seedFunction(t, bt.env, "page_settings", "lua",
		"function handler(ctx, event) return { statusCode = 200 } end")

	bt.Login("#!/functions/page_settings/settings").
		WaitVisible(`.function-details-title`).
		AssertText(`.function-details-title`, "page_settings").
		AssertElementExists(`#save-response`).
		AssertElementExists(`#logRetention`)
}

// TestExecutionsPage_ShowsInvocations verifies that invoking a function records
// an execution that then appears on the executions page.
func TestExecutionsPage_ShowsInvocations(t *testing.T) {
	bt := newBrowserTest(t)
	seedFunction(t, bt.env, "page_exec", "lua",
		"function handler(ctx, event) return { statusCode = 200, body = 'ok' } end")

	// Two invocations -> two execution records.
	invoke(t, bt.env, "GET", "/fn/page_exec", "", nil)
	invoke(t, bt.env, "GET", "/fn/page_exec", "", nil)

	bt.Login("#!/functions/page_exec/executions").
		WaitVisible(`tbody tr`).
		AssertElementCount(`tbody tr`, 2)
}

// TestExecutionsPage_Empty verifies the executions page renders for a function
// that has never been invoked (empty state, no crash).
func TestExecutionsPage_Empty(t *testing.T) {
	bt := newBrowserTest(t)
	seedFunction(t, bt.env, "page_exec_empty", "lua",
		"function handler(ctx, event) return { statusCode = 200 } end")

	bt.Login("#!/functions/page_exec_empty/executions").
		WaitVisible(`.function-details-title`).
		Sleep(300 * time.Millisecond).
		AssertElementCount(`tbody tr`, 0)
}
