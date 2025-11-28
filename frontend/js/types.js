/**
 * @fileoverview Type definitions for the FaaS Dashboard frontend.
 * These JSDoc typedefs provide IDE autocompletion and type checking.
 */

// ============================================================================
// API Response Types
// ============================================================================

/**
 * @typedef {Object} Pagination
 * @property {number} total - Total number of items
 * @property {number} limit - Items per page
 * @property {number} offset - Current offset
 */

/**
 * @typedef {Object} FunctionVersion
 * @property {string} id - Version ID
 * @property {string} function_id - Parent function ID
 * @property {number} version - Version number
 * @property {string} code - Function source code
 * @property {string} created_at - ISO timestamp
 */

/**
 * @typedef {Object} FaaSFunction
 * @property {string} id - Function ID
 * @property {string} name - Function name
 * @property {string} [description] - Optional description
 * @property {boolean} disabled - Whether function is disabled
 * @property {FunctionVersion} active_version - Currently active version
 * @property {Object.<string, string>} [env_vars] - Environment variables
 * @property {string} created_at - ISO timestamp
 * @property {string} updated_at - ISO timestamp
 */

/**
 * @typedef {Object} FunctionsListResponse
 * @property {FaaSFunction[]} functions - List of functions
 * @property {Pagination} pagination - Pagination info
 */

/**
 * @typedef {Object} VersionsListResponse
 * @property {FunctionVersion[]} versions - List of versions
 * @property {Pagination} pagination - Pagination info
 */

/**
 * @typedef {Object} Execution
 * @property {string} id - Execution ID
 * @property {string} function_id - Function ID
 * @property {string} version_id - Version ID that was executed
 * @property {number} version - Version number
 * @property {string} status - Execution status (success, error, timeout)
 * @property {number} duration_ms - Execution duration in milliseconds
 * @property {number} [status_code] - HTTP status code returned
 * @property {string} created_at - ISO timestamp
 */

/**
 * @typedef {Object} ExecutionsListResponse
 * @property {Execution[]} executions - List of executions
 * @property {Pagination} pagination - Pagination info
 */

/**
 * @typedef {Object} ExecutionLog
 * @property {string} id - Log entry ID
 * @property {string} execution_id - Parent execution ID
 * @property {string} level - Log level (INFO, WARN, ERROR, DEBUG)
 * @property {string} message - Log message
 * @property {string} timestamp - ISO timestamp
 */

/**
 * @typedef {Object} ExecutionLogsResponse
 * @property {ExecutionLog[]} logs - List of log entries
 * @property {Pagination} pagination - Pagination info
 */

/**
 * @typedef {Object} AIRequest
 * @property {string} id - AI request ID
 * @property {string} execution_id - Parent execution ID
 * @property {string} provider - AI provider (openai, anthropic)
 * @property {string} model - Model name
 * @property {string} endpoint - API endpoint
 * @property {string} request_json - Request JSON
 * @property {string} [response_json] - Response JSON
 * @property {string} status - Status (success, error)
 * @property {string} [error_message] - Error message if failed
 * @property {number} [input_tokens] - Input token count
 * @property {number} [output_tokens] - Output token count
 * @property {number} duration_ms - Duration in milliseconds
 * @property {number} created_at - Unix timestamp
 */

/**
 * @typedef {Object} AIRequestsResponse
 * @property {AIRequest[]} ai_requests - List of AI requests
 * @property {Pagination} pagination - Pagination info
 */

/**
 * @typedef {Object} DiffResponse
 * @property {string} diff - Unified diff string
 * @property {FunctionVersion} version1 - First version
 * @property {FunctionVersion} version2 - Second version
 */

// ============================================================================
// Component Prop Types
// ============================================================================

/**
 * @typedef {Object} MithrilVnode
 * @property {Object} attrs - Component attributes/props
 * @property {Array} children - Child elements
 * @property {string} [key] - Optional key for list rendering
 */

/**
 * @typedef {Object} TabItem
 * @property {string} id - Tab identifier
 * @property {string} label - Tab display label
 * @property {string} href - Tab link URL
 */

/**
 * @typedef {Object} ToastMessage
 * @property {number} id - Unique message ID
 * @property {string} message - Message text
 * @property {('success'|'error'|'warning'|'info')} type - Toast type
 */

// ============================================================================
// Request Types
// ============================================================================

/**
 * @typedef {Object} FunctionCreateRequest
 * @property {string} name - Function name
 * @property {string} [description] - Optional description
 * @property {string} code - Initial function code
 */

/**
 * @typedef {Object} FunctionUpdateRequest
 * @property {string} [name] - New function name
 * @property {string} [description] - New description
 * @property {string} [code] - New code (creates new version)
 * @property {boolean} [disabled] - Enable/disable function
 */

/**
 * @typedef {Object} ExecuteRequest
 * @property {string} [method] - HTTP method (GET, POST, etc.)
 * @property {string|Object.<string, string>} [query] - Query parameters
 * @property {Object.<string, string>} [headers] - Request headers
 * @property {*} [body] - Request body
 */

/**
 * @typedef {Object} ExecuteResponse
 * @property {number} status - HTTP status code
 * @property {string} body - Response body
 * @property {Object} headers - Response headers with execution metadata
 * @property {string} headers.X-Function-Id - Function ID
 * @property {string} headers.X-Function-Version-Id - Version ID
 * @property {string} headers.X-Execution-Id - Execution ID
 * @property {string} headers.X-Execution-Duration-Ms - Duration in ms
 */

// ============================================================================
// Icon Types
// ============================================================================

/**
 * @typedef {() => string} IconFunction
 * A function that returns an SVG string for use with m.trust()
 */

/**
 * @typedef {'eye'|'eyeSlash'|'pencil'|'play'|'pause'|'plus'|'minus'|'plusSmall'|'minusSmall'|'arrowLeft'|'xMark'|'sun'|'arrowPath'|'moon'|'arrowsRightLeft'|'clipboard'|'clipboardCheck'|'trash'|'undo'|'bolt'|'chevronLeft'|'chevronRight'|'check'|'copy'|'hashtag'|'globe'|'clock'|'server'|'document'|'code'|'cog'|'chartBar'|'beaker'|'listBullet'|'inbox'|'exclamationTriangle'|'search'|'spinner'|'funnel'|'magnifyingGlass'|'key'|'network'} IconName
 */

export {};
