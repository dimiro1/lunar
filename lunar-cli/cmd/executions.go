package cmd

import "github.com/spf13/cobra"

// executionSummaryFields drops the large event/response JSON for the list view.
const executionSummaryFields = `
	id
	function_id: functionId
	function_version_id: functionVersionId
	status
	duration_ms: durationMs
	error_message: errorMessage
	trigger
	created_at: createdAt
`

// executionFields is the full selection for a single execution.
const executionFields = `
	id
	function_id: functionId
	function_version_id: functionVersionId
	status
	duration_ms: durationMs
	error_message: errorMessage
	event_json: eventJson
	response_json: responseJson
	trigger
	created_at: createdAt
`

const aiRequestFields = `
	id
	execution_id: executionId
	provider
	model
	endpoint
	request_json: requestJson
	response_json: responseJson
	status
	error_message: errorMessage
	input_tokens: inputTokens
	output_tokens: outputTokens
	duration_ms: durationMs
	created_at: createdAt
`

const emailRequestFields = `
	id
	execution_id: executionId
	from
	to
	subject
	has_text: hasText
	has_html: hasHtml
	request_json: requestJson
	response_json: responseJson
	status
	error_message: errorMessage
	email_id: emailId
	duration_ms: durationMs
	created_at: createdAt
`

var executionsCmd = &cobra.Command{
	Use:   "executions",
	Short: "Function execution history and logs",
}

func init() {
	rootCmd.AddCommand(executionsCmd)
}

// pageFlags registers --limit/--offset on a command and returns pointers to them.
func pageFlags(c *cobra.Command) (*int, *int) {
	limit := new(int)
	offset := new(int)
	c.Flags().IntVar(limit, "limit", 20, "Maximum number of items to return (default 20, max 100)")
	c.Flags().IntVar(offset, "offset", 0, "Number of items to skip")
	return limit, offset
}

// ─── list ────────────────────────────────────────────────

var listExecutionsCmd = &cobra.Command{
	Use:   "list <id>",
	Short: "List executions of a function",
	Args:  cobra.ExactArgs(1),
	RunE:  runListExecutions,
}

var listExecutionsLimit, listExecutionsOffset *int

func init() {
	executionsCmd.AddCommand(listExecutionsCmd)
	listExecutionsLimit, listExecutionsOffset = pageFlags(listExecutionsCmd)
}

func runListExecutions(cmd *cobra.Command, args []string) error {
	query := `query ($id: ID!, $limit: Int!, $offset: Int!) {
		executions(functionId: $id, limit: $limit, offset: $offset) {
			nodes {` + executionSummaryFields + `}
			pageInfo { total limit offset }
		}
	}`
	vars := map[string]any{"id": args[0], "limit": *listExecutionsLimit, "offset": *listExecutionsOffset}
	return gqlConnection(cmd.Context(), query, vars, "executions", "executions")
}

// ─── get ────────────────────────────────────────────────

var getExecutionCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get execution details",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetExecution,
}

func init() {
	executionsCmd.AddCommand(getExecutionCmd)
}

func runGetExecution(cmd *cobra.Command, args []string) error {
	query := `query ($id: ID!) { execution(id: $id) {` + executionFields + `} }`
	return gqlObject(cmd.Context(), query, map[string]any{"id": args[0]}, "execution")
}

// ─── logs ────────────────────────────────────────────────

var getExecutionLogsCmd = &cobra.Command{
	Use:   "logs <id>",
	Short: "Get execution logs",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetExecutionLogs,
}

var logsLimit, logsOffset *int

func init() {
	executionsCmd.AddCommand(getExecutionLogsCmd)
	logsLimit, logsOffset = pageFlags(getExecutionLogsCmd)
}

func runGetExecutionLogs(cmd *cobra.Command, args []string) error {
	query := `query ($id: ID!, $limit: Int!, $offset: Int!) {
		executionLogs(executionId: $id, limit: $limit, offset: $offset) {
			nodes { level message created_at: createdAt }
			pageInfo { total limit offset }
		}
	}`
	vars := map[string]any{"id": args[0], "limit": *logsLimit, "offset": *logsOffset}
	return gqlConnection(cmd.Context(), query, vars, "executionLogs", "logs")
}

// ─── ai-requests ────────────────────────────────────────────────

var getExecutionAIRequestsCmd = &cobra.Command{
	Use:   "ai-requests <id>",
	Short: "Get AI requests for an execution",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetExecutionAIRequests,
}

var aiRequestsLimit, aiRequestsOffset *int

func init() {
	executionsCmd.AddCommand(getExecutionAIRequestsCmd)
	aiRequestsLimit, aiRequestsOffset = pageFlags(getExecutionAIRequestsCmd)
}

func runGetExecutionAIRequests(cmd *cobra.Command, args []string) error {
	query := `query ($id: ID!, $limit: Int!, $offset: Int!) {
		executionAiRequests(executionId: $id, limit: $limit, offset: $offset) {
			nodes {` + aiRequestFields + `}
			pageInfo { total limit offset }
		}
	}`
	vars := map[string]any{"id": args[0], "limit": *aiRequestsLimit, "offset": *aiRequestsOffset}
	return gqlConnection(cmd.Context(), query, vars, "executionAiRequests", "ai_requests")
}

// ─── email-requests ────────────────────────────────────────────────

var getExecutionEmailRequestsCmd = &cobra.Command{
	Use:   "email-requests <id>",
	Short: "Get email requests for an execution",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetExecutionEmailRequests,
}

var emailRequestsLimit, emailRequestsOffset *int

func init() {
	executionsCmd.AddCommand(getExecutionEmailRequestsCmd)
	emailRequestsLimit, emailRequestsOffset = pageFlags(getExecutionEmailRequestsCmd)
}

func runGetExecutionEmailRequests(cmd *cobra.Command, args []string) error {
	query := `query ($id: ID!, $limit: Int!, $offset: Int!) {
		executionEmailRequests(executionId: $id, limit: $limit, offset: $offset) {
			nodes {` + emailRequestFields + `}
			pageInfo { total limit offset }
		}
	}`
	vars := map[string]any{"id": args[0], "limit": *emailRequestsLimit, "offset": *emailRequestsOffset}
	return gqlConnection(cmd.Context(), query, vars, "executionEmailRequests", "email_requests")
}
