package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/dimiro1/lunar/frontend"
	"github.com/dimiro1/lunar/internal/api"
	"github.com/dimiro1/lunar/internal/migrate"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/store"
	_ "modernc.org/sqlite"
)

// testAPIKey is the admin API key the dashboard signs in with.
const testAPIKey = "test-api-key-12345"

// browserTimeout bounds a single scenario's browser interactions. It is generous
// for local runs and leaves headroom for slower CI runners where Chrome
// cold-start plus the first Monaco load can be slow.
const browserTimeout = 90 * time.Second

// testEnv is a running, isolated Lunar server. The browser drives it as the
// dashboard, and function invocations hit its public /fn endpoint over HTTP.
type testEnv struct {
	server  *httptest.Server
	closeFn func()
}

// newTestEnv boots an in-process Lunar server backed by an in-memory SQLite
// database. Each scenario gets a fresh one for isolation.
func newTestEnv() (*testEnv, error) {
	// A ":memory:" database is private to a single connection, so pin the pool
	// to one connection — otherwise data written on one connection is invisible
	// to requests served on another.
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if err := migrate.Run(db, migrate.FS); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	server := api.NewServer(api.ServerConfig{
		DB:         store.NewSQLiteDB(db),
		Logger:     logger.NewSQLiteLogger(db),
		KVStore:    kv.NewSQLiteStore(db),
		EnvStore:   env.NewSQLiteStore(db),
		HTTPClient: internalhttp.NewDefaultClient(),
		// Wire the AI/email request trackers exactly as production does. Without
		// them the GraphQL resolvers for an execution's aiRequests/emailRequests
		// nil-panic, which surfaces when the dashboard opens an execution's
		// detail page.
		AITracker:        ai.NewSQLiteTracker(db),
		EmailTracker:     email.NewSQLiteTracker(db),
		ExecutionTimeout: 30 * time.Second,
		FrontendHandler:  frontend.Handler(),
		APIKey:           testAPIKey,
		BaseURL:          "http://localhost:8080",
	})

	ts := httptest.NewServer(server.Handler())
	return &testEnv{
		server: ts,
		closeFn: func() {
			ts.Close()
			_ = db.Close()
		},
	}, nil
}

// ── shared headless-Chrome allocator ────────────────────────────────────────
//
// Launching Chrome is expensive, so the suite starts one browser process and
// every scenario opens its own fresh tab from it. The allocator is created
// lazily on first use and torn down in AfterSuite.

var (
	allocOnce   sync.Once
	allocCtx    context.Context
	allocCancel context.CancelFunc
)

func sharedAllocator() context.Context {
	allocOnce.Do(func() {
		opts := append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			// CI runners give the container a tiny /dev/shm (~64MB); headless
			// Chrome can exhaust it and crash on startup. Writing shared memory to
			// /tmp instead keeps the browser stable on the first (cold) launch.
			chromedp.Flag("disable-dev-shm-usage", true),
		)
		// Headless by default. Set E2E_HEADED=1 (or true/yes/on) to watch the
		// suite drive a real, visible browser window — handy when debugging a
		// failing scenario. Pair it with GODOG_FEATURE to watch a single file.
		if headedMode() {
			opts = append(opts,
				chromedp.Flag("headless", false),
				chromedp.Flag("start-maximized", true),
			)
		} else {
			opts = append(opts, chromedp.Flag("headless", true))
		}
		allocCtx, allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	})
	return allocCtx
}

// headedMode reports whether the suite should run a visible browser.
func headedMode() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("E2E_HEADED"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func shutdownAllocator() {
	if allocCancel != nil {
		allocCancel()
	}
}

// warmUpBrowser launches the shared headless Chrome process and performs a
// trivial navigation before any scenario runs. The first Chrome launch on a cold
// CI runner can take well over a minute; paying that one-time cost here — outside
// any scenario's browserTimeout budget — keeps it from being charged to the
// first scenario, where it has caused flaky "context deadline exceeded" failures
// during sign-in. Errors are intentionally ignored: if warmup fails, the first
// scenario simply pays the cold start as before, so this can only help.
func warmUpBrowser() {
	tabCtx, tabCancel := chromedp.NewContext(sharedAllocator())
	defer tabCancel()
	// A generous, one-time budget — deliberately larger than browserTimeout so the
	// cold start itself never trips a deadline here.
	timedCtx, timeoutCancel := context.WithTimeout(tabCtx, 3*time.Minute)
	defer timeoutCancel()
	_ = chromedp.Run(timedCtx, chromedp.Navigate("about:blank"))
}

// world is the per-scenario state shared between steps. A fresh world is created
// for every scenario, so there is no cross-scenario leakage.
type world struct {
	env *testEnv

	browserCtx    context.Context
	browserCancel context.CancelFunc

	// Registry of function display-name -> id, captured from the dashboard URL
	// when a function is created/opened, so later steps can invoke it over HTTP.
	functions map[string]string
	lastFunc  string

	// Last function invocation result (HTTP against /fn).
	resp     *http.Response
	respBody string
}

func (w *world) setup() error {
	e, err := newTestEnv()
	if err != nil {
		return err
	}
	w.env = e
	w.functions = map[string]string{}
	w.lastFunc = ""
	w.resp = nil
	w.respBody = ""
	return nil
}

func (w *world) teardown() {
	if w.browserCancel != nil {
		w.browserCancel()
		w.browserCancel = nil
		w.browserCtx = nil
	}
	if w.env != nil {
		w.env.closeFn()
		w.env = nil
	}
}

func (w *world) baseURL() string { return w.env.server.URL }

// ── browser plumbing ────────────────────────────────────────────────────────

// browser lazily opens the scenario's Chrome tab. It auto-accepts native
// JavaScript dialogs (the dashboard uses confirm() for destructive actions like
// deleting a function or a version) so those flows don't hang the test.
func (w *world) browser() context.Context {
	if w.browserCtx != nil {
		return w.browserCtx
	}
	tabCtx, tabCancel := chromedp.NewContext(sharedAllocator())
	timedCtx, timeoutCancel := context.WithTimeout(tabCtx, browserTimeout)
	w.browserCtx = timedCtx
	w.browserCancel = func() {
		timeoutCancel()
		tabCancel()
	}

	chromedp.ListenTarget(w.browserCtx, func(ev any) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				_ = chromedp.Run(w.browserCtx, page.HandleJavaScriptDialog(true))
			}()
		}
	})
	// Enable the Page domain so dialog-opening events are delivered.
	_ = chromedp.Run(w.browserCtx, page.Enable())
	return w.browserCtx
}

