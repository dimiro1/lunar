# Per-function metrics via a pre-aggregated rollup table

## Context

We want a per-function metrics dashboard (execution count, error rate, average
and max duration, trended over time). Every execution already writes a row to
`executions` with `status`, `duration_ms`, and `created_at`, so the metrics
*could* be computed on read by aggregating that table directly. But executions
are deleted by per-function retention — **default 7 days** — so on-read
aggregation can only ever show the last week. We want metrics to survive much
longer than the raw executions they came from.

## Decision

Store metrics as a **pre-aggregated rollup table**, `execution_metrics`, keyed
by `(function_id, bucket_hour)` (one row per function per UTC hour), holding
`count`, `error_count`, `sum_duration_ms`, and `max_duration_ms`. Buckets are
updated by a **best-effort** SQLite UPSERT at execution completion (alongside the
existing `UpdateExecution` write); they are kept for `METRICS_RETENTION_DAYS`
(default 365) and cleaned up by the existing housekeeping scheduler. The
GraphQL API exposes a lazy `Function.metrics(from, to, granularity)` field that
returns server-computed derived values (`errorRate`, `avgDurationMs`) and
downsamples hourly buckets to daily for long ranges. The table-creation
migration backfills buckets from existing executions.

## Considered Options

- **Compute-on-read from `executions` (no new table).** Rejected: cannot show
  history beyond the 7-day execution-retention window, and aggregates scan many
  rows per dashboard load for high-volume functions.
- **In-memory counters.** Rejected: lost on restart and inconsistent with the
  SQLite-everything datastore (ADR-0008).
- **Transactional (exact) bucket writes.** Rejected in favour of best-effort to
  keep the execution write path simple and never let a metrics write roll back a
  real execution record. See consequences.
- **Percentile latency (p50/p95).** Deferred: percentiles cannot be recovered
  from sums and counts; true percentiles would require per-bucket histograms.
  v1 ships average and max duration only.

## Consequences

- The rollup is **durable but approximate**: a crash or a failed UPSERT can
  permanently undercount a bucket, and because the source executions are
  eventually deleted, buckets can never be reconciled. These counts are an
  observability signal, not exact accounting.
- Adding breakdown dimensions later (by trigger, by version) or percentiles
  (histograms) is an additive migration — the v1 schema deliberately carries
  neither.
