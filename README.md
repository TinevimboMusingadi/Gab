## Gab: Auditing and Versioned Activity Tracking for AI Agents

Gab is an append-only, tamper-evident activity tracking system for AI agents, inspired by Git. It records Events (actions) with pre/post State snapshots and Artifacts, forming a Merkle DAG for complete forensic traceability. This repository implements Phase 1 (core foundation) in Go, with a CLI and Badger-based content-addressable storage. Python integration and the deductive graph layer follow in later phases.

### Phases & Stages

- Phase 1: Core Foundation (Go)
  - Models: Artifact, State, ActionDescriptor, Event with canonical hashing
  - Storage: BadgerDB content-addressable object store
  - Index: Embedded heads and time-based indices
  - Recorder: Snapshot directory scope, execute command, record Event
  - CLI: init, record-cmd, log, show
  - Tests: models and storage

- Phase 2: Python Integration
  - gab-python client (subprocess wrapper) communicating with the Go daemon
  - End-to-end event recording

- Phase 3: Deductive Query Engine
  - External graph DB integration (e.g., Dgraph or PostgreSQL)
  - gab query command

- Phase 4: Expansion & Security
  - Additional language clients; mTLS and auth; performance optimizations

### Quickstart

1. Initialize the repo data directory:
   ```bash
   gab init --data ./gab_data
   ```
2. Record a command with directory snapshot scope:
   ```bash
   gab record-cmd --agent demo --scope . --cwd . -- ls -l
   ```
3. View the log for an agent:
   ```bash
   gab log --agent demo
   ```
4. Show an event by hash:
   ```bash
   gab show <event_hash>
   ```

### Repository Layout

- `cmd/gab/`: CLI entrypoint and commands
- `internal/models/`: Core data types and hashing
- `internal/hashutil/`: Canonicalization and hashing helpers
- `internal/storage/`: Badger-based object store
- `internal/index/`: Embedded indices (heads, time-based)
- `internal/recorder/`: Snapshot/execute/record logic
- `test/`: Unit tests for models and storage


