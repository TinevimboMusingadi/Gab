# Gab Codebase Summary

This document provides a comprehensive overview of the current Gab codebase implementation, its architecture, components, and how they work together.

## Overview

Gab is a tamper-evident activity tracking system for AI agents that records Events (actions) with pre/post State snapshots and Artifacts, forming a Merkle DAG. The system is implemented primarily in Go with a Python client wrapper.

## Core Architecture

### Data Model (Merkle DAG)

The system is built around three core data structures that form a content-addressable Merkle DAG:

1. **Artifact** (`internal/models/models.go`)
   - Content-addressed blob of bytes with a `ContentType` (e.g., "file", "stdout", "stderr", "exit_code")
   - ID computed as: `SHA256(ContentType || 0x00 || Data)`
   - Represents immutable content (file contents, command outputs, etc.)

2. **State** (`internal/models/models.go`)
   - Mapping of `identifier -> ArtifactID` (e.g., relative file path to its content artifact)
   - Canonicalized and hashed deterministically (sorted keys)
   - Represents a snapshot of system state at a point in time

3. **Event** (`internal/models/models.go`)
   - References parent event, pre/post state IDs, action descriptor, and resulting artifact IDs
   - ID computed as: `SHA256(JSON(event without ID field))`
   - Forms the chain of actions taken by an agent

**Security Property**: Any retroactive change to a past object ripples through hashes, invalidating the entire descendant chain.

## Components

### 1. Storage Layer (`internal/storage/badger_store.go`)

- **Technology**: BadgerDB (embedded key-value store)
- **Location**: `<data_dir>/objects/`
- **Features**:
  - Content-addressable storage (objects stored by their SHA-256 hash)
  - JSON serialization with zlib compression
  - Put/Get operations for any object with an `ID` field

**Key Functions**:
- `Init(dataDir)`: Initialize BadgerDB store
- `Open(dataDir)`: Open existing store
- `PutObject(id, object)`: Store object by hash
- `GetObject(id, out)`: Retrieve object by hash

### 2. Index Layer (`internal/index/index.go`)

- **Technology**: BadgerDB (separate database)
- **Location**: `<data_dir>/index/`
- **Indexes**:
  - **Head Index**: `head:<agentID>` → latest event hash for each agent
  - **Time Index**: `time:<agentID>:<reverse_ts>:<eventHash>` → time-ordered events

**Key Functions**:
- `SetHead(agentID, eventHash)`: Update agent's latest event
- `GetHead(agentID)`: Get agent's latest event hash
- `AddEventTime(agentID, eventHash, time)`: Index event by timestamp
- `ListByTime(agentID, limit)`: Get recent events in chronological order

**Time Indexing Algorithm**: Uses reversed UnixNano timestamps for lexicographic descending order, enabling efficient time-range queries.

### 3. Hash Utilities (`internal/hashutil/hashutil.go`)

Provides canonical hashing functions:
- `HashBytes(b)`: SHA-256 hash of bytes
- `HashStrings(parts)`: Deterministic hash of ordered string list
- `CanonicalizeMap(m)`: Stable key ordering and representation for maps

### 4. Recorder (`internal/recorder/recorder.go`)

The core component that captures and records events.

**Record Command Algorithm**:
1. **Pre-snapshot**: Walk scope directory, create Artifacts for files, assemble State
2. **Execute**: Run command, capture stdout/stderr/exit_code as Artifacts
3. **Post-snapshot**: Repeat snapshot to generate new State
4. **Persist**: Create Event with parent pointer, pre/post states, action, result artifacts
5. **Index**: Update agent head and time index
6. **Emit Facts**: Append Datalog-style EDB facts for deductive reasoning

**Key Features**:
- Automatic exclusion of `.git` and `gab_data` directories
- Command execution with timeout (5 minutes default)
- Environment variable passthrough
- Automatic artifact creation for command outputs

### 5. Deductive Facts Writer (`internal/deduce/writer.go`)

Emits Datalog-style facts for each event to enable deductive reasoning.

**Facts Emitted**:
- `Event_happened(eventID, agentID, tool, timestamp)`
- `Parent_of(childEventID, parentEventID)`
- `Action_executed(eventID, command)`
- `Action_produced_output(eventID, artifactID, kind)`
- `File_created(eventID, path, artifactID)`
- `File_modified(eventID, path, oldArtifactID, newArtifactID)`
- `File_deleted(eventID, path, artifactID)`

**Location**: `<data_dir>/deduce/facts.dl` (append-only)

### 6. Rules System (`internal/deduce/rules.go`)

Provides default Datalog rules for deductive reasoning:
- `Ancestor_of`: Transitive closure of parent relationships
- `Tainted_content`: Content propagation from secrets
- `Potential_exfiltration`: Security analysis rules

**Location**: `<data_dir>/deduce/rules.dl`

### 7. CLI (`cmd/gab/`)

Command-line interface built with Cobra:

**Commands**:
- `gab init`: Initialize Gab data directory
- `gab record-cmd`: Record command execution with state snapshots
- `gab log`: Show event history for an agent (reverse chronological)
- `gab show <hash>`: Display full JSON for an event
- `gab timeline`: List recent events in time order
- `gab facts`: Show path and preview of facts file
- `gab query`: Placeholder Datalog query executor
- `gab rules`: Install default starter rules

**Global Flags**:
- `--data <dir>`: Data directory (default: `./gab_data`)

### 8. Python Client (`python/gab/`)

