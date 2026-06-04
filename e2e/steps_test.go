package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/cucumber/godog"
)

// templateIndex maps a starter-template name to its position among the
// .template-card elements on the create page. The order is fixed by
// FunctionTemplates in the frontend: http, api, aiChat, email, router, blank.
var templateIndex = map[string]int{
	"HTTP": 1, "http": 1,
	"REST API": 2, "API": 2, "api": 2,
	"AI Chat": 3, "aiChat": 3,
	"Email": 4, "email": 4,
	"Router": 5, "router": 5,
	"Blank": 6, "blank": 6,
}

// registerSteps binds every Gherkin step to a browser/HTTP flow. The phrasing is
// deliberately product-level; all selector and timing detail lives here, not in
// the feature files.
func (w *world) registerSteps(sc *godog.ScenarioContext) {
	// Sign in / out
	sc.Step(`^I am signed in$`, w.givenSignedIn)
	sc.Step(`^I sign in with a valid API key$`, w.signInValid)
	sc.Step(`^I sign in with an invalid API key$`, w.signInInvalid)
	sc.Step(`^I sign out$`, w.doSignOut)
	sc.Step(`^I should reach the dashboard$`, w.shouldReachDashboard)
	sc.Step(`^I should be told the key is invalid$`, w.shouldSeeSignInError)
	sc.Step(`^I should be signed out$`, w.shouldBeSignedOut)

	// Functions list
	sc.Step(`^I have not created any functions yet$`, noop)
	sc.Step(`^I view my functions$`, w.viewFunctions)
	sc.Step(`^my functions list should be empty$`, w.listShouldBeEmpty)
	sc.Step(`^I should see "([^"]*)" in my functions list$`, w.shouldSeeInList)
	sc.Step(`^I should no longer see "([^"]*)" in my functions list$`, w.shouldNotSeeInList)
	sc.Step(`^"([^"]*)" should be shown as a supported language$`, w.languageShown)

	// Create
	sc.Step(`^I create a function named "([^"]*)"$`, w.createLua)
	sc.Step(`^I create a Starlark function named "([^"]*)"$`, w.createStarlark)
	sc.Step(`^I create a function named "([^"]*)" from the "([^"]*)" template$`, w.createFromTemplate)
	sc.Step(`^I try to create a function without giving it a name$`, w.createWithoutName)
	sc.Step(`^I should land back on my functions list$`, w.shouldLandOnList)
	sc.Step(`^I should be told the function needs a name$`, w.shouldSeeValidationError)

	// Open a function
	sc.Step(`^I open the function "([^"]*)"$`, w.openFunctionStep)

	// Code & versions
	sc.Step(`^I change the function's code to:$`, w.changeCode)
	sc.Step(`^I view the function's version history$`, w.viewVersions)
	sc.Step(`^the function should have (\d+) versions?$`, w.shouldHaveVersions)
	sc.Step(`^version (\d+) should be the active version$`, w.versionShouldBeActive)
	sc.Step(`^I roll back to version (\d+)$`, w.rollBackTo)
	sc.Step(`^I compare version (\d+) with version (\d+)$`, w.compareVersionsStep)
	sc.Step(`^I should see the differences between the two versions$`, w.shouldSeeDiff)

	// Invocation (the public HTTP endpoint a real client would call)
	sc.Step(`^I call the function$`, w.callFunction)
	sc.Step(`^I call the function with the (GET|POST|PUT|DELETE|PATCH) method$`, w.callWithMethod)
	sc.Step(`^I call the function at "([^"]*)"$`, w.callAtPath)
	sc.Step(`^the function should respond successfully$`, w.respondedOK)
	sc.Step(`^the response should contain "([^"]*)"$`, w.responseContains)
	sc.Step(`^the response should be "([^"]*)"$`, w.responseEquals)
	sc.Step(`^the call should be refused because the function is disabled$`, w.callForbidden)
	sc.Step(`^the function should no longer be reachable$`, w.callNotFound)

	// Settings
	sc.Step(`^I disable the function$`, w.disableFunction)
	sc.Step(`^I rename the function to "([^"]*)"$`, w.renameFunction)
	sc.Step(`^I delete the function$`, w.deleteFunctionStep)
	sc.Step(`^I give the function an environment variable "([^"]*)" set to "([^"]*)"$`, w.setEnvVar)
	sc.Step(`^I schedule the function to run every 5 minutes$`, w.scheduleEvery5Min)
	sc.Step(`^I should see when the function will next run$`, w.shouldSeeNextRun)

	// KV store
	sc.Step(`^I store the value "([^"]*)" under the key "([^"]*)"$`, w.storeScopedValue)
	sc.Step(`^the stored value "([^"]*)" should be kept under the key "([^"]*)"$`, w.storedValueKept)
	sc.Step(`^I store the value "([^"]*)" under the global key "([^"]*)"$`, w.storeGlobalValue)
	sc.Step(`^the global value "([^"]*)" should be kept under the key "([^"]*)"$`, w.storedGlobalValueKept)

	// Test page
	sc.Step(`^I try the function from the Test page$`, w.tryFromTestPage)
	sc.Step(`^the Test page should show a successful response$`, w.testPageSuccessful)
	sc.Step(`^the Test page response should contain "([^"]*)"$`, w.testPageResponseContains)

	// Executions
	sc.Step(`^I view the function's execution history$`, w.viewExecutions)
	sc.Step(`^the execution history should be empty$`, w.executionsEmpty)
	sc.Step(`^the function should have (\d+) recorded executions?$`, w.shouldHaveExecutions)
	sc.Step(`^the latest execution should be successful$`, w.latestExecutionSuccessful)
	sc.Step(`^I open the most recent execution$`, w.openLatestExecution)
	sc.Step(`^the execution should be shown as successful$`, w.executionDetailSuccessful)

	// Connected clients
	sc.Step(`^I open the connected clients page$`, w.openClients)
	sc.Step(`^the connected clients page should be shown$`, w.clientsShown)
}

