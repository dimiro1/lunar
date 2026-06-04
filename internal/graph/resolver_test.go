package graph_test

import (
	"context"
	"maps"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/dimiro1/lunar/internal/graph"
	"github.com/dimiro1/lunar/internal/services/env"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/store"
)

// fakeEnvStore is an in-memory env.Store that also records how often All is
// called, so tests can assert the env store is untouched unless envVars is
// selected. Unused interface methods are inherited from the embedded nil.
type fakeEnvStore struct {
	env.Store
	vars     map[string]string
	allCalls int
}

func (f *fakeEnvStore) Set(_, key, value string) error { f.vars[key] = value; return nil }
func (f *fakeEnvStore) Delete(_, key string) error     { delete(f.vars, key); return nil }
func (f *fakeEnvStore) All(string) (map[string]string, error) {
	f.allCalls++
	return maps.Clone(f.vars), nil
}

// fakeKVStore is an in-memory kv.Store keyed by scope ("" == global) that records
// All / AllGlobal calls.
type fakeKVStore struct {
	kv.Store
	data           map[string]map[string]string
	allCalls       int
	allGlobalCalls int
}

func (f *fakeKVStore) scope(s string) map[string]string {
	if f.data[s] == nil {
		f.data[s] = map[string]string{}
	}
	return f.data[s]
}
func (f *fakeKVStore) Set(functionID, key, value string) error { f.scope(functionID)[key] = value; return nil }
func (f *fakeKVStore) Delete(functionID, key string) error     { delete(f.scope(functionID), key); return nil }
func (f *fakeKVStore) All(functionID string) (map[string]string, error) {
	f.allCalls++
	return maps.Clone(f.scope(functionID)), nil
}
func (f *fakeKVStore) AllGlobal() (map[string]string, error) {
	f.allGlobalCalls++
	return maps.Clone(f.scope("")), nil
}

// newTestClient builds the real gqlgen server (same constructor used in
// production wiring) backed by an in-memory store and counting env/kv fakes, so
// these tests exercise the schema → resolver → store path end to end without
// HTTP or auth.
func newTestClient(t *testing.T) (*client.Client, store.DB, *fakeEnvStore, *fakeKVStore) {
	t.Helper()
	db := store.NewMemoryDB()
	envStore := &fakeEnvStore{vars: map[string]string{}}
	kvStore := &fakeKVStore{data: map[string]map[string]string{}}
	srv := graph.NewServer(&graph.Resolver{DB: db, EnvStore: envStore, KVStore: kvStore})
	return client.New(srv), db, envStore, kvStore
}

func seedFunction(t *testing.T, db store.DB, id, name, code string) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.CreateFunction(ctx, store.Function{ID: id, Name: name}); err != nil {
		t.Fatalf("CreateFunction(%q): %v", id, err)
	}
	if _, err := db.CreateVersion(ctx, id, code, "", nil); err != nil {
		t.Fatalf("CreateVersion(%q): %v", id, err)
	}
}

func TestFunctionsQuery(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	seedFunction(t, db, "fn1", "hello", "return 1")

	var resp struct {
		Functions struct {
			Nodes []struct {
				ID            string
				Name          string
				ActiveVersion struct {
					Version  int
					Code     string
					IsActive bool
				}
			}
			PageInfo struct {
				Total  int
				Limit  int
				Offset int
			}
		}
	}

	c.MustPost(`{
		functions {
			nodes { id name activeVersion { version code isActive } }
			pageInfo { total limit offset }
		}
	}`, &resp)

	if got := resp.Functions.PageInfo.Total; got != 1 {
		t.Fatalf("pageInfo.total = %d, want 1", got)
	}
	if got := resp.Functions.PageInfo.Limit; got != 20 {
		t.Errorf("pageInfo.limit = %d, want default 20", got)
	}
	if n := len(resp.Functions.Nodes); n != 1 {
		t.Fatalf("got %d nodes, want 1", n)
	}
	node := resp.Functions.Nodes[0]
	if node.ID != "fn1" || node.Name != "hello" {
		t.Errorf("node = {id:%q name:%q}, want {fn1 hello}", node.ID, node.Name)
	}
	if node.ActiveVersion.Code != "return 1" || !node.ActiveVersion.IsActive {
		t.Errorf("activeVersion = %+v, want code=%q isActive=true", node.ActiveVersion, "return 1")
	}
}

