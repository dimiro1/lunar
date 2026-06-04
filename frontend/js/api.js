/**
 * @fileoverview API client for the Lunar Dashboard.
 * Provides methods for authentication, function management, and execution.
 */

/**
 * @typedef {import('./types.js').LunarFunction} LunarFunction
 * @typedef {import('./types.js').FunctionsListResponse} FunctionsListResponse
 * @typedef {import('./types.js').FunctionVersion} FunctionVersion
 * @typedef {import('./types.js').VersionsListResponse} VersionsListResponse
 * @typedef {import('./types.js').Execution} Execution
 * @typedef {import('./types.js').ExecutionsListResponse} ExecutionsListResponse
 * @typedef {import('./types.js').ExecutionLogsResponse} ExecutionLogsResponse
 * @typedef {import('./types.js').DiffResponse} DiffResponse
 * @typedef {import('./types.js').ExecuteRequest} ExecuteRequest
 * @typedef {import('./types.js').ExecuteResponse} ExecuteResponse
 */

/**
 * Makes an API request with credentials and error handling.
 * @param {Object} config - Mithril request config
 * @param {string} config.method - HTTP method
 * @param {string} config.url - Request URL
 * @param {Object} [config.body] - Request body
 * @param {Object} [config.headers] - Request headers
 * @returns {Promise<*>} Response data
 * @throws {Error} API error with message from response
 */
const apiRequest = async (config) => {
  try {
    return await m.request({
      ...config,
      // Include cookies in requests
      credentials: "same-origin",
    });
  } catch (err) {
    // Mithril parses JSON responses automatically
    // On error, err.response contains the parsed JSON body
    if (err.response && err.response.error) {
      // Throw a proper Error object with the error message
      const error = new Error(err.response.error);
      error.code = err.code;
      throw error;
    }
    throw err;
  }
};

// Global error handler for auth failures
const originalRequest = m.request;
m.request = function (options) {
  return originalRequest(options).catch((error) => {
    // If we get a 401, redirect to login
    if (error.code === 401) {
      m.route.set("/login");
    }
    throw error;
  });
};

/**
 * Executes a GraphQL operation against /graphql.
 *
 * GraphQL returns HTTP 200 with an `errors` array for resolver/validation
 * failures, so those are surfaced here as a thrown Error (message from the
 * first error). HTTP-level failures (e.g. an unauthenticated 401 from the auth
 * middleware that fronts /graphql) reject through the global m.request handler
 * above, which redirects to /login — exactly as the REST calls did.
 *
 * @param {string} query - GraphQL query or mutation document
 * @param {Object} [variables] - GraphQL variables
 * @returns {Promise<*>} The `data` payload of the response
 * @throws {Error} When the response contains GraphQL errors
 */
const gqlRequest = async (query, variables = {}) => {
  const res = await m.request({
    method: "POST",
    url: "/graphql",
    body: { query, variables },
    credentials: "same-origin",
  });
  if (res && res.errors && res.errors.length > 0) {
    const error = new Error(res.errors[0].message);
    error.graphqlErrors = res.errors;
    throw error;
  }
  return res ? res.data : null;
};

/**
 * Maps a GraphQL FunctionVersion to the snake_case shape the views consume.
 * Fields a query did not select are simply left undefined.
 * @param {Object} v - GraphQL FunctionVersion
 * @returns {Object} snake_case version
 */
const mapVersion = (v) =>
  v == null ? v : {
    id: v.id,
    function_id: v.functionId,
    version: v.version,
    code: v.code,
    language: v.language,
    created_at: v.createdAt,
    created_by: v.createdBy,
    is_active: v.isActive,
  };

/**
 * Maps a GraphQL Function to the snake_case shape the views consume. Works for
 * both trimmed (list) and full (detail) selections; absent fields stay undefined.
 * @param {Object} f - GraphQL Function
 * @returns {Object} snake_case function
 */
const mapFunction = (f) =>
  f == null ? f : {
    id: f.id,
    name: f.name,
    description: f.description,
    disabled: f.disabled,
    retention_days: f.retentionDays,
    cron_schedule: f.cronSchedule,
    cron_status: f.cronStatus,
    save_response: f.saveResponse,
    created_at: f.createdAt,
    updated_at: f.updatedAt,
    active_version: f.activeVersion
      ? mapVersion(f.activeVersion)
      : f.activeVersion,
    env_vars: f.envVars,
    scoped_data: f.scopedData,
    global_data: f.globalData,
  };