Thin wrapper that shells out to the CLI for early integration.

**API**:
- `gab.init(data_dir)`: Initialize Gab store
- `gab.exec_cmd(args, agent_id, scope_dir, cwd, env, data_dir)`: Execute command and return event JSON
- `gab.run(...)`: Alias for `exec_cmd`

**Implementation**: Uses subprocess to call `gab.exe` (Windows) or `gab` (Unix), parses output to extract event hash, then calls `gab show` to get full event JSON.

## Data Flow

### Recording an Event

```
User/Agent
    ↓
gab record-cmd (CLI) or gab.exec() (Python)
    ↓
Recorder.RecordCommand()
    ↓
1. Snapshot directory → Create Artifacts → Create Pre-State
    ↓
2. Execute command → Capture stdout/stderr/exit_code → Create Artifacts
    ↓
3. Snapshot directory again → Create Post-State
    ↓
4. Create Event (with parent, pre/post states, action, artifacts)
    ↓
5. Store Event in BadgerDB (objects/)
    ↓
6. Update Index (head + time)
    ↓
7. Emit facts to facts.dl
    ↓
Return Event ID
```

### Querying Events

```
User
    ↓
gab log --agent <id>
    ↓
Index.GetHead(agentID)
    ↓
Storage.GetObject(headHash) → Event
    ↓
Follow ParentEventHash chain backwards
    ↓
Display event lineage
```

## Storage Structure

```
<data_dir>/
├── objects/          # BadgerDB: Content-addressable object store
│   ├── MANIFEST
│   ├── *.sst         # SSTable files
│   └── *.vlog        # Value log files
├── index/            # BadgerDB: Index database
│   ├── MANIFEST
│   ├── *.sst
│   └── *.vlog
└── deduce/
    ├── facts.dl      # Append-only EDB facts (Datalog)
    └── rules.dl      # IDB rules (Datalog)
```

## Testing

**Test Files**:
- `test/models_test.go`: Tests for deterministic hashing (Artifact, State)
- `test/storage_test.go`: Tests for storage round-trip operations

**Run Tests**:
```bash
go test ./test/...
```

## Dependencies

**Go Modules** (`go.mod`):
- `github.com/dgraph-io/badger/v4`: Embedded key-value store
- `github.com/spf13/cobra`: CLI framework

**Python** (`python/setup.cfg`):
- Standard library only (subprocess, json, os, etc.)

## Current Implementation Status

### ✅ Completed (Phase 1)

- [x] Core data models (Artifact, State, Event) with canonical hashing
- [x] BadgerDB content-addressable storage
- [x] Embedded index (heads, time-based)
- [x] Recorder with directory snapshot and command execution
- [x] CLI commands (init, record-cmd, log, show, timeline, facts, query, rules)
- [x] Python client wrapper
- [x] Datalog facts emission
- [x] Basic unit tests

### 🚧 In Progress / Planned

- [ ] Full Datalog query engine integration (currently placeholder)
- [ ] Graph database integration (Dgraph/PostgreSQL)
- [ ] gRPC daemon for Python client
- [ ] Additional language clients
- [ ] mTLS and authentication
- [ ] Performance optimizations

## Key Algorithms

### 1. Content Addressing

Every object is identified by its content hash:
- **Artifact**: `SHA256(ContentType || 0x00 || Data)`
- **State**: `SHA256("State" || canonicalized_map_entries)`
- **Event**: `SHA256(JSON(event_without_ID))`

### 2. Canonical Map Serialization

Maps are serialized deterministically:
1. Sort keys alphabetically
2. For each key: `key + "\x00" + value`
3. Concatenate with null separators
4. Hash the result

### 3. Time Indexing

Uses reversed timestamps for efficient range queries:
- `reverse_ts = max_int64 - UnixNano`
- Keys: `time:<agent>:<reverse_ts>:<hash>`
- Lexicographic order = reverse chronological order

### 4. Merkle DAG Validation

Any change to an object invalidates all descendants:
- Event hash depends on parent hash
- Event hash depends on state hashes
- State hash depends on artifact hashes
- Artifact hash depends on content

## Usage Examples

### Go CLI

```bash
# Initialize
gab init --data ./gab_data

# Record a command
gab record-cmd --agent demo --scope . --cwd . -- echo "hello"

# View history
gab log --agent demo

# Show event details
gab show <event_hash>

# Timeline view
gab timeline --agent demo --limit 10

# View facts
gab facts
```

### Python Client

```python
from gab import init, exec_cmd

# Initialize
init(data_dir="./gab_data")

# Execute and record
event = exec_cmd(
    ["echo", "hello"],
    agent_id="demo",
    scope_dir=".",
    cwd="."
)

print(f"Event ID: {event['id']}")
print(f"Command: {event['action']['params']['command']}")
```

## Design Principles

1. **Immutability**: All objects are content-addressed and immutable
2. **Tamper-Evidence**: Cryptographic hashing ensures any modification is detectable
3. **Append-Only**: History is append-only; no deletion or modification
4. **Deduction-Ready**: Facts are emitted for deductive reasoning
5. **Language-Agnostic**: Core in Go, clients in any language
6. **Embedded First**: Uses embedded databases (BadgerDB) for simplicity

## Future Enhancements

- Full Datalog engine integration (Mangle/Cozo)
- Graph database backend for complex queries
- gRPC daemon for better performance
- Network event tracking
- File system watch integration
- Web UI for visualization
- Export/import capabilities
- Multi-agent correlation

