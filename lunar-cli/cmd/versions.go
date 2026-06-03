package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// versionFields is the full GraphQL selection for a version (includes code).
const versionFields = `
	id
	function_id: functionId
	version
	code
	created_at: createdAt
	created_by: createdBy
	is_active: isActive
`

// versionSummaryFields drops the code for the list view.
const versionSummaryFields = `
	id
	function_id: functionId
	version
	created_at: createdAt
	created_by: createdBy
	is_active: isActive
`

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Function version management",
}

func init() {
	rootCmd.AddCommand(versionsCmd)
}

// ─── list ────────────────────────────────────────────────

var listVersionsCmd = &cobra.Command{
	Use:   "list <id>",
	Short: "List all versions of a function",
	Args:  cobra.ExactArgs(1),
	RunE:  runListVersions,
}

var (
	listVersionsCmdLimit  int
	listVersionsCmdOffset int
)

func init() {
	versionsCmd.AddCommand(listVersionsCmd)
	listVersionsCmd.Flags().IntVar(&listVersionsCmdLimit, "limit", 20, "Maximum number of items to return (default 20, max 100)")
	listVersionsCmd.Flags().IntVar(&listVersionsCmdOffset, "offset", 0, "Number of items to skip")
}

func runListVersions(cmd *cobra.Command, args []string) error {
	query := `query ($id: ID!, $limit: Int!, $offset: Int!) {
		versions(functionId: $id, limit: $limit, offset: $offset) {
			nodes {` + versionSummaryFields + `}
			pageInfo { total limit offset }
		}
	}`
	vars := map[string]any{"id": args[0], "limit": listVersionsCmdLimit, "offset": listVersionsCmdOffset}
	return gqlConnection(cmd.Context(), query, vars, "versions", "versions")
}

// ─── get ────────────────────────────────────────────────

var getVersionCmd = &cobra.Command{
	Use:   "get <id> <version>",
	Short: "Get a specific version",
	Args:  cobra.ExactArgs(2),
	RunE:  runGetVersion,
}

func init() {
	versionsCmd.AddCommand(getVersionCmd)
}

func runGetVersion(cmd *cobra.Command, args []string) error {
	version, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}
	query := `query ($id: ID!, $version: Int!) {
		version(functionId: $id, version: $version) {` + versionFields + `}
	}`
	vars := map[string]any{"id": args[0], "version": version}
	return gqlObject(cmd.Context(), query, vars, "version")
}

// ─── activate ────────────────────────────────────────────────

var activateVersionCmd = &cobra.Command{
	Use:   "activate <id> <version-id>",
	Short: "Activate a version",
	Args:  cobra.ExactArgs(2),
	RunE:  runActivateVersion,
}

func init() {
	versionsCmd.AddCommand(activateVersionCmd)
}

func runActivateVersion(cmd *cobra.Command, args []string) error {
	query := `mutation ($id: ID!, $versionId: ID!) {
		activateVersion(functionId: $id, versionId: $versionId) {` + functionFields + `}
	}`
	vars := map[string]any{"id": args[0], "versionId": args[1]}
	return gqlObject(cmd.Context(), query, vars, "activateVersion")
}

// ─── delete ────────────────────────────────────────────────

var deleteVersionCmd = &cobra.Command{
	Use:   "delete <id> <version-id>",
	Short: "Delete a version",
	Args:  cobra.ExactArgs(2),
	RunE:  runDeleteVersion,
}

func init() {
	versionsCmd.AddCommand(deleteVersionCmd)
}

func runDeleteVersion(cmd *cobra.Command, args []string) error {
	query := `mutation ($id: ID!, $versionId: ID!) {
		deleteVersion(functionId: $id, versionId: $versionId)
	}`
	vars := map[string]any{"id": args[0], "versionId": args[1]}
	return gqlSuccess(cmd.Context(), query, vars, "deleteVersion")
}

// ─── diff ────────────────────────────────────────────────

var getVersionDiffCmd = &cobra.Command{
	Use:   "diff <id> <v1> <v2>",
	Short: "Get diff between two versions",
	Args:  cobra.ExactArgs(3),
	RunE:  runGetVersionDiff,
}

func init() {
	versionsCmd.AddCommand(getVersionDiffCmd)
}

func runGetVersionDiff(cmd *cobra.Command, args []string) error {
	v1, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid v1: %w", err)
	}
	v2, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid v2: %w", err)
	}
	query := `query ($id: ID!, $v1: Int!, $v2: Int!) {
		versionDiff(functionId: $id, oldVersion: $v1, newVersion: $v2) {
			old_version: oldVersion
			new_version: newVersion
			diff: lines { line_type: lineType old_line: oldLine new_line: newLine content }
		}
	}`
	vars := map[string]any{"id": args[0], "v1": v1, "v2": v2}
	return gqlObject(cmd.Context(), query, vars, "versionDiff")
}
