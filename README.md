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