/** Maps a GraphQL Execution to the snake_case shape the views consume. */
const mapExecution = (e) =>
  e == null ? e : {
    id: e.id,
    function_id: e.functionId,
    function_version_id: e.functionVersionId,
    status: e.status,
    duration_ms: e.durationMs,
    error_message: e.errorMessage,
    event_json: e.eventJson,
    response_json: e.responseJson,
    trigger: e.trigger,
    created_at: e.createdAt,
  };

/** Maps a GraphQL AIRequest to the snake_case shape the views consume. */
const mapAIRequest = (a) =>
  a == null ? a : {
    id: a.id,
    execution_id: a.executionId,
    provider: a.provider,
    model: a.model,
    endpoint: a.endpoint,
    request_json: a.requestJson,
    response_json: a.responseJson,
    status: a.status,
    error_message: a.errorMessage,
    input_tokens: a.inputTokens,
    output_tokens: a.outputTokens,
    duration_ms: a.durationMs,
    created_at: a.createdAt,
  };

/** Maps a GraphQL EmailRequest to the snake_case shape the views consume. */
const mapEmailRequest = (em) =>
  em == null ? em : {
    id: em.id,
    execution_id: em.executionId,
    from: em.from,
    to: em.to,
    subject: em.subject,
    has_text: em.hasText,
    has_html: em.hasHtml,
    request_json: em.requestJson,
    response_json: em.responseJson,
    status: em.status,
    error_message: em.errorMessage,
    email_id: em.emailId,
    duration_ms: em.durationMs,
    created_at: em.createdAt,
  };

/** Maps a GraphQL LogEntry to the snake_case shape the views consume. */
const mapLog = (l) =>
  l == null
    ? l
    : { level: l.level, message: l.message, created_at: l.createdAt };

/** Maps a GraphQL APIToken to the snake_case shape the views consume. */
const mapToken = (tk) =>
  tk == null ? tk : {
    id: tk.id,
    name: tk.name,
    created_at: tk.createdAt,
    last_used: tk.lastUsed,
    revoked: tk.revoked,
  };

/** Maps a GraphQL NextRun to the snake_case shape the views consume. */
const mapNextRun = (n) =>
  n == null ? n : {
    has_schedule: n.hasSchedule,
    cron_schedule: n.cronSchedule,
    cron_status: n.cronStatus,
    is_paused: n.isPaused,
    next_run: n.nextRun,
    next_run_human: n.nextRunHuman,
  };

/** Maps a GraphQL VersionDiff to the snake_case shape the diff view consumes. */
const mapDiff = (d) =>
  d == null ? d : {
    old_version: d.oldVersion,
    new_version: d.newVersion,
    diff: (d.lines || []).map((l) => ({
      line_type: l.lineType,
      content: l.content,
      old_line: l.oldLine,
      new_line: l.newLine,
    })),
  };

/**
 * Builds a GraphQL UpdateFunctionInput from the snake_case update object the
 * views pass, including only the fields actually present.
 * @param {Object} data - snake_case update fields
 * @returns {Object} camelCase UpdateFunctionInput
 */
const toUpdateFunctionInput = (data) => {
  const input = {};
  if (data.name !== undefined) input.name = data.name;
  if (data.description !== undefined) input.description = data.description;
  if (data.code !== undefined) input.code = data.code;
  if (data.disabled !== undefined) input.disabled = data.disabled;
  if (data.retention_days !== undefined) {
    input.retentionDays = data.retention_days;
  }
  if (data.cron_schedule !== undefined) input.cronSchedule = data.cron_schedule;
  if (data.cron_status !== undefined) input.cronStatus = data.cron_status;
  if (data.save_response !== undefined) input.saveResponse = data.save_response;
  return input;
};

// Shared GraphQL field selections. Defined once so the standalone queries and
// the combined (multi-resource) queries below always request the same shape —
// avoiding drift between, say, `functions.get` and `functions.getWithNextRun`.
const PAGE_INFO = `total limit offset`;
const VERSION_FULL_FIELDS =
  `id functionId version code language createdAt createdBy isActive`;
