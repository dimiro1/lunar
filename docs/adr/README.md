# Architecture Decision Records

This directory records the significant architectural decisions made on Lunar,
using lightweight [ADRs](https://adr.github.io/). Each record captures the
context in force at the time, the decision taken, and its consequences — so a
future reader can understand *why* the code looks the way it does without having
to reverse-engineer it or interrupt the people who were there.

## Conventions

- One decision per file, named `NNNN-kebab-case-title.md`, numbered in order.
- Records are immutable once `Accepted`. We don't rewrite history: to change a
  past decision, add a new ADR and set the old one's status to `Superseded by
  ADR-NNNN`.
- Start from [`0000-template.md`](0000-template.md).

## Index

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | Accepted |
| [0002](0002-configuration-via-environment-variables.md) | Configuration via environment variables | Accepted |
| [0003](0003-frontend-with-mithril.md) | Frontend with Mithril.js | Accepted |
| [0004](0004-no-nodejs-vendored-js.md) | No Node.js: vendored JS, Deno for tooling | Accepted |
| [0005](0005-mise-for-toolchain-and-tasks.md) | mise for toolchain and task management | Accepted |
| [0006](0006-dependency-injection-with-fx.md) | Dependency injection with uber/fx | Accepted |
| [0007](0007-single-binary-embedded-frontend.md) | Single binary with embedded frontend | Accepted |
| [0008](0008-sqlite-as-the-datastore.md) | Pure-Go SQLite as the datastore | Accepted |
| [0009](0009-lua-as-the-function-language.md) | Lua as the function language | Accepted |
| [0010](0010-in-house-i18n.md) | In-house i18n with locale modules | Accepted |
| [0011](0011-testing-strategy.md) | Layered testing strategy without a JS test runner | Accepted |
