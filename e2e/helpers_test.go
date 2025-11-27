package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dimiro1/faas-go/frontend"
	"github.com/dimiro1/faas-go/internal/api"
	"github.com/dimiro1/faas-go/internal/env"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
	"github.com/dimiro1/faas-go/internal/migrate"
	"github.com/dimiro1/faas-go/internal/store"
	_ "modernc.org/sqlite"
)

const testAPIKey = "test-api-key-12345"

// testEnv holds the test server and database for e2e tests
type testEnv struct {
	Server *httptest.Server
	Store  *store.SQLiteDB
}

// browserTest provides a fluent API for writing e2e tests
type browserTest struct {
	t          *testing.T
	env        *testEnv
	browserCtx context.Context
	cancel     context.CancelFunc
}

// newBrowserTest creates a new browser test with server and browser context
func newBrowserTest(t *testing.T) *browserTest {
	t.Helper()
	env := startTestServer(t)
	ctx, cancel := newBrowserContext(t, 30*time.Second)

	t.Cleanup(func() {
		cancel()
	})

	return &browserTest{
		t:          t,
		env:        env,
		browserCtx: ctx,
		cancel:     cancel,
	}
}

// Run executes chromedp actions
func (bt *browserTest) Run(actions ...chromedp.Action) *browserTest {
	bt.t.Helper()
	if err := chromedp.Run(bt.browserCtx, actions...); err != nil {
		bt.t.Fatalf("chromedp run failed: %v", err)
	}
	return bt
}

// Login logs in and navigates to the specified path
func (bt *browserTest) Login(path string) *browserTest {
	bt.t.Helper()
	return bt.Run(loginAndNavigate(bt.env.Server, path)...)
}

// NavigateTo navigates to a path (assumes already logged in)
func (bt *browserTest) NavigateTo(path string) *browserTest {
	bt.t.Helper()
	return bt.Run(
		chromedp.Navigate(bt.env.Server.URL+path),
		chromedp.Sleep(500*time.Millisecond),
	)
}

// WaitVisible waits for an element to be visible
func (bt *browserTest) WaitVisible(selector string) *browserTest {
	bt.t.Helper()
	return bt.Run(chromedp.WaitVisible(selector, chromedp.ByQuery))
}

// Click clicks an element
func (bt *browserTest) Click(selector string) *browserTest {
	bt.t.Helper()
	return bt.Run(chromedp.Click(selector, chromedp.ByQuery))
}

// Type types text into an input field
func (bt *browserTest) Type(selector, text string) *browserTest {
	bt.t.Helper()
	return bt.Run(chromedp.SendKeys(selector, text, chromedp.ByQuery))
}

// Sleep waits for a duration
func (bt *browserTest) Sleep(d time.Duration) *browserTest {
	bt.t.Helper()
	return bt.Run(chromedp.Sleep(d))
}

// GetURL returns the current URL
func (bt *browserTest) GetURL() string {
	bt.t.Helper()
	var url string
	bt.Run(chromedp.Location(&url))
	return url
}

// GetText returns the text content of an element
func (bt *browserTest) GetText(selector string) string {
	bt.t.Helper()
	var text string
	bt.Run(chromedp.Text(selector, &text, chromedp.ByQuery))
	return text
}

// GetHTML returns the outer HTML of an element
func (bt *browserTest) GetHTML(selector string) string {
	bt.t.Helper()
	var html string
	bt.Run(chromedp.OuterHTML(selector, &html, chromedp.ByQuery))
	return html
}

// ElementExists checks if an element exists
func (bt *browserTest) ElementExists(selector string) bool {
	bt.t.Helper()
	var exists bool
	bt.Run(chromedp.Evaluate(
		fmt.Sprintf(`document.querySelector('%s') !== null`, selector),
		&exists,
	))
	return exists
}

// ElementCount returns the number of elements matching a selector
func (bt *browserTest) ElementCount(selector string) int {
	bt.t.Helper()
	var count int
	bt.Run(chromedp.Evaluate(
		fmt.Sprintf(`document.querySelectorAll('%s').length`, selector),
		&count,
	))
	return count
}

// AssertURL asserts the current URL contains the expected substring
func (bt *browserTest) AssertURL(contains string) *browserTest {
	bt.t.Helper()
	url := bt.GetURL()
	if !strings.Contains(url, contains) {
		bt.t.Errorf("expected URL to contain %q, got: %s", contains, url)
	}
	return bt
}

// AssertURLNot asserts the current URL does not contain a substring
func (bt *browserTest) AssertURLNot(notContains string) *browserTest {
	bt.t.Helper()
	url := bt.GetURL()
	if strings.Contains(url, notContains) {
		bt.t.Errorf("expected URL to NOT contain %q, got: %s", notContains, url)
	}
	return bt
}