func noop() error { return nil }

// ── sign in / out ───────────────────────────────────────────────────────────

func (w *world) givenSignedIn() error {
	if err := w.signIn(testAPIKey); err != nil {
		return err
	}
	return w.waitVisible(`.navbar`)
}

func (w *world) signInValid() error   { return w.signIn(testAPIKey) }
func (w *world) signInInvalid() error { return w.signIn("not-the-real-key") }

func (w *world) doSignOut() error {
	if err := w.click(`button.navbar__action`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(700 * time.Millisecond))
}

func (w *world) shouldReachDashboard() error { return w.urlContains("functions") }
func (w *world) shouldBeSignedOut() error    { return w.urlContains("login") }

func (w *world) shouldSeeSignInError() error {
	if err := w.urlContains("login"); err != nil {
		return err
	}
	ok, err := w.exists(`.form-help--error`)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected a sign-in error to be shown")
	}
	return nil
}

// ── functions list ──────────────────────────────────────────────────────────

func (w *world) viewFunctions() error {
	if err := w.navigate("#!/functions"); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(400 * time.Millisecond))
}

func (w *world) listShouldBeEmpty() error {
	if err := w.viewFunctions(); err != nil {
		return err
	}
	ok, err := w.exists(`.table__empty`)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected the empty-state message on the functions list")
	}
	return nil
}

func (w *world) shouldSeeInList(name string) error    { return w.assertInList(name, true) }
func (w *world) shouldNotSeeInList(name string) error { return w.assertInList(name, false) }

func (w *world) assertInList(name string, want bool) error {
	if err := w.viewFunctions(); err != nil {
		return err
	}
	var present bool
	js := fmt.Sprintf(`[...document.querySelectorAll('tbody td')].some(td => td.textContent.trim() === %q)`, name)
	if err := w.run(chromedp.Evaluate(js, &present)); err != nil {
		return err
	}
	if present != want {
		return fmt.Errorf("function %q presence in list = %v, want %v", name, present, want)
	}
	return nil
}