const VERSION_SUMMARY_FIELDS =
  `id functionId version language createdAt createdBy isActive`;
// Full function detail (settings/detail views): includes active version code and
// the env/KV maps.
const FUNCTION_DETAIL_FIELDS = `
  id name description disabled
  retentionDays cronSchedule cronStatus saveResponse createdAt updatedAt
  activeVersion { ${VERSION_FULL_FIELDS} }
  envVars scopedData globalData
`;
// Lightweight function selection for list rows and detail-page headers: no code,
// env, or KV — just enough to render a name/status/active version number.
const FUNCTION_SUMMARY_FIELDS =
  `id name description disabled activeVersion { version language }`;
const EXECUTION_FULL_FIELDS = `
  id functionId functionVersionId status durationMs
  errorMessage eventJson responseJson trigger createdAt
`;
const EXECUTION_SUMMARY_FIELDS = `
  id functionId functionVersionId status durationMs
  errorMessage trigger createdAt
`;
const LOG_FIELDS = `level message createdAt`;
const AI_REQUEST_FIELDS = `
  id executionId provider model endpoint requestJson responseJson
  status errorMessage inputTokens outputTokens durationMs createdAt
`;
const EMAIL_REQUEST_FIELDS = `
  id executionId from to subject hasText hasHtml requestJson
  responseJson status errorMessage emailId durationMs createdAt
`;
const NEXT_RUN_FIELDS =
  `hasSchedule cronSchedule cronStatus isPaused nextRun nextRunHuman`;
const METRICS_FIELDS = `
  summary { count errorCount errorRate avgDurationMs maxDurationMs }
  buckets { bucketStart count errorCount avgDurationMs maxDurationMs }
  granularity
`;
const TOKEN_FIELDS = `id name createdAt lastUsed revoked`;
const DIFF_FIELDS =
  `oldVersion newVersion lines { lineType oldLine newLine content }`;

/**
 * API client for the lunar Dashboard.
 * @namespace
 */
