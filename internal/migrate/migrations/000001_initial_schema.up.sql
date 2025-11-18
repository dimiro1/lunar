-- KV Store
CREATE TABLE IF NOT EXISTS kv_store (
	function_id TEXT NOT NULL,
	key TEXT NOT NULL,
	value TEXT NOT NULL,
	PRIMARY KEY (function_id, key)
);

CREATE INDEX IF NOT EXISTS idx_function_id ON kv_store(function_id);

-- Environment Variables
CREATE TABLE IF NOT EXISTS env_vars (
	function_id TEXT NOT NULL,
	key TEXT NOT NULL,
	value TEXT NOT NULL,
	PRIMARY KEY (function_id, key)
);

CREATE INDEX IF NOT EXISTS idx_env_function_id ON env_vars(function_id);

-- Logs
CREATE TABLE IF NOT EXISTS logs (
	id TEXT PRIMARY KEY,
	execution_id TEXT NOT NULL,
	level INTEGER NOT NULL,
	message TEXT NOT NULL,
	timestamp INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_logs_execution_id ON logs(execution_id);
CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);

-- Functions
CREATE TABLE IF NOT EXISTS functions (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);

-- Function Versions
CREATE TABLE IF NOT EXISTS function_versions (
	id TEXT PRIMARY KEY,
	function_id TEXT NOT NULL,
	version INTEGER NOT NULL,
	code TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	created_by TEXT,
	is_active INTEGER NOT NULL DEFAULT 0,
	FOREIGN KEY (function_id) REFERENCES functions(id) ON DELETE CASCADE,
	UNIQUE(function_id, version)
);

CREATE INDEX IF NOT EXISTS idx_function_versions_function_id ON function_versions(function_id);
CREATE INDEX IF NOT EXISTS idx_function_versions_active ON function_versions(function_id, is_active);

-- Executions
CREATE TABLE IF NOT EXISTS executions (
	id TEXT PRIMARY KEY,
	function_id TEXT NOT NULL,
	function_version_id TEXT NOT NULL,
	status TEXT NOT NULL,
	duration_ms INTEGER,
	error_message TEXT,
	created_at INTEGER NOT NULL,
	FOREIGN KEY (function_id) REFERENCES functions(id) ON DELETE CASCADE,
	FOREIGN KEY (function_version_id) REFERENCES function_versions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_executions_function_id ON executions(function_id);
