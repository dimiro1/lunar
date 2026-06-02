# 0001. Record architecture decisions

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar is a self-hosted FaaS platform with a number of deliberate, non-obvious
architectural choices: configuration through the environment, a build-step-free
Mithril frontend, no Node.js in the toolchain, `mise` for tasks, `uber/fx` for
wiring, a single embedded binary, and pure-Go SQLite. These decisions are easy
to misread as accidents when only the resulting code is visible. New
contributors (and our future selves) repeatedly ask "why is it done this way?",
and without a record the answer lives only in memory and scattered commit
messages.

We want a durable, low-ceremony way to capture the *reasoning* behind decisions
that shape the codebase, separate from the code that implements them and from the
user-facing README.

## Decision

We will keep Architecture Decision Records as Markdown files under `docs/adr/`,
one decision per numbered file, following the lightweight format in
[`0000-template.md`](0000-template.md) (Status, Context, Decision,
Consequences).

ADRs are append-only. Once a record is `Accepted` we do not edit its substance;
a decision that changes is recorded as a new ADR that supersedes the old one,
and the old one's status is updated to point at its replacement.

## Consequences

- The rationale behind a choice travels with the repository and is versioned
  alongside the code, reviewable in the same pull requests.
- There is a small, well-understood cost to writing a record when a decision is
  made. We accept this as cheaper than re-litigating decisions later.
- Reviewers gain a natural place to push back on direction before it is encoded
  in the codebase.
- The README stays focused on *using* Lunar; the ADRs explain *building* it.
