-- Email Requests tracking table
CREATE TABLE IF NOT EXISTS email_requests (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL,
    from_address TEXT NOT NULL,
    to_addresses TEXT NOT NULL,
    subject TEXT NOT NULL,
    has_text INTEGER NOT NULL DEFAULT 0,
    has_html INTEGER NOT NULL DEFAULT 0,
    request_json TEXT NOT NULL,
    response_json TEXT,
    status TEXT NOT NULL,
    error_message TEXT,
    email_id TEXT,
    duration_ms INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (execution_id) REFERENCES executions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_email_requests_execution_id ON email_requests(execution_id);
CREATE INDEX IF NOT EXISTS idx_email_requests_created_at ON email_requests(created_at);
