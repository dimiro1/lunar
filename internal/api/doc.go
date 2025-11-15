// Package api provides the HTTP API server for the FaaS platform.
//
// The API implements the OpenAPI specification defined in openapi.yaml and provides
// endpoints for managing functions, versions, executions, and runtime execution.
//
// Main endpoint groups:
//   - /api/functions - Function management (CRUD)
//   - /api/functions/{id}/versions - Version management
//   - /api/executions - Execution history and logs
//   - /fn/{function_id} - Runtime function execution
package api