func TestVersionsQuery(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	seedFunction(t, db, "fn1", "hello", "return 1")

	var resp struct {
		Versions struct {
			Nodes []struct {
				Version  int
				Code     string
				IsActive bool
			}
			PageInfo struct{ Total int }
		}
	}

	c.MustPost(`{
		versions(functionId: "fn1") {
			nodes { version code isActive }
			pageInfo { total }
		}
	}`, &resp)

	if got := resp.Versions.PageInfo.Total; got != 1 {
		t.Fatalf("versions total = %d, want 1", got)
	}
	if n := len(resp.Versions.Nodes); n != 1 {
		t.Fatalf("got %d versions, want 1", n)
	}
	if v := resp.Versions.Nodes[0]; v.Version != 1 || v.Code != "return 1" || !v.IsActive {
		t.Errorf("version = %+v, want {1 \"return 1\" true}", v)
	}
}

// TestGraphEdges exercises the relation edges that let the graph be traversed in
// both directions: Function.versions[].function (reverse), Function.executions,
// Function.nextRun, Execution.function, and Execution.version. The logs/AI/email
// edges need the logger/trackers (covered by live verification) and are omitted
// here, where only the DB is wired.
func TestGraphEdges(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	ctx := context.Background()
	if _, err := db.CreateFunction(ctx, store.Function{ID: "fn1", Name: "edges"}); err != nil {
		t.Fatalf("CreateFunction: %v", err)
	}
	ver, err := db.CreateVersion(ctx, "fn1", "return 1", "", nil)
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}
	if _, err := db.CreateExecution(ctx, store.Execution{
		ID:                "ex1",
		FunctionID:        "fn1",
		FunctionVersionID: ver.ID,
		Status:            store.ExecutionStatus("success"),
		Trigger:           store.ExecutionTriggerHTTP,
	}); err != nil {
		t.Fatalf("CreateExecution: %v", err)
	}

	var resp struct {
		Function struct {
			Name    string
			NextRun struct{ HasSchedule bool }
			Versions struct {
				Nodes []struct {
					Version  int
					Function struct{ ID, Name string }
				}
			}
			Executions struct {
				Nodes []struct {
					ID       string
					Function struct{ ID string }
					Version  struct {
						ID      string
						Version int
					}
				}
			}
		}
	}

	c.MustPost(`{
		function(id: "fn1") {
			name
			nextRun { hasSchedule }
			versions { nodes { version function { id name } } }
			executions { nodes { id function { id } version { id version } } }
		}
	}`, &resp)

	f := resp.Function
	if f.NextRun.HasSchedule {
		t.Error("nextRun.hasSchedule = true, want false (no cron configured)")
	}
	if n := len(f.Versions.Nodes); n != 1 {
		t.Fatalf("got %d versions, want 1", n)
	}
	if fn := f.Versions.Nodes[0].Function; fn.ID != "fn1" || fn.Name != "edges" {
		t.Errorf("versions[0].function (reverse edge) = %+v, want {fn1 edges}", fn)
	}
	if n := len(f.Executions.Nodes); n != 1 {
		t.Fatalf("got %d executions, want 1", n)
	}
	e := f.Executions.Nodes[0]
	if e.ID != "ex1" {
		t.Errorf("executions[0].id = %q, want ex1", e.ID)
	}
	if e.Function.ID != "fn1" {
		t.Errorf("execution.function.id = %q, want fn1", e.Function.ID)
	}
	if e.Version.ID != ver.ID || e.Version.Version != 1 {
		t.Errorf("execution.version = %+v, want {%s 1}", e.Version, ver.ID)
	}
}