func (w *world) languageShown(lang string) error {
	if err := w.viewFunctions(); err != nil {
		return err
	}
	if err := w.waitVisible(`tbody tr`); err != nil {
		return err
	}
	found, err := w.bodyContains(lang)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("expected %q to be shown as a language in the list", lang)
	}
	return nil
}

// ── create ──────────────────────────────────────────────────────────────────

func (w *world) createLua(name string) error      { return w.createFunction(name, "lua", "http") }
func (w *world) createStarlark(name string) error { return w.createFunction(name, "starlark", "http") }

func (w *world) createFromTemplate(name, template string) error {
	return w.createFunction(name, "lua", template)
}

func (w *world) createFunction(name, language, template string) error {
	if err := w.navigate("#!/functions/new"); err != nil {
		return err
	}
	if err := w.waitVisible(`#function-name`); err != nil {
		return err
	}
	if err := w.sendKeys(`#function-name`, name); err != nil {
		return err
	}
	if language != "" && language != "lua" {
		if err := w.selectCreateLanguage(language); err != nil {
			return err
		}
	}
	if template != "" && template != "http" {
		idx, ok := templateIndex[template]
		if !ok {
			return fmt.Errorf("unknown template %q", template)
		}
		if err := w.run(
			chromedp.Click(fmt.Sprintf(`.template-card:nth-child(%d)`, idx), chromedp.ByQuery),
			chromedp.Sleep(200*time.Millisecond),
		); err != nil {
			return err
		}
	}
	if err := w.click(`.create-function-actions button`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(1200 * time.Millisecond))
}

func (w *world) createWithoutName() error {
	if err := w.navigate("#!/functions/new"); err != nil {
		return err
	}
	if err := w.waitVisible(`#function-name`); err != nil {
		return err
	}
	if err := w.click(`.create-function-actions button`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(500 * time.Millisecond))
}

func (w *world) selectCreateLanguage(value string) error {
	var ok bool
	js := fmt.Sprintf(
		`(() => { const el = document.querySelector("#function-language"); if (!el) return false; el.value = %q; el.dispatchEvent(new Event("change", { bubbles: true })); return true; })()`,
		value,
	)
	return w.run(chromedp.Evaluate(js, &ok), chromedp.Sleep(200*time.Millisecond))
}

func (w *world) shouldLandOnList() error {
	if err := w.urlContains("#!/functions"); err != nil {
		return err
	}
	return w.urlNotContains("/new")
}

func (w *world) shouldSeeValidationError() error {
	if err := w.urlContains("/new"); err != nil {
		return err
	}
	for _, sel := range []string{`.form-input--error`, `.form-help--error`, `.toast--error`} {
		ok, err := w.exists(sel)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return fmt.Errorf("expected a validation error after submitting without a name")
}

// ── open a function ─────────────────────────────────────────────────────────

func (w *world) openFunctionStep(name string) error {
	if err := w.navigate("#!/functions"); err != nil {
		return err
	}
	if err := w.waitVisible(`tbody tr`); err != nil {
		return err
	}
	var ok bool
	js := fmt.Sprintf(`(() => {
		const rows = [...document.querySelectorAll('tbody tr')];
		for (const r of rows) {
			const cell = r.querySelector('td');
			if (cell && cell.textContent.trim() === %q) { r.click(); return true; }
		}
		return false;
	})()`, name)
	if err := w.run(chromedp.Evaluate(js, &ok)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("function %q not found in the list", name)
	}
	if err := w.run(chromedp.Sleep(800 * time.Millisecond)); err != nil {
		return err
	}
	return w.rememberCurrentFunction(name)
}

// ── code & versions ─────────────────────────────────────────────────────────

func (w *world) changeCode(code *godog.DocString) error {
	id, err := w.functionID(w.lastFunc)
	if err != nil {
		return err
	}
	if err := w.navigate("#!/functions/" + id); err != nil {
		return err
	}
	if err := w.waitVisible(`.code-editor-lang`); err != nil {
		return err
	}
	if err := w.setEditorContent(code.Content); err != nil {
		return err
	}
	if err := w.run(chromedp.Sleep(500 * time.Millisecond)); err != nil {
		return err
	}
	// The save button enables once the editor reports a change.
	saveBtn := `.function-details-actions button.btn--primary`
	if err := w.waitEnabled(saveBtn); err != nil {
		return err
	}
	if err := w.click(saveBtn); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(1200 * time.Millisecond))
}

