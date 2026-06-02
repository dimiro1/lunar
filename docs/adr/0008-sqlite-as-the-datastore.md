# 0008. Pure-Go SQLite as the datastore

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar stores functions, versions, execution history, logs, KV data, and related
state. As a self-hosted, single-binary tool (see
[ADR-0007](0007-single-binary-embedded-frontend.md)), requiring operators to
stand up and connect a separate database server (Postgres, MySQL) would
contradict the "no external dependencies, just run the binary" promise. We need
durable, transactional, SQL-capable storage that lives in the same process and on
the local filesystem.

The classic embedded choice is SQLite, but the most common Go driver
(`mattn/go-sqlite3`) requires CGo. CGo complicates cross-compilation, slows
builds, and undercuts the goal of a clean, portable single binary.

## Decision

We will use SQLite as the datastore, via the **pure-Go**
[`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) driver — no CGo.
The database is a file under the configured `DATA_DIR`, and we run it in
**WAL** mode for better read/write concurrency under the server's workload.

## Consequences

- Storage is embedded: no database server to provision, secure, or back up
  separately — backups are file copies.
- Staying CGo-free keeps cross-compilation simple and builds fast, preserving the
  portable single-binary story end to end.
- WAL mode lets readers proceed concurrently with a writer, which suits the
  dashboard-plus-execution workload. (WAL behaviour was tightened in commit
  `ba3e33a`.)
- Trade-off: SQLite is single-writer and local to one host, so this design does
  not target multi-node horizontal scaling. That is consistent with Lunar's
  self-hosted, lightweight positioning; a different topology would warrant a new
  ADR.
- The pure-Go driver historically trails the C library slightly on raw
  performance, which is an acceptable price for portability at this scale.
