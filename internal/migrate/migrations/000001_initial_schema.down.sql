-- Drop executions
DROP INDEX IF EXISTS idx_executions_function_id;
DROP TABLE IF EXISTS executions;

-- Drop function versions
DROP INDEX IF EXISTS idx_function_versions_active;
DROP INDEX IF EXISTS idx_function_versions_function_id;
DROP TABLE IF EXISTS function_versions;

-- Drop functions
DROP TABLE IF EXISTS functions;

-- Drop logs
DROP INDEX IF EXISTS idx_logs_timestamp;
DROP INDEX IF EXISTS idx_logs_level;
DROP INDEX IF EXISTS idx_logs_execution_id;
DROP TABLE IF EXISTS logs;

-- Drop environment variables
DROP INDEX IF EXISTS idx_env_function_id;
DROP TABLE IF EXISTS env_vars;

-- Drop KV store
DROP INDEX IF EXISTS idx_function_id;
DROP TABLE IF EXISTS kv_store;
