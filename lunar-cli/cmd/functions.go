package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// functionFields is the full GraphQL selection for a function, aliased to the
// snake_case shape the CLI output (and existing scripts) expect.
const functionFields = `
	id
	name
	description
	disabled
	retention_days: retentionDays
	cron_schedule: cronSchedule
	cron_status: cronStatus
	save_response: saveResponse
	created_at: createdAt
	updated_at: updatedAt
	active_version: activeVersion {
		id
		function_id: functionId
		version
		code
		created_at: createdAt
		created_by: createdBy
		is_active: isActive
	}
	env_vars: envVars
	scoped_data: scopedData
	global_data: globalData
`

// functionSummaryFields is the trimmed selection used for the list view: no
// code or env/KV maps, so listing many functions stays cheap.
const functionSummaryFields = `
	id
	name
	description
	disabled
	cron_schedule: cronSchedule
	cron_status: cronStatus
	save_response: saveResponse
	created_at: createdAt
	updated_at: updatedAt
	active_version: activeVersion { version }
`

var functionsCmd = &cobra.Command{
	Use:   "functions",
	Short: "Function management operations",
}

func init() {
	rootCmd.AddCommand(functionsCmd)
}

// ─── list ────────────────────────────────────────────────

var listFunctionsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all functions",
	RunE:  runListFunctions,
}

var (
	listFunctionsCmdLimit  int
	listFunctionsCmdOffset int
)

func init() {
	functionsCmd.AddCommand(listFunctionsCmd)
	listFunctionsCmd.Flags().IntVar(&listFunctionsCmdLimit, "limit", 20, "Maximum number of items to return (default 20, max 100)")
	listFunctionsCmd.Flags().IntVar(&listFunctionsCmdOffset, "offset", 0, "Number of items to skip")
}

func runListFunctions(cmd *cobra.Command, args []string) error {
	query := `query ($limit: Int!, $offset: Int!) {
		functions(limit: $limit, offset: $offset) {
			nodes {` + functionSummaryFields + `}
			pageInfo { total limit offset }
		}
	}`
	vars := map[string]any{"limit": listFunctionsCmdLimit, "offset": listFunctionsCmdOffset}
	return gqlConnection(cmd.Context(), query, vars, "functions", "functions")
}

// ─── create ────────────────────────────────────────────────

var createFunctionCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new function",
	RunE:  runCreateFunction,
}

var (
	createFunctionCmdCode        string
	createFunctionCmdDescription string
	createFunctionCmdName        string
)

func init() {
	functionsCmd.AddCommand(createFunctionCmd)
	createFunctionCmd.Flags().StringVar(&createFunctionCmdCode, "code", "", "Lua code for the function (use \"-\" to read from stdin)")
	_ = createFunctionCmd.MarkFlagRequired("code")
	createFunctionCmd.Flags().StringVar(&createFunctionCmdDescription, "description", "", "Optional description")
	createFunctionCmd.Flags().StringVar(&createFunctionCmdName, "name", "", "Name for the function")
	_ = createFunctionCmd.MarkFlagRequired("name")
}

func runCreateFunction(cmd *cobra.Command, args []string) error {
	code, err := maybeStdin(createFunctionCmdCode)
	if err != nil {
		return err
	}
	input := map[string]any{"name": createFunctionCmdName, "code": code}
	if cmd.Flags().Changed("description") {
		input["description"] = createFunctionCmdDescription
	}
	query := `mutation ($input: CreateFunctionInput!) {
		createFunction(input: $input) {` + functionFields + `}
	}`
	return gqlObject(cmd.Context(), query, map[string]any{"input": input}, "createFunction")
}

// ─── get ────────────────────────────────────────────────

var getFunctionCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a specific function",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetFunction,
}

func init() {
	functionsCmd.AddCommand(getFunctionCmd)
}

func runGetFunction(cmd *cobra.Command, args []string) error {
	query := `query ($id: ID!) { function(id: $id) {` + functionFields + `} }`
	return gqlObject(cmd.Context(), query, map[string]any{"id": args[0]}, "function")
}

// ─── update ────────────────────────────────────────────────

var updateFunctionCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a function",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdateFunction,
}

var (
	updateFunctionCmdCode          string
	updateFunctionCmdCronSchedule  string
	updateFunctionCmdCronStatus    string
	updateFunctionCmdDescription   string
	updateFunctionCmdDisabled      bool
	updateFunctionCmdName          string
	updateFunctionCmdRetentionDays int
	updateFunctionCmdSaveResponse  bool
)

func init() {
	functionsCmd.AddCommand(updateFunctionCmd)
	updateFunctionCmd.Flags().StringVar(&updateFunctionCmdCode, "code", "", "New code (creates a new version; use \"-\" to read from stdin)")
	updateFunctionCmd.Flags().StringVar(&updateFunctionCmdCronSchedule, "cron-schedule", "", "Cron expression (5-field). Empty string clears the schedule.")
	updateFunctionCmd.Flags().StringVar(&updateFunctionCmdCronStatus, "cron-status", "", "Cron schedule status (\"active\" or \"paused\")")
	updateFunctionCmd.Flags().StringVar(&updateFunctionCmdDescription, "description", "", "New description")
	updateFunctionCmd.Flags().BoolVar(&updateFunctionCmdDisabled, "disabled", false, "Set true to disable the function, false to enable it")
	updateFunctionCmd.Flags().StringVar(&updateFunctionCmdName, "name", "", "New name for the function")
	updateFunctionCmd.Flags().IntVar(&updateFunctionCmdRetentionDays, "retention-days", 0, "Number of days to retain execution history")
	updateFunctionCmd.Flags().BoolVar(&updateFunctionCmdSaveResponse, "save-response", false, "Whether to save HTTP responses with executions")
}

