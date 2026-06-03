package graph_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/dimiro1/lunar/internal/store"
)

func TestFunctionMutations(t *testing.T) {
	c, db, envStore, _ := newTestClient(t)
	ctx := context.Background()

	var created struct {
		CreateFunction struct {
			ID            string
			Name          string
			ActiveVersion struct{ Code string }
		}
	}
	c.MustPost(`mutation {
		createFunction(input: {name: "hello", code: "return 1"}) {
			id name activeVersion { code }
		}
	}`, &created)
	id := created.CreateFunction.ID
	if id == "" || created.CreateFunction.Name != "hello" || created.CreateFunction.ActiveVersion.Code != "return 1" {
		t.Fatalf("createFunction = %+v", created.CreateFunction)
	}
	if _, err := db.GetFunction(ctx, id); err != nil {
		t.Fatalf("GetFunction after create: %v", err)
	}

	// Update metadata and supply new code → a new active version.
	var updated struct {
		UpdateFunction struct {
			Name          string
			ActiveVersion struct {
				Version int
				Code    string
			}
		}
	}
	c.MustPost(`mutation($id: ID!) {
		updateFunction(id: $id, input: {name: "renamed", code: "return 2"}) {
			name activeVersion { version code }
		}
	}`, &updated, client.Var("id", id))
	if updated.UpdateFunction.Name != "renamed" {
		t.Errorf("updated name = %q, want renamed", updated.UpdateFunction.Name)
	}
	if updated.UpdateFunction.ActiveVersion.Version != 2 || updated.UpdateFunction.ActiveVersion.Code != "return 2" {
		t.Errorf("updated activeVersion = %+v, want v2 'return 2'", updated.UpdateFunction.ActiveVersion)
	}

	// Replace env vars (Map scalar input), then read them back.
	var env struct {
		SetFunctionEnv struct {
			EnvVars map[string]string
		}
	}
	c.MustPost(`mutation($id: ID!, $env: Map!) {
		setFunctionEnv(id: $id, env: $env) { envVars }
	}`, &env, client.Var("id", id), client.Var("env", map[string]string{"API_KEY": "secret"}))
	if env.SetFunctionEnv.EnvVars["API_KEY"] != "secret" {
		t.Errorf("setFunctionEnv envVars = %+v", env.SetFunctionEnv.EnvVars)
	}
	if envStore.vars["API_KEY"] != "secret" {
		t.Errorf("env store not updated: %+v", envStore.vars)
	}

	// Delete, then confirm it is gone (null).
	var deleted struct{ DeleteFunction bool }
	c.MustPost(`mutation($id: ID!) { deleteFunction(id: $id) }`, &deleted, client.Var("id", id))
	if !deleted.DeleteFunction {
		t.Error("deleteFunction = false, want true")
	}
	var gone struct {
		Function *struct{ ID string }
	}
	c.MustPost(`query($id: ID!) { function(id: $id) { id } }`, &gone, client.Var("id", id))
	if gone.Function != nil {
		t.Errorf("function after delete = %+v, want nil", gone.Function)
	}
}

// TestExecutionEnums verifies the enums bound to the store string types marshal
// to their GraphQL enum values over the wire.
func TestExecutionEnums(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	ctx := context.Background()
	seedFunction(t, db, "fn1", "hello", "return 1")

	dur := int64(42)
	if _, err := db.CreateExecution(ctx, store.Execution{
		ID:                "exec1",
		FunctionID:        "fn1",
		FunctionVersionID: "v1",
		Status:            store.ExecutionStatusSuccess,
		Trigger:           store.ExecutionTriggerHTTP,
		DurationMs:        &dur,
		CreatedAt:         1000,
	}); err != nil {
		t.Fatalf("CreateExecution: %v", err)
	}

	var resp struct {
		Execution *struct {
			ID         string
			Status     string
			Trigger    string
			DurationMs *int
		}
	}
	c.MustPost(`{ execution(id: "exec1") { id status trigger durationMs } }`, &resp)
	if resp.Execution == nil {
		t.Fatal("execution(exec1) = nil")
	}
	if resp.Execution.Status != "success" {
		t.Errorf("status = %q, want success", resp.Execution.Status)
	}
	if resp.Execution.Trigger != "http" {
		t.Errorf("trigger = %q, want http", resp.Execution.Trigger)
	}
	if resp.Execution.DurationMs == nil || *resp.Execution.DurationMs != 42 {
		t.Errorf("durationMs = %v, want 42", resp.Execution.DurationMs)
	}

	var list struct {
		Executions struct {
			Nodes    []struct{ ID string }
			PageInfo struct{ Total int }
		}
	}
	c.MustPost(`{ executions(functionId: "fn1") { nodes { id } pageInfo { total } } }`, &list)
	if list.Executions.PageInfo.Total != 1 || len(list.Executions.Nodes) != 1 {
		t.Errorf("executions = %+v, want 1", list.Executions)
	}
}

func TestVersionDiff(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	ctx := context.Background()
	if _, err := db.CreateFunction(ctx, store.Function{ID: "fn1", Name: "f"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateVersion(ctx, "fn1", "line1\nline2", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateVersion(ctx, "fn1", "line1\nCHANGED", nil); err != nil {
		t.Fatal(err)
	}

	var resp struct {
		VersionDiff struct {
			OldVersion int
			NewVersion int
			Lines      []struct {
				LineType string
				Content  string
			}
		}
	}
	c.MustPost(`{
		versionDiff(functionId: "fn1", oldVersion: 1, newVersion: 2) {
			oldVersion newVersion lines { lineType content }
		}
	}`, &resp)

	if resp.VersionDiff.OldVersion != 1 || resp.VersionDiff.NewVersion != 2 {
		t.Errorf("versions = %d/%d, want 1/2", resp.VersionDiff.OldVersion, resp.VersionDiff.NewVersion)
	}
	var added, removed int
	for _, l := range resp.VersionDiff.Lines {
		switch l.LineType {
		case "added":
			added++
		case "removed":
			removed++
		}
	}
	if added == 0 || removed == 0 {
		t.Errorf("expected added & removed diff lines; got added=%d removed=%d", added, removed)
	}
}

func TestTokens(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	ctx := context.Background()
	if _, err := db.CreateAPIToken(ctx, store.APIToken{ID: "tok1", Name: "cli", TokenHash: "h"}); err != nil {
		t.Fatal(err)
	}

	var list struct {
		APITokens []struct {
			ID      string
			Name    string
			Revoked bool
		}
	}
	c.MustPost(`{ apiTokens { id name revoked } }`, &list)
	if len(list.APITokens) != 1 || list.APITokens[0].ID != "tok1" || list.APITokens[0].Revoked {
		t.Fatalf("apiTokens = %+v", list.APITokens)
	}

	var revoke struct{ RevokeApiToken bool }
	c.MustPost(`mutation { revokeApiToken(id: "tok1") }`, &revoke)
	if !revoke.RevokeApiToken {
		t.Error("revokeApiToken = false, want true")
	}

	c.MustPost(`{ apiTokens { id revoked } }`, &list)
	if len(list.APITokens) != 1 || !list.APITokens[0].Revoked {
		t.Errorf("after revoke = %+v, want revoked=true", list.APITokens)
	}
}
