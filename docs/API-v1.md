# BotLogic HTTP API v1

BotLogic 1.x exposes the stable `v1` HTTP contract described here.

## Compatibility

Within BotLogic 1.x:

- existing v1 endpoints and successful response fields are not removed or renamed;
- new optional request fields and new response fields may be added;
- clients must ignore response fields they do not understand;
- incompatible HTTP contract changes require a new API version.

BotLogic is standalone. The v1 API does not require PBMP, BotWeb or BotAI.

## Service endpoints

### GET /healthz
Liveness response.

### GET /readyz
Readiness response.

### GET /v1/version
Returns service, release version and API version.

### GET /v1/status
Returns privacy-safe service status and loaded ruleset count.

## Rulesets

### GET /v1/rulesets
Lists loaded ruleset names.

### POST /v1/validate\n\nCompiles Prolog source in a fresh interpreter without creating or replacing a ruleset, writing files, or changing revisions. Request: `{"source":"may_execute(A,C) :- allowed(A,C)."}`. Returns `{"ok":true,"valid":true}` when valid.\n\n### POST /v1/consult
Creates or atomically replaces static Prolog source.

Request:
```json
{"ruleset":"irc-policy","source":"may_voice(X) :- trusted(X)."}
```

## Queries

### POST /v1/query
Request:
```json
{"ruleset":"irc-policy","query":"may_voice(X)."}
```
Response includes `ruleset`, snapshot `revision`, and `solutions`.

### POST /v1/explain
Like query, but the goal must bind `Proof`. BotLogic returns only proof data supplied by the ruleset.

## Facts

Facts use:
```json
{"predicate":"trusted","args":["alice"]}
```

### GET /v1/facts?ruleset=NAME
Lists structured dynamic facts and the current revision.

### POST /v1/facts
Asserts one fact.

### DELETE /v1/facts
Retracts one fact.

Single mutations may include `expected_revision`.

### POST /v1/facts/batch
Atomically applies 1–256 operations:
```json
{"ruleset":"irc-policy","expected_revision":4,"operations":[{"op":"assert","fact":{"predicate":"trusted","args":["alice"]}}]}
```

A revision conflict rejects the complete transaction. Successful mutations return the new revision.

## Limits

- consult body: 64 KiB
- query/explain/single fact body: 16 KiB
- batch body: 256 KiB
- batch operations: 256
- fact arguments: 16
- fact argument length: 1024 bytes
- query/explain timeout: 2 seconds