// AssertText asserts an element's text contains the expected substring
func (bt *browserTest) AssertText(selector, contains string) *browserTest {
	bt.t.Helper()
	text := bt.GetText(selector)
	if !strings.Contains(text, contains) {
		bt.t.Errorf("expected %q text to contain %q, got: %s", selector, contains, text)
	}
	return bt
}

// AssertElementExists asserts an element exists
func (bt *browserTest) AssertElementExists(selector string) *browserTest {
	bt.t.Helper()
	if !bt.ElementExists(selector) {
		bt.t.Errorf("expected element %q to exist", selector)
	}
	return bt
}

// AssertElementCount asserts the number of elements
func (bt *browserTest) AssertElementCount(selector string, expected int) *browserTest {
	bt.t.Helper()
	count := bt.ElementCount(selector)
	if count != expected {
		bt.t.Errorf("expected %d elements for %q, got: %d", expected, selector, count)
	}
	return bt
}

// GetFunction retrieves a function from the store by name
func (bt *browserTest) GetFunction(name string) *store.FunctionWithActiveVersion {
	bt.t.Helper()
	functions, _, err := bt.env.Store.ListFunctions(context.Background(), store.PaginationParams{Limit: 100})
	if err != nil {
		bt.t.Fatalf("failed to list functions: %v", err)
	}
	for _, fn := range functions {
		if fn.Name == name {
			return &fn
		}
	}
	return nil
}

// AssertFunctionExists asserts a function exists in the database
func (bt *browserTest) AssertFunctionExists(name string) *store.FunctionWithActiveVersion {
	bt.t.Helper()
	fn := bt.GetFunction(name)
	if fn == nil {
		bt.t.Errorf("function %q not found in database", name)
	}
	return fn
}

// AssertFunctionNotExists asserts a function does not exist in the database
func (bt *browserTest) AssertFunctionNotExists(name string) *browserTest {
	bt.t.Helper()
	fn := bt.GetFunction(name)
	if fn != nil {
		bt.t.Errorf("function %q should not exist in database", name)
	}
	return bt
}

// AssertFunctionCode asserts the function's code contains expected strings
func (bt *browserTest) AssertFunctionCode(name string, contains ...string) *browserTest {
	bt.t.Helper()
	fn := bt.AssertFunctionExists(name)
	if fn == nil {
		return bt
	}
	code := fn.ActiveVersion.Code
	for _, s := range contains {
		if !strings.Contains(code, s) {
			bt.t.Errorf("function %q code should contain %q", name, s)
		}
	}
	return bt
}

// AssertFunctionCodeNot asserts the function's code does NOT contain strings
func (bt *browserTest) AssertFunctionCodeNot(name string, notContains ...string) *browserTest {
	bt.t.Helper()
	fn := bt.GetFunction(name)
	if fn == nil {
		return bt
	}
	code := fn.ActiveVersion.Code
	for _, s := range notContains {
		if strings.Contains(code, s) {
			bt.t.Errorf("function %q code should NOT contain %q", name, s)
		}
	}
	return bt
}

// startTestServer creates a test server with an in-memory database
func startTestServer(t *testing.T) *testEnv {
	t.Helper()

	// Create in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run migrations
	if err := migrate.Run(db, migrate.FS); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create dependencies
	apiDB := store.NewSQLiteDB(db)
	kvStore := kv.NewSQLiteStore(db)
	envStore := env.NewSQLiteStore(db)
	appLogger := logger.NewSQLiteLogger(db)
	httpClient := internalhttp.NewDefaultClient()

	// Create server
	server := api.NewServer(api.ServerConfig{
		DB:               apiDB,
		Logger:           appLogger,
		KVStore:          kvStore,
		EnvStore:         envStore,
		HTTPClient:       httpClient,
		ExecutionTimeout: 30 * time.Second,
		FrontendHandler:  frontend.Handler(),
		APIKey:           testAPIKey,
		BaseURL:          "http://localhost:8080",
	})

	// Create test server
	ts := httptest.NewServer(server.Handler())

	// Register cleanup
	t.Cleanup(func() {
		ts.Close()
		_ = db.Close()
	})

	return &testEnv{
		Server: ts,
		Store:  apiDB,
	}
}

// newBrowserContext creates a chromedp context with timeout
func newBrowserContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()

	// Create allocator context (uses headless Chrome by default)
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)

	// Create browser context
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	// Add timeout
	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)

	// Combined cleanup
	cancel := func() {
		timeoutCancel()
		ctxCancel()
		allocCancel()
	}

	return ctx, cancel
}

// loginAndNavigate is a helper that logs in and navigates to a given path
func loginAndNavigate(srv *httptest.Server, path string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(srv.URL + "#!/login"),
		chromedp.WaitVisible(`input[type="password"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[type="password"]`, testAPIKey, chromedp.ByQuery),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(1 * time.Second),
		chromedp.Navigate(srv.URL + path),
		chromedp.Sleep(500 * time.Millisecond),
	}
}