func runUpdateFunction(cmd *cobra.Command, args []string) error {
	input := map[string]any{}
	if cmd.Flags().Changed("code") {
		code, err := maybeStdin(updateFunctionCmdCode)
		if err != nil {
			return err
		}
		input["code"] = code
	}
	if cmd.Flags().Changed("cron-schedule") {
		input["cronSchedule"] = updateFunctionCmdCronSchedule
	}
	if cmd.Flags().Changed("cron-status") {
		input["cronStatus"] = updateFunctionCmdCronStatus
	}
	if cmd.Flags().Changed("description") {
		input["description"] = updateFunctionCmdDescription
	}
	if cmd.Flags().Changed("disabled") {
		input["disabled"] = updateFunctionCmdDisabled
	}
	if cmd.Flags().Changed("name") {
		input["name"] = updateFunctionCmdName
	}
	if cmd.Flags().Changed("retention-days") {
		input["retentionDays"] = updateFunctionCmdRetentionDays
	}
	if cmd.Flags().Changed("save-response") {
		input["saveResponse"] = updateFunctionCmdSaveResponse
	}
	query := `mutation ($id: ID!, $input: UpdateFunctionInput!) {
		updateFunction(id: $id, input: $input) {` + functionFields + `}
	}`
	vars := map[string]any{"id": args[0], "input": input}
	return gqlObject(cmd.Context(), query, vars, "updateFunction")
}

// ─── delete ────────────────────────────────────────────────

var deleteFunctionCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a function",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeleteFunction,
}

func init() {
	functionsCmd.AddCommand(deleteFunctionCmd)
}

func runDeleteFunction(cmd *cobra.Command, args []string) error {
	query := `mutation ($id: ID!) { deleteFunction(id: $id) }`
	return gqlSuccess(cmd.Context(), query, map[string]any{"id": args[0]}, "deleteFunction")
}

// ─── env ────────────────────────────────────────────────

var updateEnvVarsCmd = &cobra.Command{
	Use:   "env <id>",
	Short: "Update environment variables",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdateEnvVars,
}

var updateEnvVarsCmdEnvVars []string

func init() {
	functionsCmd.AddCommand(updateEnvVarsCmd)
	updateEnvVarsCmd.Flags().StringArrayVar(&updateEnvVarsCmdEnvVars, "env", nil, "Environment variables to set as KEY=VALUE (repeatable). Replaces the full set.")
	_ = updateEnvVarsCmd.MarkFlagRequired("env")
}

func runUpdateEnvVars(cmd *cobra.Command, args []string) error {
	envMap, err := parseKeyValues(updateEnvVarsCmdEnvVars, "--env")
	if err != nil {
		return err
	}
	query := `mutation ($id: ID!, $env: Map!) {
		setFunctionEnv(id: $id, env: $env) {` + functionFields + `}
	}`
	vars := map[string]any{"id": args[0], "env": envMap}
	return gqlObject(cmd.Context(), query, vars, "setFunctionEnv")
}

// ─── kv ────────────────────────────────────────────────

var updateKVCmd = &cobra.Command{
	Use:   "kv <id>",
	Short: "Update key-value store",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdateKV,
}

var (
	updateKVCmdGlobal bool
	updateKVCmdKv     []string
)

func init() {
	functionsCmd.AddCommand(updateKVCmd)
	updateKVCmd.Flags().BoolVar(&updateKVCmdGlobal, "global", false, "Write to the global KV store instead of the function scope")
	_ = updateKVCmd.MarkFlagRequired("global")
	updateKVCmd.Flags().StringArrayVar(&updateKVCmdKv, "kv", nil, "KV pairs to set as KEY=VALUE (repeatable). Replaces the full set.")
	_ = updateKVCmd.MarkFlagRequired("kv")
}

func runUpdateKV(cmd *cobra.Command, args []string) error {
	kvMap, err := parseKeyValues(updateKVCmdKv, "--kv")
	if err != nil {
		return err
	}
	query := `mutation ($id: ID!, $kv: Map!, $global: Boolean!) {
		setFunctionKv(id: $id, kv: $kv, global: $global) {` + functionFields + `}
	}`
	vars := map[string]any{"id": args[0], "kv": kvMap, "global": updateKVCmdGlobal}
	return gqlObject(cmd.Context(), query, vars, "setFunctionKv")
}

// ─── next-run ────────────────────────────────────────────────

var getNextRunCmd = &cobra.Command{
	Use:   "next-run <id>",
	Short: "Get next scheduled run time",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetNextRun,
}

func init() {
	functionsCmd.AddCommand(getNextRunCmd)
}

func runGetNextRun(cmd *cobra.Command, args []string) error {
	query := `query ($id: ID!) {
		nextRun(functionId: $id) {
			has_schedule: hasSchedule
			cron_schedule: cronSchedule
			cron_status: cronStatus
			is_paused: isPaused
			next_run: nextRun
			next_run_human: nextRunHuman
		}
	}`
	return gqlObject(cmd.Context(), query, map[string]any{"id": args[0]}, "nextRun")
}

// maybeStdin returns the contents of stdin when value is "-", otherwise value.
func maybeStdin(value string) (string, error) {
	if value != "-" {
		return value, nil
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	return string(b), nil
}

// parseKeyValues parses repeated KEY=VALUE flag values into a map.
func parseKeyValues(pairs []string, flag string) (map[string]string, error) {
	out := make(map[string]string, len(pairs))
	for _, kv := range pairs {
		key, value, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("invalid %s format %q, expected KEY=VALUE", flag, kv)
		}
		out[key] = value
	}
	return out, nil
}
