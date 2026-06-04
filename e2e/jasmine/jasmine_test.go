// Package jasmine runs the frontend's Jasmine unit-test suite end-to-end in a
// real headless Chrome browser and asserts that every spec passes.
//
// The Jasmine specs (frontend/test/spec/**) exercise the dashboard's
// client-side code — components, routing, i18n, utilities — directly in the
// browser, the way they actually run. `mise run test-frontend` (cmd/testserver)
// serves the very same SpecRunner.html for a human to eyeball; this test
// automates the verdict so a frontend regression fails CI instead of waiting for
// someone to look.
//
// It lives under e2e/ so it is gated behind the e2e task (which already requires
// a Chrome/Chromium on the machine) and excluded from the unit-test task, which
// runs `go list ./... | grep -v /e2e`.
//
// Run it on its own with:
//
//	go test ./e2e/jasmine/...
//
// It is also covered by `mise run test-e2e`, which runs `go test ./e2e/...`.
package jasmine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// frontendDir resolves the absolute path to the repository's frontend/ directory
// from this test file's location, so the test works regardless of the working
// directory `go test` happens to be invoked from.
func frontendDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	// e2e/jasmine/jasmine_test.go -> e2e -> repo root -> frontend
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "frontend")
}

// jasmineResults mirrors the window.__jasmineResults object that SpecRunner.html
// publishes once the whole suite has finished.
type jasmineResults struct {
	OverallStatus string `json:"overallStatus"`
	Total         int    `json:"total"`
	Failures      []struct {
		FullName string   `json:"fullName"`
		Messages []string `json:"messages"`
	} `json:"failures"`
}

func TestJasmineSuitePasses(t *testing.T) {
	// Serve the on-disk frontend directory exactly as cmd/testserver does. The
	// SpecRunner and its specs are not embedded in the production binary, so they
	// must be served straight from the source tree.
	ts := httptest.NewServer(http.FileServer(http.Dir(frontendDir())))
	defer ts.Close()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("headless", true),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	// Generous bound: Chrome cold-start plus loading and running the full spec
	// suite (it imports the real app modules) can be slow on CI runners.
	ctx, cancel := context.WithTimeout(browserCtx, 90*time.Second)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Navigate(ts.URL+"/test/SpecRunner.html")); err != nil {
		t.Fatalf("navigate to SpecRunner.html: %v", err)
	}

	// The specs are ES modules executed asynchronously, and Jasmine itself only
	// starts on window.onload, so the verdict lags the page load — poll for it.
	if err := waitJasmineDone(ctx); err != nil {
		t.Fatalf("waiting for the Jasmine suite to finish: %v", err)
	}

	var res jasmineResults
	if err := chromedp.Run(ctx, chromedp.Evaluate(`window.__jasmineResults`, &res)); err != nil {
		t.Fatalf("reading Jasmine results: %v", err)
	}

	// Zero specs means the spec modules failed to load (a bad import path or a
	// syntax error) rather than a genuinely empty suite — treat it as a failure.
	if res.Total == 0 {
		t.Fatal("Jasmine reported 0 specs — the spec files likely failed to load")
	}

	if res.OverallStatus != "passed" {
		var b strings.Builder
		fmt.Fprintf(&b, "Jasmine suite did not pass: overall status = %q, %d/%d specs failed\n",
			res.OverallStatus, len(res.Failures), res.Total)
		for _, f := range res.Failures {
			fmt.Fprintf(&b, "  ✗ %s\n", f.FullName)
			for _, m := range f.Messages {
				fmt.Fprintf(&b, "      %s\n", strings.ReplaceAll(m, "\n", "\n      "))
			}
		}
		t.Fatal(b.String())
	}

	t.Logf("Jasmine suite passed: %d specs, 0 failures", res.Total)
}

// waitJasmineDone polls until SpecRunner.html has published its final results
// (up to ~30s once the page is loaded).
func waitJasmineDone(ctx context.Context) error {
	for range 120 {
		var done bool
		if err := chromedp.Run(ctx, chromedp.Evaluate(`window.__jasmineResults !== null`, &done)); err != nil {
			return err
		}
		if done {
			return nil
		}
		if err := chromedp.Run(ctx, chromedp.Sleep(250*time.Millisecond)); err != nil {
			return err
		}
	}
	return context.DeadlineExceeded
}
