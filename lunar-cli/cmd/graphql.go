package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	graphql "github.com/hasura/go-graphql-client"
)

// authTransport injects the bearer token into every request.
type authTransport struct {
	token string
	base  http.RoundTripper
}

func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	return t.base.RoundTrip(req)
}

// mustGraphQLClient builds a GraphQL client targeting <server>/graphql using the
// server URL and token resolved in root.go's PersistentPreRunE.
func mustGraphQLClient() *graphql.Client {
	if serverURL == "" {
		fmt.Fprintln(os.Stderr, "error: no server configured (use --server or LUNAR_SERVER)")
		os.Exit(1)
	}
	httpClient := &http.Client{
		Transport: authTransport{token: apiToken, base: http.DefaultTransport},
	}
	return graphql.NewClient(serverURL+"/graphql", httpClient)
}

// execRaw runs a GraphQL operation and decodes the top-level `data` object into
// a map keyed by operation/alias name. GraphQL and HTTP errors (including an
// unauthenticated 401) are returned as a non-nil error.
func execRaw(ctx context.Context, query string, vars map[string]any) (map[string]json.RawMessage, error) {
	raw, err := mustGraphQLClient().ExecRaw(ctx, query, vars)
	if err != nil {
		return nil, err
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return data, nil
}

// gqlObject runs query and prints the single top-level field as an object,
// erroring when it is null (e.g. a missing resource — the GraphQL equivalent of
// the old REST 404).
func gqlObject(ctx context.Context, query string, vars map[string]any, field string) error {
	data, err := execRaw(ctx, query, vars)
	if err != nil {
		return err
	}
	body := data[field]
	if len(body) == 0 || string(body) == "null" {
		return fmt.Errorf("not found")
	}
	return printJSON(body)
}

// gqlConnection runs query and reshapes the {nodes, pageInfo} connection at
// `field` into the REST-style {<listKey>: [...], "pagination": {...}} envelope
// the output renderer (and existing scripts) expect.
func gqlConnection(ctx context.Context, query string, vars map[string]any, field, listKey string) error {
	data, err := execRaw(ctx, query, vars)
	if err != nil {
		return err
	}
	var conn struct {
		Nodes    json.RawMessage `json:"nodes"`
		PageInfo json.RawMessage `json:"pageInfo"`
	}
	if err := json.Unmarshal(data[field], &conn); err != nil {
		return fmt.Errorf("decoding %s: %w", field, err)
	}
	out, err := json.Marshal(map[string]json.RawMessage{
		listKey:      conn.Nodes,
		"pagination": conn.PageInfo,
	})
	if err != nil {
		return err
	}
	return printJSON(out)
}

// gqlList runs query and prints the list at `field` as {<listKey>: [...]}.
func gqlList(ctx context.Context, query string, vars map[string]any, field, listKey string) error {
	data, err := execRaw(ctx, query, vars)
	if err != nil {
		return err
	}
	out, err := json.Marshal(map[string]json.RawMessage{listKey: data[field]})
	if err != nil {
		return err
	}
	return printJSON(out)
}

// gqlSuccess runs a mutation whose result is a boolean and prints
// {"success": <bool>}.
func gqlSuccess(ctx context.Context, query string, vars map[string]any, field string) error {
	data, err := execRaw(ctx, query, vars)
	if err != nil {
		return err
	}
	var ok bool
	_ = json.Unmarshal(data[field], &ok)
	out, _ := json.Marshal(map[string]bool{"success": ok})
	return printJSON(out)
}