func (w *world) run(actions ...chromedp.Action) error {
	return chromedp.Run(w.browser(), actions...)
}

// navigate goes to a dashboard route (a hashbang path such as "#!/functions")
// and gives the SPA a moment to render.
func (w *world) navigate(hashPath string) error {
	return w.run(
		chromedp.Navigate(w.baseURL()+hashPath),
		chromedp.Sleep(400*time.Millisecond),
	)
}

func (w *world) waitVisible(sel string) error {
	return w.run(chromedp.WaitVisible(sel, chromedp.ByQuery))
}

func (w *world) click(sel string) error {
	return w.run(chromedp.Click(sel, chromedp.ByQuery))
}

func (w *world) sendKeys(sel, text string) error {
	return w.run(chromedp.SendKeys(sel, text, chromedp.ByQuery))
}

func (w *world) currentURL() (string, error) {
	var u string
	err := w.run(chromedp.Location(&u))
	return u, err
}

func (w *world) exists(sel string) (bool, error) {
	var ok bool
	err := w.run(chromedp.Evaluate(fmt.Sprintf(`document.querySelector(%q) !== null`, sel), &ok))
	return ok, err
}

func (w *world) count(sel string) (int, error) {
	var n int
	err := w.run(chromedp.Evaluate(fmt.Sprintf(`document.querySelectorAll(%q).length`, sel), &n))
	return n, err
}

// bodyContains reports whether the visible page text contains s (case-insensitive).
func (w *world) bodyContains(s string) (bool, error) {
	var found bool
	js := fmt.Sprintf(`document.body && document.body.innerText.toLowerCase().includes(%q)`, strings.ToLower(s))
	err := w.run(chromedp.Evaluate(js, &found))
	return found, err
}

// selectorContains reports whether the element matching sel contains want in its
// visible text (case-insensitive).
func (w *world) selectorContains(sel, want string) (bool, error) {
	var found bool
	js := fmt.Sprintf(`(() => { const e = document.querySelector(%q); return !!e && e.innerText.toLowerCase().includes(%q); })()`,
		sel, strings.ToLower(want))
	err := w.run(chromedp.Evaluate(js, &found))
	return found, err
}

// ── dashboard sign-in ───────────────────────────────────────────────────────

func (w *world) signIn(key string) error {
	return w.run(
		chromedp.Navigate(w.baseURL()+"#!/login"),
		chromedp.WaitVisible(`#api-key`, chromedp.ByQuery),
		chromedp.SendKeys(`#api-key`, key, chromedp.ByQuery),
		chromedp.Sleep(150*time.Millisecond),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
	)
}

// ── function id registry ────────────────────────────────────────────────────

var functionIDRe = regexp.MustCompile(`/functions/([^/?#]+)`)

// rememberCurrentFunction reads the function id out of the current dashboard URL
// (a function detail route is #!/functions/<id>[/tab]) and records it under name.
func (w *world) rememberCurrentFunction(name string) error {
	u, err := w.currentURL()
	if err != nil {
		return err
	}
	m := functionIDRe.FindStringSubmatch(u)
	if m == nil || m[1] == "new" {
		return fmt.Errorf("could not read a function id from URL %q", u)
	}
	w.functions[name] = m[1]
	w.lastFunc = name
	return nil
}

func (w *world) functionID(name string) (string, error) {
	if id, ok := w.functions[name]; ok {
		return id, nil
	}
	return "", fmt.Errorf("unknown function %q (was it created/opened in a prior step?)", name)
}

// ── HTTP function invocation (the public /fn passthrough) ────────────────────

func (w *world) invoke(method, name, suffix, body string, headers map[string]string) error {
	id, err := w.functionID(name)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, w.baseURL()+"/fn/"+id+suffix, strings.NewReader(body))
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	w.resp = resp
	w.respBody = string(b)
	return nil
}
