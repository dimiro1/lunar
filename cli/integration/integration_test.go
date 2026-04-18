//go:build integration

// Package integration runs end-to-end tests against a real Lunar server.
//
// The server starts in-process via httptest.NewServer with an in-memory SQLite
// database — no binary build required, tests are isolated per run.
//
// Run with:
//
//	go test -tags integration -v ./...
package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dimiro1/lunar/cli/cmd"
	"github.com/dimiro1/lunar/frontend"
	"github.com/dimiro1/lunar/internal/api"
	"github.com/dimiro1/lunar/internal/migrate"
	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/store"
	_ "modernc.org/sqlite"
)

const testAPIKey = "integration-test-key-lunar-cli"

var testServer *httptest.Server

// TestMain starts a single in-process server for the entire test run.
func TestMain(m *testing.M) {
	ts, cleanup := newTestServer()
	testServer = ts
	defer cleanup()
	m.Run()
}

// newTestServer creates an in-process Lunar server backed by an in-memory
// SQLite database. Returns the server and a cleanup function.
func newTestServer() (*httptest.Server, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(fmt.Sprintf("open db: %v", err))
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		panic(fmt.Sprintf("foreign keys: %v", err))
	}
	if err := migrate.Run(db, migrate.FS); err != nil {
		panic(fmt.Sprintf("migrate: %v", err))
	}

	apiDB := store.NewSQLiteDB(db)
	srv := api.NewServer(api.ServerConfig{
		DB:               apiDB,
		Logger:           logger.NewSQLiteLogger(db),
		KVStore:          kv.NewSQLiteStore(db),
		EnvStore:         env.NewSQLiteStore(db),
		HTTPClient:       internalhttp.NewDefaultClient(),
		ExecutionTimeout: 30 * time.Second,
		FrontendHandler:  frontend.Handler(),
		APIKey:           testAPIKey,
		BaseURL:          "http://localhost:8080",
	})

	ts := httptest.NewServer(srv.Handler())
	return ts, func() {
		ts.Close()
		_ = db.Close()
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// run executes CLI args in-process and returns the captured output.
// Global --server, --token, and --output json are prepended automatically.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	full := append(
		[]string{"--server", testServer.URL, "--token", testAPIKey, "--output", "json"},
		args...,
	)
	var buf bytes.Buffer
	err := cmd.Run(full, &buf)
	return buf.String(), err
}

// must calls run and fails the test if there is an error.
func must(t *testing.T, args ...string) string {
	t.Helper()
	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("cmd %v: %v\noutput: %s", args, err, out)
	}
	return out
}

// parseJSON unmarshals CLI output JSON into v.
func parseJSON(t *testing.T, out string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), v); err != nil {
		t.Fatalf("parseJSON: %v\nraw output: %s", err, out)
	}
}

// createFunction is a test helper that creates a function and registers
// a cleanup to delete it, returning the function ID.
func createFunction(t *testing.T, name string) string {
	t.Helper()
	out := must(t, "functions", "create",
		"--name", name,
		"--code", `function handler(ctx, event) return { statusCode = 200, body = "ok" } end`,
	)
	var fn struct {
		ID string `json:"id"`
	}
	parseJSON(t, out, &fn)
	if fn.ID == "" {
		t.Fatal("create returned empty id")
	}
	t.Cleanup(func() {
		if _, err := run(t, "functions", "delete", fn.ID); err != nil {
			t.Logf("cleanup delete %s: %v", fn.ID, err)
		}
	})
	return fn.ID
}

// firstVersion returns the ID and number of the first version of a function.
func firstVersion(t *testing.T, fnID string) (id string, number string) {
	t.Helper()
	out := must(t, "versions", "list", fnID)
	var resp struct {
		Versions []struct {
			ID      string `json:"id"`
			Version int    `json:"version"`
		} `json:"versions"`
	}
	parseJSON(t, out, &resp)
	if len(resp.Versions) == 0 {
		t.Fatal("no versions found")
	}
	return resp.Versions[0].ID, fmt.Sprintf("%d", resp.Versions[0].Version)
}

// ── functions ─────────────────────────────────────────────────────────────────

