# BotLogic

BotLogic is a standalone symbolic reasoning service for bots and automation.

It embeds a Prolog engine and exposes a small deterministic HTTP API for loading facts/rules and running queries. BotLogic is intentionally independent of BotAI, BotWeb and PBMP; integrations are optional layers.

## M0

M0 provides:

- embedded Prolog runtime
- `POST /v1/consult` for facts and rules
- `POST /v1/query` for deterministic queries
- `GET /healthz`
- `GET /readyz`
- bounded request sizes
- query timeout
- unit tests
- Alpine OCI image
- GitHub Actions CI

Default listen address: `127.0.0.1:8091`.
Override with `BOTLOGIC_LISTEN`.

## Example

Load rules:

```json
{"source":"trusted(alice). may_voice(X) :- trusted(X)."}
```

Query:

```json
{"query":"may_voice(X)."}
```

Response:

```json
{"solutions":[{"X":"alice"}]}
```

## M0.1

**M0.1 — PASS.** BotLogic supports named, isolated and persistent rulesets.

- rulesets are separate Prolog engines
- `GET /v1/rulesets` lists loaded rulesets
- `POST /v1/consult` accepts `ruleset` and atomically replaces that ruleset
- `POST /v1/query` requires `ruleset`
- `BOTLOGIC_DATA_DIR` selects persistent storage (default `./data`)
- rulesets reload on startup
- names are validated to prevent path traversal

Example:

```json
{"ruleset":"irc-policy","source":"trusted(alice). may_voice(X) :- trusted(X)."}
```

```json
{"ruleset":"irc-policy","query":"may_voice(X)."}
```

## M0.2

**M0.2 — PASS.** Rulesets can return explicit deterministic explanations through `POST /v1/explain`.

Explanation queries must bind a `Proof` variable. M0.2 defines `Proof` as a portable textual proof identifier/trace supplied by the ruleset; BotLogic never invents an explanation for an arbitrary goal.

Example rule:

```prolog
trusted(alice).
may_voice(X, 'trusted-user') :- trusted(X).
```

Explain query:

```json
{"ruleset":"irc-policy","query":"may_voice(alice, Proof)."}
```

This boundary lets a caller such as BotAI turn the deterministic proof into natural language without making the LLM the authority for the decision.

## M0.3

**M0.3 — PASS.** Dynamic state can be managed as structured facts instead of generated Prolog source.

- `POST /v1/facts` asserts a fact
- `DELETE /v1/facts` retracts a fact
- `GET /v1/facts?ruleset=...` lists facts
- predicates and arity are bounded and validated
- arguments are serialized as data, preventing Prolog source injection
- facts persist separately from static `.pl` rules and reload on restart
- retracting the last fact leaves an empty relation rather than an undefined-predicate error

Example fact:

```json
{"ruleset":"irc-policy","fact":{"predicate":"trusted","args":["alice"]}}
```

## M0.4

**M0.4 — PASS.** Dynamic facts support bounded atomic batch transactions.

- `POST /v1/facts/batch` applies up to 256 operations atomically
- operations are `assert` or `retract`
- the complete batch is validated before state changes
- a candidate Prolog engine is built before commit
- persistent fact state is atomically replaced before the live engine is published
- invalid batches leave both persisted and live state unchanged
- single-fact assert/retract use the same transaction core

Example:

```json
{"ruleset":"irc-state","operations":[{"op":"assert","fact":{"predicate":"online","args":["alice"]}},{"op":"assert","fact":{"predicate":"authenticated","args":["alice"]}}]}
```

## M0.5

**M0.5 — PASS.** Fact state is revisioned for snapshot identification and optimistic concurrency.

- every ruleset starts at revision `0`
- each committed fact transaction increments the revision
- revisions persist across restart
- fact mutations and batches accept optional `expected_revision`
- stale writers are rejected without changing state
- query, explain and fact-list responses report their revision
- successful mutation responses return the new revision

Example guarded transaction:

```json
{"ruleset":"irc-state","expected_revision":7,"operations":[{"op":"assert","fact":{"predicate":"online","args":["alice"]}}]}
```

A client can read revision 7, compute its update, and safely reject the write if another client has already advanced the ruleset to revision 8.

## M1.0

BotLogic M1.0 freezes the first stable standalone HTTP contract as **API v1**.

- release version: `1.0.0`
- API version: `v1`
- `GET /v1/version`
- `GET /v1/status`
- compatibility policy documented in `docs/API-v1.md`
- PBMP, BotWeb and BotAI remain optional and outside the core API dependency graph

**M1.0 — PASS.** The complete API v1 baseline passes tests, vet and Alpine OCI build.

## Architecture

```text
IRC bot / automation
        |
        +---- optional HTTP ----> BotLogic

BotWeb --PBMP--> bot             (optional management path)
bot ------------> BotAI          (optional LLM path)
bot ------------> BotLogic       (optional symbolic reasoning path)
```

BotLogic must remain useful as a standalone service. A bot must never require BotWeb, PBMP or BotAI merely to use BotLogic.

## License

MIT