func TestFunctionEnvKvLazy(t *testing.T) {
	c, db, envStore, kvStore := newTestClient(t)
	envStore.vars = map[string]string{"API_KEY": "secret"}
	kvStore.data = map[string]map[string]string{
		"fn1": {"counter": "1"},
		"":    {"shared": "x"},
	}
	seedFunction(t, db, "fn1", "hello", "return 1")

	// Selecting only scalar fields must NOT touch the env/kv stores — this is the
	// overfetch fix: the list/detail view pays for env/kv only when it asks.
	var bare struct {
		Function *struct {
			ID   string
			Name string
		}
	}
	c.MustPost(`{ function(id: "fn1") { id name } }`, &bare)
	if envStore.allCalls != 0 || kvStore.allCalls != 0 || kvStore.allGlobalCalls != 0 {
		t.Fatalf("scalar-only query touched stores; env=%d kvScoped=%d kvGlobal=%d, want all 0",
			envStore.allCalls, kvStore.allCalls, kvStore.allGlobalCalls)
	}

	// Selecting the map fields fetches them lazily, exactly once each.
	var withMaps struct {
		Function *struct {
			EnvVars    map[string]string
			ScopedData map[string]string
			GlobalData map[string]string
		}
	}
	c.MustPost(`{ function(id: "fn1") { envVars scopedData globalData } }`, &withMaps)
	if withMaps.Function == nil {
		t.Fatal("function(fn1) = nil")
	}
	if got := withMaps.Function.EnvVars["API_KEY"]; got != "secret" {
		t.Errorf("envVars[API_KEY] = %q, want secret", got)
	}
	if got := withMaps.Function.ScopedData["counter"]; got != "1" {
		t.Errorf("scopedData[counter] = %q, want 1", got)
	}
	if got := withMaps.Function.GlobalData["shared"]; got != "x" {
		t.Errorf("globalData[shared] = %q, want x", got)
	}
	if envStore.allCalls != 1 || kvStore.allCalls != 1 || kvStore.allGlobalCalls != 1 {
		t.Errorf("expected one lookup each; got env=%d kvScoped=%d kvGlobal=%d",
			envStore.allCalls, kvStore.allCalls, kvStore.allGlobalCalls)
	}
}

func TestFunctionByID(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	seedFunction(t, db, "fn1", "hello", "return 1")

	var resp struct {
		Function *struct {
			ID            string
			ActiveVersion struct{ Code string }
		}
		Missing *struct{ ID string }
	}

	c.MustPost(`{
		function(id: "fn1") { id activeVersion { code } }
		missing: function(id: "does-not-exist") { id }
	}`, &resp)

	if resp.Function == nil {
		t.Fatal("function(fn1) = nil, want a result")
	}
	if resp.Function.ID != "fn1" || resp.Function.ActiveVersion.Code != "return 1" {
		t.Errorf("function = %+v, want id=fn1 code=%q", resp.Function, "return 1")
	}
	// A missing function resolves to null rather than erroring.
	if resp.Missing != nil {
		t.Errorf("missing function = %+v, want nil", resp.Missing)
	}
}

// TestExecutionRequestConnectionsWithoutTrackers verifies that the AI and email
// request connections degrade to an empty result — rather than panicking — when
// no tracker is wired (newTestClient leaves AITracker and EmailTracker nil, as
// happens whenever AI/email tracking is not configured).
func TestExecutionRequestConnectionsWithoutTrackers(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	seedFunction(t, db, "fn1", "hello", "return 1")

	// The connections are only reached once the execution exists, so seed one.
	ctx := context.Background()
	version, err := db.GetActiveVersion(ctx, "fn1")
	if err != nil {
		t.Fatalf("GetActiveVersion: %v", err)
	}
	if _, err := db.CreateExecution(ctx, store.Execution{
		ID:                "exec1",
		FunctionID:        "fn1",
		FunctionVersionID: version.ID,
		Status:            store.ExecutionStatusSuccess,
		Trigger:           store.ExecutionTriggerHTTP,
	}); err != nil {
		t.Fatalf("CreateExecution: %v", err)
	}

	var resp struct {
		AI struct {
			Nodes    []struct{ ID string }
			PageInfo struct{ Total int }
		} `json:"executionAiRequests"`
		Email struct {
			Nodes    []struct{ ID string }
			PageInfo struct{ Total int }
		} `json:"executionEmailRequests"`
	}

	// MustPost fails the test on any GraphQL error, so a resolver panic (which
	// gqlgen surfaces as an error) would be caught here.
	c.MustPost(`{
		executionAiRequests(executionId: "exec1") { nodes { id } pageInfo { total } }
		executionEmailRequests(executionId: "exec1") { nodes { id } pageInfo { total } }
	}`, &resp)

	if n := len(resp.AI.Nodes); n != 0 || resp.AI.PageInfo.Total != 0 {
		t.Errorf("aiRequests = %d nodes / total %d, want empty", n, resp.AI.PageInfo.Total)
	}
	if n := len(resp.Email.Nodes); n != 0 || resp.Email.PageInfo.Total != 0 {
		t.Errorf("emailRequests = %d nodes / total %d, want empty", n, resp.Email.PageInfo.Total)
	}
}
