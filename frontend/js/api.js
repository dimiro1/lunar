/**
 * @fileoverview API client for the FaaS Dashboard.
 * Provides methods for authentication, function management, and execution.
 */

/**
 * @typedef {import('./types.js').FaaSFunction} FaaSFunction
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
 * API client for the FaaS Dashboard.
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
      apiRequest({
        method: "GET",
        url: `/api/functions?limit=${limit}&offset=${offset}`,
      }),

    /**
     * Gets a single function by ID.
     * @param {string} id - Function ID
     * @returns {Promise<FaaSFunction>} The function
     */
    get: (id) => apiRequest({ method: "GET", url: `/api/functions/${id}` }),

    /**
     * Creates a new function.
     * @param {Object} data - Function data
     * @param {string} data.name - Function name
     * @param {string} [data.description] - Function description
     * @param {string} data.code - Initial function code
     * @returns {Promise<FaaSFunction>} The created function
     */
    create: (data) =>
      apiRequest({ method: "POST", url: "/api/functions", body: data }),

    /**
     * Updates an existing function.
     * @param {string} id - Function ID
     * @param {Object} data - Update data
     * @param {string} [data.name] - New name
     * @param {string} [data.description] - New description
     * @param {string} [data.code] - New code (creates new version)
     * @param {boolean} [data.disabled] - Enable/disable function
     * @returns {Promise<FaaSFunction>} The updated function
     */
    update: (id, data) =>
      apiRequest({ method: "PUT", url: `/api/functions/${id}`, body: data }),

    /**
     * Deletes a function.
     * @param {string} id - Function ID
     * @returns {Promise<void>}
     */
    delete: (id) =>
      apiRequest({ method: "DELETE", url: `/api/functions/${id}` }),

    /**
     * Updates environment variables for a function.
     * @param {string} id - Function ID
     * @param {Object.<string, string>} env_vars - Environment variables
     * @returns {Promise<FaaSFunction>} The updated function
     */
    updateEnv: (id, env_vars) =>
      apiRequest({
        method: "PUT",
        url: `/api/functions/${id}/env`,
        body: { env_vars },
      }),
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
      apiRequest({
        method: "GET",
        url:
          `/api/functions/${functionId}/versions?limit=${limit}&offset=${offset}`,
      }),

    /**
     * Gets a specific version.
     * @param {string} functionId - Function ID
     * @param {number} version - Version number
     * @returns {Promise<FunctionVersion>} The version
     */
    get: (functionId, version) =>
      apiRequest({
        method: "GET",
        url: `/api/functions/${functionId}/versions/${version}`,
      }),

    /**
     * Activates a specific version.
     * @param {string} functionId - Function ID
     * @param {number} version - Version number to activate
     * @returns {Promise<FaaSFunction>} The updated function
     */
    activate: (functionId, version) =>
      apiRequest({
        method: "POST",
        url: `/api/functions/${functionId}/versions/${version}/activate`,
      }),

    /**
     * Gets a diff between two versions.
     * @param {string} functionId - Function ID
     * @param {number} v1 - First version number
     * @param {number} v2 - Second version number
     * @returns {Promise<DiffResponse>} The diff result
     */
    diff: (functionId, v1, v2) =>
      apiRequest({
        method: "GET",
        url: `/api/functions/${functionId}/diff/${v1}/${v2}`,
      }),
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
      apiRequest({
        method: "GET",
        url:
          `/api/functions/${functionId}/executions?limit=${limit}&offset=${offset}`,
      }),

    /**
     * Gets a specific execution.
     * @param {string} executionId - Execution ID
     * @returns {Promise<Execution>} The execution
     */
    get: (executionId) =>
      apiRequest({ method: "GET", url: `/api/executions/${executionId}` }),

    /**
     * Gets logs for an execution.
     * @param {string} executionId - Execution ID
     * @param {number} [limit=20] - Maximum number of log entries to return
     * @param {number} [offset=0] - Number of log entries to skip
     * @returns {Promise<ExecutionLogsResponse>} Paginated list of logs
     */
    getLogs: (executionId, limit = 20, offset = 0) =>
      apiRequest({
        method: "GET",
        url:
          `/api/executions/${executionId}/logs?limit=${limit}&offset=${offset}`,
      }),

    /**
     * Gets AI requests for an execution.
     * @param {string} executionId - Execution ID
     * @param {number} [limit=20] - Maximum number of AI requests to return
     * @param {number} [offset=0] - Number of AI requests to skip
     * @returns {Promise<AIRequestsResponse>} Paginated list of AI requests
     */
    getAIRequests: (executionId, limit = 20, offset = 0) =>
      apiRequest({
        method: "GET",
        url:
          `/api/executions/${executionId}/ai-requests?limit=${limit}&offset=${offset}`,
      }),
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
    const url = `/fn/${functionId}${queryString}`;

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