export const API = {
  /**
   * Authentication methods.
   * @namespace
   */
  auth: {
    /**
     * Authenticates with an API key.
     * @param {string} apiKey - The API key to authenticate with
     * @returns {Promise<{success: boolean}>} Success response
     * @throws {Error} Authentication error
     */
    login: (apiKey) =>
      // Use originalRequest to avoid the global 401 redirect
      originalRequest({
        method: "POST",
        url: "/api/auth/login",
        body: { apiKey },
        credentials: "same-origin",
      }).catch((err) => {
        // Mithril parses JSON responses automatically
        // On error, err.response contains the parsed JSON body
        if (err.response && err.response.error) {
          const error = new Error(err.response.error);
          error.error = err.response.error;
          throw error;
        }
        throw err;
      }),

    /**
     * Logs out the current session.
     * @returns {Promise<void>}
     */
    logout: () =>
      apiRequest({
        method: "POST",
        url: "/api/auth/logout",
      }),

    getDeviceApproveStatus: (code) =>
      apiRequest({
        method: "GET",
        url: `/api/auth/device-approve?code=${code}`,
      }),

    approveDevice: (deviceCode, action) =>
      apiRequest({
        method: "POST",
        url: "/api/auth/device-approve",
        body: { device_code: deviceCode, action },
      }),
  },

  tokens: {
    list: () =>
      gqlRequest(
        `query { apiTokens { ${TOKEN_FIELDS} } }`,
      ).then((data) => ({ tokens: data.apiTokens.map(mapToken) })),

    revoke: (id) =>
      gqlRequest(`mutation ($id: ID!) { revokeApiToken(id: $id) }`, { id }),
  },

  /**
   * Function management methods.
   * @namespace
   */
  functions: {
    /**
     * Lists all functions with pagination.
     * @param {number} [limit=20] - Maximum number of functions to return
     * @param {number} [offset=0] - Number of functions to skip
     * @returns {Promise<FunctionsListResponse>} Paginated list of functions
     */
    list: (limit = 20, offset = 0) =>
      gqlRequest(
        `query ($limit: Int!, $offset: Int!) {
          functions(limit: $limit, offset: $offset) {
            nodes { ${FUNCTION_SUMMARY_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        { limit, offset },
      ).then((data) => ({
        functions: data.functions.nodes.map(mapFunction),
        pagination: data.functions.pageInfo,
      })),

    /**
     * Gets a single function by ID.
     * @param {string} id - Function ID
     * @returns {Promise<LunarFunction>} The function
     */
    get: (id) =>
      gqlRequest(
        `query ($id: ID!) {
          function(id: $id) { ${FUNCTION_DETAIL_FIELDS} }
        }`,
        { id },
      ).then((data) => mapFunction(data.function)),

    /**
     * Gets a function together with its next scheduled run in one round-trip.
     * Used by the settings view, which needs both on load.
     * @param {string} id - Function ID
     * @returns {Promise<{func: LunarFunction, nextRun: Object}>}
     */
    getWithNextRun: (id) =>
      gqlRequest(
        `query ($id: ID!) {
          function(id: $id) { ${FUNCTION_DETAIL_FIELDS} }
          nextRun(functionId: $id) { ${NEXT_RUN_FIELDS} }
        }`,
        { id },
      ).then((data) => ({
        func: mapFunction(data.function),
        nextRun: mapNextRun(data.nextRun),
      })),

    /**
     * Gets a function (header fields) together with a page of its versions in
     * one round-trip. Used by the versions view on load.
     * @param {string} id - Function ID
     * @param {number} [limit=20]
     * @param {number} [offset=0]
     * @returns {Promise<{func: LunarFunction, versions: FunctionVersion[], pagination: Object}>}
     */
    getWithVersions: (id, limit = 20, offset = 0) =>
      gqlRequest(
        `query ($id: ID!, $limit: Int!, $offset: Int!) {
          function(id: $id) { ${FUNCTION_SUMMARY_FIELDS} }
          versions(functionId: $id, limit: $limit, offset: $offset) {
            nodes { ${VERSION_SUMMARY_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        { id, limit, offset },
      ).then((data) => ({
        func: mapFunction(data.function),
        versions: data.versions.nodes.map(mapVersion),
        pagination: data.versions.pageInfo,
      })),

    /**
     * Gets a function (header fields) together with a page of its executions in
     * one round-trip. Used by the executions view on load.
     * @param {string} id - Function ID
     * @param {number} [limit=20]
     * @param {number} [offset=0]
     * @returns {Promise<{func: LunarFunction, executions: Execution[], pagination: Object}>}
     */
    getWithExecutions: (id, limit = 20, offset = 0) =>
      gqlRequest(
        `query ($id: ID!, $limit: Int!, $offset: Int!) {
          function(id: $id) { ${FUNCTION_SUMMARY_FIELDS} }
          executions(functionId: $id, limit: $limit, offset: $offset) {
            nodes { ${EXECUTION_SUMMARY_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        { id, limit, offset },
      ).then((data) => ({
        func: mapFunction(data.function),
        executions: data.executions.nodes.map(mapExecution),
        pagination: data.executions.pageInfo,
      })),

    /**
     * Gets a function (header fields) together with its aggregated execution
     * metrics over the half-open window [from, to) (unix seconds) in one
     * round-trip. Used by the metrics view on load and on range changes.
     * @param {string} id - Function ID
     * @param {number} from - Window start (unix seconds, inclusive)
     * @param {number} to - Window end (unix seconds, exclusive)
     * @param {"hour"|"day"} granularity - Bucket granularity
     * @returns {Promise<{func: LunarFunction, metrics: Object}>}
     */
    getWithMetrics: (id, from, to, granularity) =>
      gqlRequest(
        `query ($id: ID!, $from: Int!, $to: Int!, $granularity: MetricGranularity!) {
          function(id: $id) {
            ${FUNCTION_SUMMARY_FIELDS}
            metrics(from: $from, to: $to, granularity: $granularity) {
              ${METRICS_FIELDS}
            }
          }
        }`,
        { id, from, to, granularity },
      ).then((data) => ({
        func: mapFunction(data.function),
        metrics: data.function ? data.function.metrics : null,
      })),

    /**
     * Creates a new function.
     * @param {Object} data - Function data
     * @param {string} data.name - Function name
     * @param {string} [data.description] - Function description
     * @param {string} data.code - Initial function code
     * @param {string} [data.language] - Function language ("lua" or "starlark")
     * @returns {Promise<LunarFunction>} The created function
     */
    create: (data) =>
      gqlRequest(
        `mutation ($input: CreateFunctionInput!) {
          createFunction(input: $input) { id }
        }`,
        {
          input: {
            name: data.name,
            description: data.description,
            code: data.code,
            language: data.language,
          },
        },
      ).then((d) => mapFunction(d.createFunction)),

    /**
     * Updates an existing function.
     * @param {string} id - Function ID
     * @param {Object} data - Update data
     * @param {string} [data.name] - New name
     * @param {string} [data.description] - New description
     * @param {string} [data.code] - New code (creates new version)
     * @param {boolean} [data.disabled] - Enable/disable function
     * @returns {Promise<LunarFunction>} The updated function
     */
    update: (id, data) =>
      gqlRequest(
        `mutation ($id: ID!, $input: UpdateFunctionInput!) {
          updateFunction(id: $id, input: $input) { id }
        }`,
        { id, input: toUpdateFunctionInput(data) },
      ).then((d) => mapFunction(d.updateFunction)),

    /**
     * Deletes a function.
     * @param {string} id - Function ID
     * @returns {Promise<void>}
     */
    delete: (id) =>
      gqlRequest(`mutation ($id: ID!) { deleteFunction(id: $id) }`, { id }),

    /**
     * Updates environment variables for a function.
     * @param {string} id - Function ID
     * @param {Object.<string, string>} env_vars - Environment variables
     * @returns {Promise<LunarFunction>} The updated function
     */
    updateEnv: (id, env_vars) =>
      gqlRequest(
        `mutation ($id: ID!, $env: Map!) {
          setFunctionEnv(id: $id, env: $env) { id }
        }`,
        { id, env: env_vars },
      ),

    /**
     * Updates kv store entries for a function.
     * @param {string} id - Function ID
     * @param {Object} updateData - Update data
     * @param {boolean} updateData.global - Whether the update is for global KV entries
     * @param {Array<{key: string, value: string}>} updateData.kvEntries - KV entries to update
     * @returns {Promise<LunarFunction>} The updated function
     */
    updateKvStore: (id, updateData) => {
      const kv = {};
      (updateData.kvEntries || []).forEach((entry) => {
        kv[entry.key] = entry.value;
      });
      return gqlRequest(
        `mutation ($id: ID!, $kv: Map!, $global: Boolean!) {
          setFunctionKv(id: $id, kv: $kv, global: $global) { id }
        }`,
        { id, kv, global: !!updateData.global },
      );
    },
  },

  /**
   * Version management methods.
   * @namespace
   */
  versions: {
    /**
     * Lists all versions for a function.
     * @param {string} functionId - Function ID
     * @param {number} [limit=20] - Maximum number of versions to return
     * @param {number} [offset=0] - Number of versions to skip
     * @returns {Promise<VersionsListResponse>} Paginated list of versions
     */
    list: (functionId, limit = 20, offset = 0) =>
      gqlRequest(
        `query ($id: ID!, $limit: Int!, $offset: Int!) {
          versions(functionId: $id, limit: $limit, offset: $offset) {
            nodes { ${VERSION_SUMMARY_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        { id: functionId, limit, offset },
      ).then((data) => ({
        versions: data.versions.nodes.map(mapVersion),
        pagination: data.versions.pageInfo,
      })),

    /**
     * Gets a specific version.
     * @param {string} functionId - Function ID
     * @param {number} version - Version number
     * @returns {Promise<FunctionVersion>} The version
     */
    get: (functionId, version) =>
      gqlRequest(
        `query ($id: ID!, $version: Int!) {
          version(functionId: $id, version: $version) { ${VERSION_FULL_FIELDS} }
        }`,
        { id: functionId, version },
      ).then((data) => mapVersion(data.version)),

    /**
     * Activates a specific version.
     * @param {string} functionId - Function ID
     * @param {string} versionId - Version ID to activate
     * @returns {Promise<void>}
     */
    activate: (functionId, versionId) =>
      gqlRequest(
        `mutation ($id: ID!, $versionId: ID!) {
          activateVersion(functionId: $id, versionId: $versionId) { id }
        }`,
        { id: functionId, versionId },
      ),

    /**
     * Gets a function (header fields) together with a version diff in one
     * round-trip. Used by the version-diff view on load.
     * @param {string} functionId - Function ID
     * @param {number} v1 - First version number
     * @param {number} v2 - Second version number
     * @returns {Promise<{func: LunarFunction, diff: DiffResponse}>}
     */
    diffWithFunction: (functionId, v1, v2) =>
      gqlRequest(
        `query ($id: ID!, $v1: Int!, $v2: Int!) {
          function(id: $id) { ${FUNCTION_SUMMARY_FIELDS} }
          versionDiff(functionId: $id, oldVersion: $v1, newVersion: $v2) {
            ${DIFF_FIELDS}
          }
        }`,
        { id: functionId, v1, v2 },
      ).then((data) => ({
        func: mapFunction(data.function),
        diff: mapDiff(data.versionDiff),
      })),

    /**
     * Deletes a specific version.
     * @param {string} functionId - Function ID
     * @param {string} versionId - Version ID to delete
     * @returns {Promise<void>}
     */
    delete: (functionId, versionId) =>
      gqlRequest(
        `mutation ($id: ID!, $versionId: ID!) {
          deleteVersion(functionId: $id, versionId: $versionId)
        }`,
        { id: functionId, versionId },
      ),
  },

  /**
   * Execution history methods.
   * @namespace
   */
  executions: {
    /**
     * Lists executions for a function.
     * @param {string} functionId - Function ID
     * @param {number} [limit=20] - Maximum number of executions to return
     * @param {number} [offset=0] - Number of executions to skip
     * @returns {Promise<ExecutionsListResponse>} Paginated list of executions
     */
    list: (functionId, limit = 20, offset = 0) =>
      gqlRequest(
        `query ($id: ID!, $limit: Int!, $offset: Int!) {
          executions(functionId: $id, limit: $limit, offset: $offset) {
            nodes { ${EXECUTION_SUMMARY_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        { id: functionId, limit, offset },
      ).then((data) => ({
        executions: data.executions.nodes.map(mapExecution),
        pagination: data.executions.pageInfo,
      })),

    /**
     * Gets an execution together with its parent function, logs, AI requests,
     * and email requests in a single round-trip — used by the execution-detail
     * view, which previously fired five separate requests on load. Pagination of
     * the individual sub-lists still uses the dedicated methods below.
     * @param {string} id - Execution ID
     * @param {Object} [opts] - Initial pagination for each sub-list
     * @returns {Promise<Object>} { execution, func, logs, logsTotal, aiRequests,
     *   aiRequestsTotal, emailRequests, emailRequestsTotal }
     */
    getDetail: (id, opts = {}) => {
      const {
        logsLimit = 20,
        logsOffset = 0,
        aiLimit = 20,
        aiOffset = 0,
        emailLimit = 20,
        emailOffset = 0,
      } = opts;
      return gqlRequest(
        `query (
          $id: ID!, $logsLimit: Int!, $logsOffset: Int!,
          $aiLimit: Int!, $aiOffset: Int!, $emailLimit: Int!, $emailOffset: Int!
        ) {
          execution(id: $id) {
            ${EXECUTION_FULL_FIELDS}
            function { ${FUNCTION_SUMMARY_FIELDS} }
          }
          executionLogs(executionId: $id, limit: $logsLimit, offset: $logsOffset) {
            nodes { ${LOG_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
          executionAiRequests(executionId: $id, limit: $aiLimit, offset: $aiOffset) {
            nodes { ${AI_REQUEST_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
          executionEmailRequests(executionId: $id, limit: $emailLimit, offset: $emailOffset) {
            nodes { ${EMAIL_REQUEST_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        {
          id,
          logsLimit,
          logsOffset,
          aiLimit,
          aiOffset,
          emailLimit,
          emailOffset,
        },
      ).then((data) => {
        const exec = data.execution;
        return {
          execution: mapExecution(exec),
          func: exec ? mapFunction(exec.function) : null,
          logs: data.executionLogs.nodes.map(mapLog),
          logsTotal: data.executionLogs.pageInfo?.total || 0,
          aiRequests: data.executionAiRequests.nodes.map(mapAIRequest),
          aiRequestsTotal: data.executionAiRequests.pageInfo?.total || 0,
          emailRequests: data.executionEmailRequests.nodes.map(mapEmailRequest),
          emailRequestsTotal: data.executionEmailRequests.pageInfo?.total || 0,
        };
      });
    },

    /**
     * Gets logs for an execution.
     * @param {string} executionId - Execution ID
     * @param {number} [limit=20] - Maximum number of log entries to return
     * @param {number} [offset=0] - Number of log entries to skip
     * @returns {Promise<ExecutionLogsResponse>} Paginated list of logs
     */
    getLogs: (executionId, limit = 20, offset = 0) =>
      gqlRequest(
        `query ($id: ID!, $limit: Int!, $offset: Int!) {
          executionLogs(executionId: $id, limit: $limit, offset: $offset) {
            nodes { ${LOG_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        { id: executionId, limit, offset },
      ).then((data) => ({
        logs: data.executionLogs.nodes.map(mapLog),
        pagination: data.executionLogs.pageInfo,
      })),

    /**
     * Gets AI requests for an execution.
     * @param {string} executionId - Execution ID
     * @param {number} [limit=20] - Maximum number of AI requests to return
     * @param {number} [offset=0] - Number of AI requests to skip
     * @returns {Promise<AIRequestsResponse>} Paginated list of AI requests
     */
    getAIRequests: (executionId, limit = 20, offset = 0) =>
      gqlRequest(
        `query ($id: ID!, $limit: Int!, $offset: Int!) {
          executionAiRequests(executionId: $id, limit: $limit, offset: $offset) {
            nodes { ${AI_REQUEST_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        { id: executionId, limit, offset },
      ).then((data) => ({
        ai_requests: data.executionAiRequests.nodes.map(mapAIRequest),
        pagination: data.executionAiRequests.pageInfo,
      })),

    /**
     * Gets email requests for an execution.
     * @param {string} executionId - Execution ID
     * @param {number} [limit=20] - Maximum number of email requests to return
     * @param {number} [offset=0] - Number of email requests to skip
     * @returns {Promise<EmailRequestsResponse>} Paginated list of email requests
     */
    getEmailRequests: (executionId, limit = 20, offset = 0) =>
      gqlRequest(
        `query ($id: ID!, $limit: Int!, $offset: Int!) {
          executionEmailRequests(executionId: $id, limit: $limit, offset: $offset) {
            nodes { ${EMAIL_REQUEST_FIELDS} }
            pageInfo { ${PAGE_INFO} }
          }
        }`,
        { id: executionId, limit, offset },
      ).then((data) => ({
        email_requests: data.executionEmailRequests.nodes.map(mapEmailRequest),
        pagination: data.executionEmailRequests.pageInfo,
      })),
  },

  /**
   * Executes a function with the given request parameters.
   * @param {string} functionId - Function ID to execute
   * @param {ExecuteRequest} request - Request parameters
   * @returns {Promise<ExecuteResponse>} Execution response with headers
   */
  execute: (functionId, request) => {
    // Handle query as either string or object
    let queryString = "";
    if (request.query) {
      if (typeof request.query === "string") {
        queryString = request.query.startsWith("?")
          ? request.query
          : "?" + request.query;
      } else {
        const params = new URLSearchParams(request.query);
        queryString = params.toString() ? "?" + params : "";
      }
    }
    const pathSuffix = request.path || "";
    const url = `/fn/${functionId}${pathSuffix}${queryString}`;

    // Parse body if it's a JSON string to avoid double-encoding
    let body;
    if (request.body) {
      try {
        body = JSON.parse(request.body);
      } catch {
        body = request.body;
      }
    }

    return m.request({
      method: request.method || "GET",
      url: url,
      body: body,
      headers: request.headers,
      /**
       * Extracts response data including execution headers.
       * @param {XMLHttpRequest} xhr - The XHR object
       * @returns {ExecuteResponse} Formatted response
       */
      extract: (xhr) => ({
        status: xhr.status,
        body: xhr.responseText,
        headers: {
          "X-Function-Id": xhr.getResponseHeader("X-Function-Id"),
          "X-Function-Version-Id": xhr.getResponseHeader(
            "X-Function-Version-Id",
          ),
          "X-Execution-Id": xhr.getResponseHeader("X-Execution-Id"),
          "X-Execution-Duration-Ms": xhr.getResponseHeader(
            "X-Execution-Duration-Ms",
          ),
        },
      }),
    });
  },
};
