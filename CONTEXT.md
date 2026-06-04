# Lunar

Lunar is a self-hosted, single-binary platform for writing and running small
functions (in Lua or Starlark) triggered over HTTP or by cron. This glossary
fixes the language used across the codebase, GraphQL API, and UI.

## Language

**Function**:
A named, versioned unit of code that runs in response to a trigger. The
top-level thing a user manages.
_Avoid_: Lambda, script, handler, endpoint.

**Function Version**:
A specific revision of a function's source code. One version of a function is
active at a time and is what executions run.
_Avoid_: Revision, snapshot, deployment.

**Execution**:
A single run of a function. Records its outcome (success/error), wall-clock
duration, trigger, and timestamp. Subject to per-function retention, so old
executions are eventually deleted.
_Avoid_: Invocation, run, call, request.

**Trigger**:
What caused an execution — `http` (an inbound request) or `cron` (a schedule).
_Avoid_: Source, cause, event type.

**Metric**:
An aggregate measure derived from many executions over a time window — e.g.
execution count, error rate, or average duration. Unlike an execution, a metric
outlives retention: it is stored as a pre-aggregated rollup so history survives
after the underlying executions are deleted.
_Avoid_: Stat, measurement, telemetry, gauge.