// setEditorContent drives the Monaco model directly — the editor has no backing
// textarea, and replacing the whole document by keystrokes is fragile. Setting
// the model value fires the same change event a human's typing would.
func (w *world) setEditorContent(code string) error {
	if err := w.waitJSTrue(`!!(window.monaco && monaco.editor.getModels().length > 0)`); err != nil {
		return fmt.Errorf("monaco editor did not become ready: %w", err)
	}
	lit, _ := json.Marshal(code)
	var ok bool
	js := fmt.Sprintf(`(() => { const m = monaco.editor.getModels(); if (!m.length) return false; m[0].setValue(%s); return m[0].getValue() === %s; })()`, lit, lit)
	if err := w.run(chromedp.Evaluate(js, &ok)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("failed to set the editor content")
	}
	return nil
}

func (w *world) viewVersions() error {
	id, err := w.functionID(w.lastFunc)
	if err != nil {
		return err
	}
	if err := w.navigate("#!/functions/" + id + "/versions"); err != nil {
		return err
	}
	return w.waitVisible(`tbody tr`)
}

func (w *world) shouldHaveVersions(n int) error {
	if err := w.viewVersions(); err != nil {
		return err
	}
	got, err := w.count(`tbody tr`)
	if err != nil {
		return err
	}
	if got != n {
		return fmt.Errorf("function has %d versions, want %d", got, n)
	}
	return nil
}

func (w *world) versionShouldBeActive(n int) error {
	if err := w.viewVersions(); err != nil {
		return err
	}
	// The active row shows its version but no action buttons (activate/delete
	// only appear on non-active rows).
	var ok bool
	js := fmt.Sprintf(`(() => {
		const rows = [...document.querySelectorAll('tbody tr')];
		for (const r of rows) {
			if (r.querySelector('button')) continue;
			const span = r.querySelector('td span');
			if (span && span.textContent.trim() === 'v%d') return true;
		}
		return false;
	})()`, n)
	if err := w.run(chromedp.Evaluate(js, &ok)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("version %d is not the active version", n)
	}
	return nil
}

func (w *world) rollBackTo(n int) error {
	if err := w.viewVersions(); err != nil {
		return err
	}
	var res string
	js := fmt.Sprintf(`(() => {
		const rows = [...document.querySelectorAll('tbody tr')];
		for (const r of rows) {
			const span = r.querySelector('td span');
			if (span && span.textContent.trim() === 'v%d') {
				const btn = r.querySelector('button.btn--outline');
				if (btn) { btn.click(); return 'ok'; }
				return 'no-activate';
			}
		}
		return 'no-row';
	})()`, n)
	if err := w.run(chromedp.Evaluate(js, &res)); err != nil {
		return err
	}
	if res != "ok" {
		return fmt.Errorf("could not roll back to version %d: %s", n, res)
	}
	return w.run(chromedp.Sleep(1200 * time.Millisecond))
}

func (w *world) compareVersionsStep(a, b int) error {
	if err := w.viewVersions(); err != nil {
		return err
	}
	for _, n := range []int{a, b} {
		var res string
		js := fmt.Sprintf(`(() => {
			const rows = [...document.querySelectorAll('tbody tr')];
			for (const r of rows) {
				const span = r.querySelector('td span');
				if (span && span.textContent.trim() === 'v%d') {
					const cb = r.querySelector('input[type=checkbox]');
					if (cb) { cb.click(); return 'ok'; }
				}
			}
			return 'no';
		})()`, n)
		if err := w.run(chromedp.Evaluate(js, &res)); err != nil {
			return err
		}
		if res != "ok" {
			return fmt.Errorf("could not select version %d for comparison", n)
		}
	}
	if err := w.run(chromedp.Sleep(200 * time.Millisecond)); err != nil {
		return err
	}
	if err := w.click(`.versions-tab-container .card__footer button`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(800 * time.Millisecond))
}