func TestFunctions_List(t *testing.T) {
	out := must(t, "functions", "list")
	var resp struct {
		Functions  any `json:"functions"`
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	parseJSON(t, out, &resp)
}

func TestFunctions_CreateAndGet(t *testing.T) {
	id := createFunction(t, "create-get-test")

	out := must(t, "functions", "get", id)
	var fn struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	parseJSON(t, out, &fn)

	if fn.ID != id {
		t.Errorf("get id=%q, want %q", fn.ID, id)
	}
	if fn.Name != "create-get-test" {
		t.Errorf("get name=%q, want create-get-test", fn.Name)
	}
}

func TestFunctions_CreateAppearsInList(t *testing.T) {
	id := createFunction(t, "list-visibility-test")

	out := must(t, "functions", "list")
	if !strings.Contains(out, id) {
		t.Errorf("created function %q not found in list output", id)
	}
}

func TestFunctions_Delete(t *testing.T) {
	out := must(t, "functions", "create",
		"--name", "delete-test",
		"--code", `function handler(ctx, event) return { statusCode = 200 } end`,
	)
	var fn struct {
		ID string `json:"id"`
	}
	parseJSON(t, out, &fn)

	must(t, "functions", "delete", fn.ID)

	listOut := must(t, "functions", "list")
	if strings.Contains(listOut, fn.ID) {
		t.Errorf("deleted function %q still appears in list", fn.ID)
	}
}

func TestFunctions_UpdateName(t *testing.T) {
	id := createFunction(t, "update-test")

	must(t, "functions", "update", id, "--name", "updated-name")

	out := must(t, "functions", "get", id)
	if !strings.Contains(out, "updated-name") {
		t.Errorf("updated name not found in get output: %s", out)
	}
}

func TestFunctions_UpdateCode_CreatesNewVersion(t *testing.T) {
	id := createFunction(t, "update-code-test")

	newCode := `function handler(ctx, event) return { statusCode = 201, body = "v2" } end`
	must(t, "functions", "update", id, "--code", newCode)

	out := must(t, "versions", "list", id)
	var resp struct {
		Versions []any `json:"versions"`
	}
	parseJSON(t, out, &resp)
	if len(resp.Versions) < 2 {
		t.Errorf("expected at least 2 versions after code update, got %d", len(resp.Versions))
	}
}

func TestFunctions_EnvVars(t *testing.T) {
	id := createFunction(t, "env-vars-test")

	must(t, "functions", "env", id,
		"--env", "API_KEY=secret",
		"--env", "DEBUG=true",
	)

	out := must(t, "functions", "get", id)
	if !strings.Contains(out, "API_KEY") {
		t.Errorf("env var API_KEY not in get output: %s", out)
	}
	if !strings.Contains(out, "secret") {
		t.Errorf("env var value not in get output: %s", out)
	}
}

func TestFunctions_KVStore(t *testing.T) {
	id := createFunction(t, "kv-test")

	must(t, "functions", "kv", id,
		"--kv", "counter=0",
		"--kv", "region=useast",
		"--global=false",
	)

	out := must(t, "functions", "get", id)
	if !strings.Contains(out, "counter") {
		t.Errorf("kv key 'counter' not in get output: %s", out)
	}
}

// ── versions ──────────────────────────────────────────────────────────────────

func TestVersions_CreatedOnFunctionCreate(t *testing.T) {
	id := createFunction(t, "versions-test")

	out := must(t, "versions", "list", id)
	var resp struct {
		Versions []struct {
			ID       string `json:"id"`
			IsActive bool   `json:"is_active"`
		} `json:"versions"`
	}
	parseJSON(t, out, &resp)

	if len(resp.Versions) == 0 {
		t.Fatal("expected at least one version after create")
	}
	if !resp.Versions[0].IsActive {
		t.Error("first version should be active")
	}
}

func TestVersions_Get(t *testing.T) {
	id := createFunction(t, "version-get-test")
	vID, vNum := firstVersion(t, id)

	out := must(t, "versions", "get", id, vNum)
	if !strings.Contains(out, vID) {
		t.Errorf("version get output missing version id: %s", out)
	}
}

func TestVersions_ActivateRollback(t *testing.T) {
	id := createFunction(t, "activate-test")
	v1, _ := firstVersion(t, id)

	// Create a second version
	must(t, "functions", "update", id,
		"--code", `function handler(ctx, event) return { statusCode = 201 } end`,
	)

	// First version should no longer be active; roll back to it
	must(t, "versions", "activate", id, v1)

	out := must(t, "versions", "list", id)
	var resp struct {
		Versions []struct {
			ID       string `json:"id"`
			IsActive bool   `json:"is_active"`
		} `json:"versions"`
	}
	parseJSON(t, out, &resp)

	for _, v := range resp.Versions {
		if v.ID == v1 && !v.IsActive {
			t.Errorf("version %s should be active after activate", v1)
		}
	}
}

func TestVersions_Delete(t *testing.T) {
	id := createFunction(t, "version-delete-test")

	// Create a second version so we can delete the inactive one
	must(t, "functions", "update", id,
		"--code", `function handler(ctx, event) return { statusCode = 201 } end`,
	)

	out := must(t, "versions", "list", id)
	var resp struct {
		Versions []struct {
			ID       string `json:"id"`
			IsActive bool   `json:"is_active"`
		} `json:"versions"`
	}
	parseJSON(t, out, &resp)
	if len(resp.Versions) < 2 {
		t.Fatal("expected 2 versions")
	}

	var inactiveID string
	for _, v := range resp.Versions {
		if !v.IsActive {
			inactiveID = v.ID
			break
		}
	}
	if inactiveID == "" {
		t.Fatal("no inactive version to delete")
	}

	must(t, "versions", "delete", id, inactiveID)

	listOut := must(t, "versions", "list", id)
	if strings.Contains(listOut, inactiveID) {
		t.Errorf("deleted version %q still appears in list", inactiveID)
	}
}

// ── invoke ────────────────────────────────────────────────────────────────────

func TestInvoke_Get(t *testing.T) {
	id := createFunction(t, "invoke-get-test")

	out := must(t, "invoke", id)
	if !strings.Contains(out, "ok") {
		t.Errorf("invoke response missing expected body: %s", out)
	}
}

func TestInvoke_Post(t *testing.T) {
	id := createFunction(t, "invoke-post-test")

	out := must(t, "invoke", id, "--method", "POST", "--body", `{"hello":"world"}`)
	if strings.Contains(out, "error") {
		t.Errorf("unexpected error in invoke output: %s", out)
	}
}

// ── executions ───────────────────────────────────────────────────────────────

func TestExecutions_EmptyForNewFunction(t *testing.T) {
	id := createFunction(t, "executions-test")

	out := must(t, "executions", "list", id)
	var resp struct {
		Executions []any `json:"executions"`
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	parseJSON(t, out, &resp)

	if resp.Pagination.Total != 0 {
		t.Errorf("expected 0 executions for new function, got %d", resp.Pagination.Total)
	}
}

func TestExecutions_AppearsAfterInvoke(t *testing.T) {
	id := createFunction(t, "executions-invoke-test")

	must(t, "invoke", id)

	// Give the server a moment to persist the execution record.
	time.Sleep(100 * time.Millisecond)

	out := must(t, "executions", "list", id)
	var resp struct {
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	parseJSON(t, out, &resp)

	if resp.Pagination.Total == 0 {
		t.Error("expected at least one execution after invoke")
	}
}

// ── tokens ────────────────────────────────────────────────────────────────────

func TestTokens_List(t *testing.T) {
	out := must(t, "tokens", "list")
	var resp struct {
		Tokens     any `json:"tokens"`
		Pagination any `json:"pagination"`
	}
	parseJSON(t, out, &resp)
}

// ── auth ──────────────────────────────────────────────────────────────────────

func TestAuth_WrongToken_Fails(t *testing.T) {
	var buf bytes.Buffer
	args := []string{"--server", testServer.URL, "--token", "wrong-token", "--output", "json", "functions", "list"}
	err := cmd.Run(args, &buf)
	out := buf.String()

	if err == nil && strings.Contains(out, `"functions"`) {
		t.Errorf("expected auth failure with wrong token, got apparent success: %s", out)
	}
}

// ── skills ────────────────────────────────────────────────────────────────────

func TestSkills_List(t *testing.T) {
	out := must(t, "skills", "list")
	if !strings.Contains(out, "lunar-cli") {
		t.Errorf("expected 'lunar-cli' in skills list, got: %s", out)
	}
	if !strings.Contains(out, "lunar-lua") {
		t.Errorf("expected 'lunar-lua' in skills list, got: %s", out)
	}
}

func TestSkills_Show_CLI(t *testing.T) {
	out := must(t, "skills", "show", "lunar-cli")
	if !strings.Contains(out, "lunar-cli") {
		t.Errorf("expected skill name in output: %s", out)
	}
	if !strings.Contains(out, "functions create") {
		t.Errorf("expected CLI command reference in output: %s", out)
	}
}

func TestSkills_Show_Lua(t *testing.T) {
	out := must(t, "skills", "show", "lunar-lua")
	if !strings.Contains(out, "handler") {
		t.Errorf("expected handler reference in lua skill: %s", out)
	}
	if !strings.Contains(out, "llms.txt") {
		t.Errorf("expected llms.txt reference in lua skill: %s", out)
	}
}

func TestSkills_Show_Unknown(t *testing.T) {
	_, err := run(t, "skills", "show", "nonexistent-skill")
	if err == nil {
		t.Error("expected error for unknown skill name")
	}
}
