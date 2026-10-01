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