func (w *world) shouldSeeDiff() error {
	if err := w.urlContains("/diff/"); err != nil {
		return err
	}
	ok, err := w.exists(`.diff-table`)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected a diff table on the comparison page")
	}
	return nil
}

// ── invocation ──────────────────────────────────────────────────────────────

func (w *world) callFunction() error { return w.invoke(http.MethodGet, w.lastFunc, "", "", nil) }

func (w *world) callWithMethod(method string) error {
	return w.invoke(method, w.lastFunc, "", "", nil)
}

func (w *world) callAtPath(suffix string) error {
	return w.invoke(http.MethodGet, w.lastFunc, suffix, "", nil)
}

func (w *world) respondedOK() error { return w.assertStatus(http.StatusOK) }

func (w *world) responseContains(want string) error {
	if !strings.Contains(w.respBody, want) {
		return fmt.Errorf("response %q does not contain %q", w.respBody, want)
	}
	return nil
}

func (w *world) responseEquals(want string) error {
	if w.respBody != want {
		return fmt.Errorf("response = %q, want %q", w.respBody, want)
	}
	return nil
}

func (w *world) callForbidden() error { return w.assertStatus(http.StatusForbidden) }
func (w *world) callNotFound() error  { return w.assertStatus(http.StatusNotFound) }

func (w *world) assertStatus(want int) error {
	if w.resp == nil {
		return fmt.Errorf("the function has not been called yet")
	}
	if w.resp.StatusCode != want {
		return fmt.Errorf("status = %d, want %d (body: %s)", w.resp.StatusCode, want, w.respBody)
	}
	return nil
}

// ── settings ────────────────────────────────────────────────────────────────

func (w *world) openSettings() error {
	id, err := w.functionID(w.lastFunc)
	if err != nil {
		return err
	}
	if err := w.navigate("#!/functions/" + id + "/settings"); err != nil {
		return err
	}
	return w.waitVisible(`.settings-tab-container`)
}

func (w *world) disableFunction() error {
	if err := w.openSettings(); err != nil {
		return err
	}
	if err := w.waitVisible(`#enable-function`); err != nil {
		return err
	}
	// #enable-function checked means enabled; uncheck to disable.
	var res string
	js := `(() => { const cb = document.querySelector('#enable-function'); if (!cb) return 'no'; if (cb.checked) cb.click(); return 'ok'; })()`
	if err := w.run(chromedp.Evaluate(js, &res)); err != nil {
		return err
	}
	if err := w.run(chromedp.Sleep(300 * time.Millisecond)); err != nil {
		return err
	}
	if err := w.saveCardOf(`#enable-function`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(1 * time.Second))
}

func (w *world) renameFunction(newName string) error {
	if err := w.openSettings(); err != nil {
		return err
	}
	nameInput := `.settings-tab-container input.form-input`
	if err := w.run(
		chromedp.WaitVisible(nameInput, chromedp.ByQuery),
		// Replace the existing name: focus, select all, type the new one.
		chromedp.Click(nameInput, chromedp.ByQuery),
		chromedp.KeyEvent("\x01"), // Ctrl-A (select all) is unreliable; clear via JS instead
	); err != nil {
		return err
	}
	// Clear and set the value via the input's native setter, then fire input.
	var ok bool
	lit, _ := json.Marshal(newName)
	js := fmt.Sprintf(`(() => {
		const el = document.querySelector(%q);
		if (!el) return false;
		const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
		setter.call(el, %s);
		el.dispatchEvent(new Event('input', { bubbles: true }));
		return true;
	})()`, nameInput, lit)
	if err := w.run(chromedp.Evaluate(js, &ok)); err != nil {
		return err
	}
	if err := w.run(chromedp.Sleep(300 * time.Millisecond)); err != nil {
		return err
	}
	if err := w.saveCardOf(nameInput); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(1 * time.Second))
}

