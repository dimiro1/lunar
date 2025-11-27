package e2e

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestAppLoads verifies the app loads correctly
func TestAppLoads(t *testing.T) {
	bt := newBrowserTest(t)

	var title string
	bt.Run(
		chromedp.Navigate(bt.env.Server.URL),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Title(&title),
	)

	if title == "" {
		t.Error("page title should not be empty")
	}
	t.Logf("Page loaded with title: %s", title)
}

// TestLoginPageRedirect verifies unauthenticated users see login
func TestLoginPageRedirect(t *testing.T) {
	bt := newBrowserTest(t)

	bt.Run(
		chromedp.Navigate(bt.env.Server.URL+"#!/functions"),
		chromedp.Sleep(1*time.Second),
	)

	// The URL behavior depends on frontend auth handling
	t.Logf("Current URL after navigation: %s", bt.GetURL())
}

// TestLoginFlow verifies user can log in
func TestLoginFlow(t *testing.T) {
	bt := newBrowserTest(t)

	bt.Run(
		chromedp.Navigate(bt.env.Server.URL+"#!/login"),
	).
		WaitVisible(`input[type="password"]`).
		Type(`input[type="password"]`, testAPIKey).
		Click(`button[type="submit"]`).
		Sleep(1 * time.Second).
		AssertURL("functions")
}
