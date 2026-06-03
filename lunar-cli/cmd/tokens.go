package cmd

import "github.com/spf13/cobra"

// tokenFields is the GraphQL selection for an API token, aliased to the
// snake_case shape the CLI output (and existing scripts) expect.
const tokenFields = `
	id
	name
	created_at: createdAt
	last_used: lastUsed
	revoked
`

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "API token management for connected clients",
}

func init() {
	rootCmd.AddCommand(tokensCmd)
}

// ─── list ────────────────────────────────────────────────

var listTokensCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API tokens",
	RunE:  runListTokens,
}

func init() {
	tokensCmd.AddCommand(listTokensCmd)
}

func runListTokens(cmd *cobra.Command, args []string) error {
	query := `query { apiTokens {` + tokenFields + `} }`
	return gqlList(cmd.Context(), query, nil, "apiTokens", "tokens")
}

// ─── revoke ────────────────────────────────────────────────

var revokeTokenCmd = &cobra.Command{
	Use:   "revoke <id>",
	Short: "Revoke an API token",
	Args:  cobra.ExactArgs(1),
	RunE:  runRevokeToken,
}

func init() {
	tokensCmd.AddCommand(revokeTokenCmd)
}

func runRevokeToken(cmd *cobra.Command, args []string) error {
	query := `mutation ($id: ID!) { revokeApiToken(id: $id) }`
	vars := map[string]any{"id": args[0]}
	return gqlSuccess(cmd.Context(), query, vars, "revokeApiToken")
}