func (w *world) deleteFunctionStep() error {
	if err := w.openSettings(); err != nil {
		return err
	}
	// The danger card's destructive button; the confirm() dialog is auto-accepted.
	if err := w.click(`.card--danger button.btn--destructive`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(1200 * time.Millisecond))
}

func (w *world) setEnvVar(key, value string) error {
	if err := w.openSettings(); err != nil {
		return err
	}
	if err := w.waitVisible(`.env-editor__actions button`); err != nil {
		return err
	}
	if err := w.jsClick(`.env-editor__actions button`); err != nil {
		return err
	}
	if err := w.run(chromedp.Sleep(300 * time.Millisecond)); err != nil {
		return err
	}
	keySel := `.env-editor__rows .env-editor__row:last-child .env-editor__key input`
	valSel := `.env-editor__rows .env-editor__row:last-child .env-editor__value input`
	if err := w.run(
		chromedp.SendKeys(keySel, key, chromedp.ByQuery),
		chromedp.SendKeys(valSel, value, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
	); err != nil {
		return err
	}
	if err := w.saveCardOf(`.env-editor`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(1 * time.Second))
}

func (w *world) scheduleEvery5Min() error {
	if err := w.openSettings(); err != nil {
		return err
	}
	cron := `input[placeholder="*/5 * * * *"]`
	if err := w.waitVisible(cron); err != nil {
		return err
	}
	if err := w.sendKeys(cron, "*/5 * * * *"); err != nil {
		return err
	}
	// Enable the schedule.
	var res string
	js := `(() => { const cb = document.querySelector('#enable-schedule'); if (!cb) return 'no'; if (!cb.checked) cb.click(); return 'ok'; })()`
	if err := w.run(chromedp.Evaluate(js, &res), chromedp.Sleep(300*time.Millisecond)); err != nil {
		return err
	}
	if err := w.saveCardOf(`#enable-schedule`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(1 * time.Second))
}

func (w *world) shouldSeeNextRun() error {
	if err := w.openSettings(); err != nil {
		return err
	}
	if err := w.waitVisible(`.next-run-info`); err != nil {
		return fmt.Errorf("expected the next-run time to be shown: %w", err)
	}
	return nil
}

// ── KV store ────────────────────────────────────────────────────────────────

func (w *world) openData() error {
	id, err := w.functionID(w.lastFunc)
	if err != nil {
		return err
	}
	if err := w.navigate("#!/functions/" + id + "/kv"); err != nil {
		return err
	}
	return w.waitVisible(`.kv-tab-container`)
}

// The data tab has two stores rendered as sibling cards: the function-scoped
// store first, the global store second.
const (
	scopedStoreCard = `.kv-tab-container .card:nth-child(1)`
	globalStoreCard = `.kv-tab-container .card:nth-child(2)`
)

func (w *world) storeScopedValue(value, key string) error {
	return w.storeValue(scopedStoreCard, value, key)
}
func (w *world) storeGlobalValue(value, key string) error {
	return w.storeValue(globalStoreCard, value, key)
}

func (w *world) storedValueKept(value, key string) error {
	return w.storedValueKeptIn(scopedStoreCard, value, key)
}

func (w *world) storedGlobalValueKept(value, key string) error {
	return w.storedValueKeptIn(globalStoreCard, value, key)
}

func (w *world) storeValue(card, value, key string) error {
	if err := w.openData(); err != nil {
		return err
	}
	if err := w.waitVisible(card + ` .env-editor__actions button`); err != nil {
		return err
	}
	if err := w.jsClick(card + ` .env-editor__actions button`); err != nil {
		return err
	}
	if err := w.run(chromedp.Sleep(300 * time.Millisecond)); err != nil {
		return err
	}
	if err := w.run(
		chromedp.SendKeys(card+` .env-editor__row:last-child .env-editor__key input`, key, chromedp.ByQuery),
		chromedp.SendKeys(card+` .env-editor__row:last-child .env-editor__value input`, value, chromedp.ByQuery),
		chromedp.Sleep(200*time.Millisecond),
	); err != nil {
		return err
	}
	if err := w.saveCardOf(card + ` .env-editor`); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(1 * time.Second))
}

