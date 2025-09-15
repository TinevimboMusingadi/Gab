# Gab Architecture and Design

Gab is an append-only, tamper-evident activity tracking and inference system for AI agents. It records Events with pre/post State snapshots and Artifacts, forming a Merkle DAG, and emits Datalog-style facts for deductive analysis. This document explains the philosophy, data structures, algorithms, storage, CLI, and the planned deductive engine integration inspired by Mangle.

## Vision

- Action-centric: the unit is an `Event` (what happened), not a file diff.
- State-oriented: we snapshot relevant `State` pre/post action to capture system context.
- Immutability: history is cryptographically linked; tampering is evident.
- Deduction-ready: every `Event` also emits facts for Datalog-style reasoning.

## Core Data Model

- Artifact: content-addressed blob of bytes with a `ContentType` (e.g., file, stdout, stderr, exit_code). ID = sha256(ContentType || 0x00 || Data).
- State: mapping `identifier -> ArtifactID` (e.g., relative file path to its content artifact). Canonicalized and hashed to ID.
- ActionDescriptor: intent and execution metadata (type, params, agent, tool, timestamps).
- Event: references parent event, pre/post state IDs, action descriptor, and resulting artifact IDs. The JSON of the event (sans ID) is hashed to form the `Event.ID`.

These form a Merkle DAG: Artifact -> State -> Event. Mutating any object changes all descendant hashes.

## Record Event Algorithm

1) Pre-snapshot: Walk the provided scope directory, skipping heavy/irrelevant paths (`.git`, `gab_data`), create `Artifact`s for files, assemble a `State` and store by ID.
2) Execute: Run the command with captured stdout/stderr/exit_code.
3) Post-snapshot: Repeat the snapshot; generate a new `State` by ID.
4) Persist: Create an `Event` with parent pointer (agent head), pre/post states, action, and result artifact IDs; compute event ID; store.
5) Index: Update agent head and a time index; record event time for timelines.
6) Emit facts: Append Datalog-like EDB facts capturing `Event_happened`, `Parent_of`, `Action_executed`, outputs, and file creates/modifies/deletes.

Security property: Any retroactive change to a past object ripples through hashes, invalidating the subsequent chain.

## Storage and Indexing

- Object store: BadgerDB under `<data>/objects` for all content-addressed JSON-serialized, zlib-compressed objects.
- Index: BadgerDB under `<data>/index` for per-agent `head:<agent>` and time-ordered entries `time:<agent>:<rev_ts>:<event>`.
- Deductive facts: `<data>/deduce/facts.dl` for append-only EDB facts; `<data>/deduce/rules.dl` for rulebook.

## CLI

- `gab init --data <dir>`: bootstrap data directories and stores.
- `gab record-cmd --agent <id> --scope <dir> --cwd <dir> -- <cmd> [args]`: run with pre/post snapshots, persist event, emit facts.
- `gab log --agent <id>`: print reverse chronological event lineage with timestamps.
- `gab show <event_hash>`: print JSON of an event.
- `gab timeline --agent <id> [--limit N]`: time-ordered listing of recent events.
- `gab facts`: path and preview of the EDB facts file.
- `gab query -e "<expr>"`: placeholder Datalog query; will be backed by a real engine.
- `gab rules`: install default starter rules.

## Python Client

A thin wrapper shells out to the CLI for early integration:
- `gab.init(data_dir=...)`
- `gab.exec([cmd, ...], agent_id=..., scope_dir=..., cwd=..., env=...)` → returns event JSON via `gab show`.

This enables end-to-end tracking from Python agents prior to gRPC.

## Deductive Database: From Graph to Datalog

Gab emits facts to enable inference, not only traversal. The design draws on Datalog and Mangle-like semantics:
- EDB (facts): ground truth asserted by the recorder per event (e.g., `Event_happened`, `Parent_of`, `Action_executed`, `Action_produced_output`, file changes, etc.).
- IDB (rules): rules that define derived relations like `Ancestor_of`, `Tainted_content`, `Potential_exfiltration`.

Example rules (starter set):

```
// Transitive closure over parent links
Ancestor_of(A, D) :- Parent_of(D, A).
Ancestor_of(A, D) :- Parent_of(D, M), Ancestor_of(A, M).

// Taint propagation from known secrets to outputs of same event
Tainted_content(C) :- Content_read_by_event(SC, E), Secret_content(SC), Action_produced_output(E, A, _), Artifact_is_file(A, _, C).

// Potential data exfiltration based on taint and network activity
Potential_exfiltration(E, C, IP) :- Action_made_network_call(E, IP, _), Content_read_by_event(C, E), Tainted_content(C).
```

Future integration will use a real deductive engine so that queries like `?[file, event] := Tainted_content(c), Artifact_is_file(_, file, c), Action_produced_output(event, _, c)` can be executed efficiently.

## Mangle-Inspired Engine Integration

Mangle provides a Datalog-derived language with practical extensions (aggregation, function calls, optional type checking). Its Go implementation can be embedded into services. See Mangle: [github.com/google/mangle](https://github.com/google/mangle).

Planned integration (Phase 3.3):
- Build-tagged implementation of a `deduce.Engine` backed by Mangle to parse and evaluate queries against in-memory or persisted facts.
- Rule loading from `<data>/deduce/rules.dl`.
- Execution of user queries from `gab query` with bindings output.

Alternative/adjacent engines for consideration: CozoDB, Soufflé, Flora-2. We will begin with a Mangle-aligned API for easier rule authoring and recursion (transitive closure) support.

## Novel Engineering Details

- Canonical hashing:
  - `ArtifactID = sha256(ContentType || 0x00 || Data)` ensures low collision risk and content-addressability across modalities.
  - `StateID = sha256("State" || canonical(sorted pairs of Resources))` for deterministic IDs.
  - `EventID = sha256(JSON(Event without ID))` to avoid map-order nondeterminism and keep an audit-friendly envelope.
- Snapshots with excludes: Pre/post file system snapshots skip `.git` and `gab_data` to prevent self-inclusion and heavy tails.
- Append-only facts: We write human-inspectable `.dl` lines alongside object storage, enabling offline investigation and deterministic replays.
- Time index using reversed timestamps in Badger for efficient most-recent-first scans.
- CLI and Python wrapper for fast adoption; gRPC comes next.

## Security and Integrity

- Append-only Merkle DAG makes tampering detectable.
- Separation of blobs (Badger) from facts/rules improves safety and performance.
- Future mTLS/auth around gRPC; access controls on query endpoints.

## Roadmap Status

- Phase 1: Core in Go: models, hashing, CAS store, indices, recorder, CLI, tests — done.
- Phase 2: Python client wrapper — done.
- Phase 3: Deductive engine: facts file and rules installer — in progress; query placeholder in place; Mangle integration next.
- Phase 4: Security hardening, performance, language SDKs — planned.

## References

- Mangle (Datalog with extensions) — [github.com/google/mangle](https://github.com/google/mangle)
