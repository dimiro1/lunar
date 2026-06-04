-- Per-function execution metrics, pre-aggregated into hourly buckets. Metrics
-- live far longer than the raw executions they are derived from (executions are
-- deleted by per-function retention, default 7 days), so the dashboard can show
-- long-range trends. One row per function per UTC hour.
-- See docs/adr/0014-metrics-rollup-table.md.
CREATE TABLE IF NOT EXISTS execution_metrics (
	function_id     TEXT    NOT NULL,
	bucket_hour     INTEGER NOT NULL, -- unix seconds truncated to the hour (UTC)
	count           INTEGER NOT NULL DEFAULT 0,
	error_count     INTEGER NOT NULL DEFAULT 0,
	sum_duration_ms INTEGER NOT NULL DEFAULT 0,
	max_duration_ms INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (function_id, bucket_hour),
	FOREIGN KEY (function_id) REFERENCES functions(id) ON DELETE CASCADE
);

-- Supports the housekeeping cleanup, which deletes buckets across all functions
-- older than the metrics-retention cutoff.
CREATE INDEX IF NOT EXISTS idx_execution_metrics_bucket_hour ON execution_metrics(bucket_hour);

-- Backfill from existing executions so the dashboard is populated on deploy.
-- Only completed executions are counted, matching the live best-effort increment
-- in the engine. History is naturally capped at whatever executions still exist
-- under their retention window.
INSERT INTO execution_metrics (function_id, bucket_hour, count, error_count, sum_duration_ms, max_duration_ms)
SELECT function_id,
       (created_at / 3600) * 3600 AS bucket_hour,
       COUNT(*),
       SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END),
       COALESCE(SUM(duration_ms), 0),
       COALESCE(MAX(duration_ms), 0)
FROM executions
WHERE status IN ('success', 'error')
GROUP BY function_id, bucket_hour;