func (w *world) storedValueKeptIn(card, value, key string) error {
	if err := w.openData(); err != nil {
		return err
	}
	if err := w.waitVisible(card + ` .env-editor__key input`); err != nil {
		return err
	}
	var ok bool
	jsKey, _ := json.Marshal(key)
	jsVal, _ := json.Marshal(value)
	js := fmt.Sprintf(`(() => {
		const keys = [...document.querySelectorAll(%q + ' .env-editor__key input')];
		return keys.some((k) => {
			if (k.value !== %s) return false;
			const row = k.closest('.env-editor__row');
			const v = row && row.querySelector('.env-editor__value input');
			return v && v.value === %s;
		});
	})()`, card, jsKey, jsVal)
	if err := w.run(chromedp.Evaluate(js, &ok)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("value %q under key %q was not kept", value, key)
	}
	return nil
}

// ── Test page ───────────────────────────────────────────────────────────────

func (w *world) tryFromTestPage() error {
	id, err := w.functionID(w.lastFunc)
	if err != nil {
		return err
	}
	if err := w.navigate("#!/functions/" + id + "/test"); err != nil {
		return err
	}
	// The request builder's Execute button is the only full-width button here.
	execute := `.test-panels .btn--full-width`
	if err := w.waitVisible(execute); err != nil {
		return err
	}
	if err := w.click(execute); err != nil {
		return err
	}
	// The response appears once the request resolves (CodeViewer renders).
	return w.waitJSTrue(`!!document.querySelector('.response-viewer .code-viewer')`)
}

func (w *world) testPageSuccessful() error {
	ok, err := w.selectorContains(`.response-viewer`, "200")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected the Test page to show a 200 response")
	}
	return nil
}

func (w *world) testPageResponseContains(want string) error {
	ok, err := w.selectorContains(`.response-viewer`, want)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected the Test page response to contain %q", want)
	}
	return nil
}

// ── executions ──────────────────────────────────────────────────────────────

func (w *world) viewExecutions() error {
	id, err := w.functionID(w.lastFunc)
	if err != nil {
		return err
	}
	if err := w.navigate("#!/functions/" + id + "/executions"); err != nil {
		return err
	}
	return w.waitVisible(`.function-details-title`)
}

func (w *world) executionsEmpty() error {
	if err := w.viewExecutions(); err != nil {
		return err
	}
	if err := w.run(chromedp.Sleep(300 * time.Millisecond)); err != nil {
		return err
	}
	got, err := w.count(`tbody tr`)
	if err != nil {
		return err
	}
	if got != 0 {
		return fmt.Errorf("expected no executions, found %d rows", got)
	}
	return nil
}

func (w *world) shouldHaveExecutions(n int) error {
	if err := w.viewExecutions(); err != nil {
		return err
	}
	if err := w.waitVisible(`tbody tr`); err != nil {
		return err
	}
	got, err := w.count(`tbody tr`)
	if err != nil {
		return err
	}
	if got != n {
		return fmt.Errorf("function has %d executions, want %d", got, n)
	}
	return nil
}

func (w *world) latestExecutionSuccessful() error {
	if err := w.viewExecutions(); err != nil {
		return err
	}
	if err := w.waitVisible(`tbody tr`); err != nil {
		return err
	}
	found, err := w.bodyContains("success")
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("expected a successful execution to be listed")
	}
	return nil
}

