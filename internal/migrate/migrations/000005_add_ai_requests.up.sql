-- AI Requests tracking table
CREATE TABLE IF NOT EXISTS ai_requests (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    request_json TEXT NOT NULL,
    response_json TEXT,
    status TEXT NOT NULL,
    error_message TEXT,
    input_tokens INTEGER,
    output_tokens INTEGER,
    duration_ms INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (execution_id) REFERENCES executions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ai_requests_execution_id ON ai_requests(execution_id);
CREATE INDEX IF NOT EXISTS idx_ai_requests_created_at ON ai_requests(created_at);