func (w *world) openLatestExecution() error {
	if err := w.viewExecutions(); err != nil {
		return err
	}
	if err := w.waitVisible(`tbody tr`); err != nil {
		return err
	}
	// Newest first: open the top row, which routes to the execution detail page.
	var ok bool
	js := `(() => { const r = document.querySelector('tbody tr'); if (!r) return false; r.click(); return true; })()`
	if err := w.run(chromedp.Evaluate(js, &ok)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no execution row to open")
	}
	return w.run(chromedp.Sleep(800 * time.Millisecond))
}

func (w *world) executionDetailSuccessful() error {
	if err := w.urlContains("/executions/"); err != nil {
		return err
	}
	if err := w.waitVisible(`.function-details-title`); err != nil {
		return err
	}
	// The detail header carries a status badge ("success" for a clean run).
	ok, err := w.selectorContains(`.function-details-title`, "success")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected the execution to be shown as successful")
	}
	return nil
}

// ── connected clients ───────────────────────────────────────────────────────

func (w *world) openClients() error {
	if err := w.navigate("#!/clients"); err != nil {
		return err
	}
	return w.run(chromedp.Sleep(400 * time.Millisecond))
}

func (w *world) clientsShown() error {
	if err := w.urlContains("clients"); err != nil {
		return err
	}
	return w.waitVisible(`.page-header h1`)
}

// ── shared browser assertions / helpers ─────────────────────────────────────

func (w *world) urlContains(want string) error {
	u, err := w.currentURL()
	if err != nil {
		return err
	}
	if !strings.Contains(u, want) {
		return fmt.Errorf("URL %q does not contain %q", u, want)
	}
	return nil
}

func (w *world) urlNotContains(notWant string) error {
	u, err := w.currentURL()
	if err != nil {
		return err
	}
	if strings.Contains(u, notWant) {
		return fmt.Errorf("URL %q should not contain %q", u, notWant)
	}
	return nil
}

// waitJSTrue polls a boolean JS expression until it is true (up to ~10s).
func (w *world) waitJSTrue(expr string) error {
	for range 40 {
		var ok bool
		if err := w.run(chromedp.Evaluate(expr, &ok)); err != nil {
			return err
		}
		if ok {
			return nil
		}
		if err := w.run(chromedp.Sleep(250 * time.Millisecond)); err != nil {
			return err
		}
	}
	return fmt.Errorf("condition never became true: %s", expr)
}

// waitEnabled waits until a button matching sel exists and is not disabled.
func (w *world) waitEnabled(sel string) error {
	expr := fmt.Sprintf(`(() => { const b = document.querySelector(%q); return !!b && !b.disabled; })()`, sel)
	return w.waitJSTrue(expr)
}

// jsClick clicks an element via the DOM (used where a native click would need
// scrolling or precise hit-testing, e.g. add-row buttons inside scrollable cards).
func (w *world) jsClick(sel string) error {
	var ok bool
	js := fmt.Sprintf(`(() => { const e = document.querySelector(%q); if (!e) return false; e.click(); return true; })()`, sel)
	if err := w.run(chromedp.Evaluate(js, &ok)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no element to click for %q", sel)
	}
	return nil
}

// saveCardOf clicks the "Save Changes" button in the card footer that contains
// the control matched by controlSel, once that button is enabled.
func (w *world) saveCardOf(controlSel string) error {
	cond := fmt.Sprintf(`(() => {
		const c = document.querySelector(%q);
		if (!c) return false;
		const card = c.closest('.card');
		if (!card) return false;
		const b = card.querySelector('.card__footer button');
		return !!b && !b.disabled;
	})()`, controlSel)
	if err := w.waitJSTrue(cond); err != nil {
		return fmt.Errorf("save button never enabled for %q: %w", controlSel, err)
	}
	var ok bool
	js := fmt.Sprintf(`(() => { const c = document.querySelector(%q); const b = c.closest('.card').querySelector('.card__footer button'); b.click(); return true; })()`, controlSel)
	return w.run(chromedp.Evaluate(js, &ok))
}
