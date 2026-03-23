## Gab: Auditing and Versioned Activity Tracking for AI Agents

Gab is an append-only, tamper-evident activity tracking system for AI agents, inspired by Git. It records Events (actions) with pre/post State snapshots and Artifacts, forming a Merkle DAG for complete forensic traceability. This repository implements Phase 1 (core foundation) in Go, with a CLI and Badger-based content-addressable storage. Python integration and the deductive graph layer follow in later phases.

### Features

- **Merkle DAG**: Cryptographically-secured event chain where any retroactive change invalidates the entire descendant chain
- **Content-Addressable Storage**: All objects (Artifacts, States, Events) are identified by SHA-256 hashes
- **State Snapshots**: Pre/post action directory snapshots capture system context
- **Datalog Facts**: Automatic emission of facts for deductive reasoning and querying
- **Time-Based Indexing**: Efficient chronological event queries
- **Python Integration**: Easy-to-use Python client for agent integration for testingn

### Phases & Stages

- **Phase 1: Core Foundation (Go)** ✅
  - Models: Artifact, State, ActionDescriptor, Event with canonical hashing
  - Storage: BadgerDB content-addressable object store
  - Index: Embedded heads and time-based indices
  - Recorder: Snapshot directory scope, execute command, record Event
  - CLI: init, record-cmd, log, show, timeline, facts, query, rules
  - Python client: Subprocess wrapper for early integration
  - Tests: models and storage

- **Phase 2: Python Integration** 🚧
  - gRPC daemon for better performance
  - Enhanced Python client

- **Phase 3: Deductive Query Engine** 🚧
  - External graph DB integration (e.g., Dgraph or PostgreSQL)
  - Full Datalog query engine (Mangle/Cozo)

- **Phase 4: Expansion & Security** 📋
  - Additional language clients; mTLS and auth; performance optimizations

## Building and Installation

### Prerequisites

- Go 1.22 or later
- Python 3.7+ (for Python client)

### Build the CLI

```bash
# Clone the repository
git clone <repository-url>
cd Gab

# Build the gab CLI
go build -o gab.exe ./cmd/gab  # Windows
# or
go build -o gab ./cmd/gab       # Linux/macOS

# Add to PATH (optional)
# Windows: Add directory to PATH environment variable
# Linux/macOS: sudo mv gab /usr/local/bin/
```

### Install Python Client (Optional)

```bash
cd python
pip install -e .
# or
python setup.py install
```

### Run Tests

```bash
go test ./test/...
```

## Quickstart

### 1. Initialize the Repository

```bash
gab init --data ./gab_data
```

This creates the data directory structure:
- `gab_data/objects/` - Content-addressable object store (BadgerDB)
- `gab_data/index/` - Index database (BadgerDB)
- `gab_data/deduce/` - Datalog facts and rules

### 2. Record a Command

Record a command execution with state snapshots:

```bash
# Linux/macOS
gab record-cmd --agent demo --scope . --cwd . -- ls -l

# Windows
gab record-cmd --agent demo --scope . --cwd . -- dir
```

This will:
- Snapshot the directory before execution
- Execute the command
- Snapshot the directory after execution
- Create an Event linking pre/post states
- Store all artifacts and emit Datalog facts

### 3. View Event History

```bash
# View reverse chronological log
gab log --agent demo

# View timeline (chronological)
gab timeline --agent demo --limit 10
```

### 4. Inspect an Event

```bash
# Show full event JSON
gab show <event_hash>

# View emitted facts
gab facts
```

### 5. Install Default Rules

```bash
gab rules
```

## Python Client Usage

The Python client provides a simple interface for recording events from Python agents:

```python
from gab import init, exec_cmd  # or use 'exec' or 'run' as aliases

# Initialize Gab store
init(data_dir="./gab_data")

# Execute and record a command
event = exec_cmd(
    ["echo", "hello"],
    agent_id="demo",
    scope_dir=".",
    cwd="."
)

# Access event information
print(f"Event ID: {event['id']}")
print(f"Command: {event['action']['params']['command']}")
print(f"Finished at: {event['action']['finished_at_unix']}")
```

**Note**: The Python client requires the `gab` CLI to be on your PATH. Set `GAB_CLI` environment variable to override the path.

## CLI Commands

| Command | Description |
|---------|-------------|
| `gab init` | Initialize Gab data directory |
| `gab record-cmd` | Record command execution with state snapshots |
| `gab log` | Show event history for an agent (reverse chronological) |
| `gab show <hash>` | Display full JSON for an event |
| `gab timeline` | List recent events in time order |
| `gab facts` | Show path and preview of facts file |
| `gab query` | Placeholder Datalog query executor |
| `gab rules` | Install default starter rules |

All commands support `--data <dir>` flag to specify the data directory (default: `./gab_data`).

## Repository Layout

```
Gab/
├── cmd/gab/              # CLI entrypoint and commands
│   ├── main.go          # Root command
│   ├── init.go          # Initialize command
│   ├── record_cmd.go    # Record command execution
│   ├── log.go           # Event history log
│   ├── show.go          # Show event details
│   ├── timeline.go      # Time-ordered events
│   ├── facts.go         # View facts file
│   ├── query.go         # Datalog query (placeholder)
│   └── rules.go         # Install rules
├── internal/
│   ├── models/          # Core data types (Artifact, State, Event)
│   ├── hashutil/        # Canonicalization and hashing helpers
│   ├── storage/         # BadgerDB-based object store
│   ├── index/           # Embedded indices (heads, time-based)
│   ├── recorder/        # Snapshot/execute/record logic
│   ├── deduce/          # EDB facts writer, rules installer
│   └── graph/           # Graph sink abstraction
├── python/gab/          # Python client library
│   ├── __init__.py
│   └── client.py
├── api/proto/           # gRPC proto definitions (future)
├── test/                # Unit tests
│   ├── models_test.go
│   └── storage_test.go
├── docs/
│   ├── ARCHITECTURE.md  # Full architecture and design document
│   └── CODEBASE.md      # Codebase summary and implementation details
├── go.mod               # Go dependencies
└── README.md            # This file
```

## How It Works

### Recording an Event

1. **Pre-snapshot**: Walk the scope directory, create Artifacts for files, assemble a State
2. **Execute**: Run the command, capture stdout/stderr/exit_code as Artifacts
3. **Post-snapshot**: Repeat snapshot to generate new State
4. **Persist**: Create Event with parent pointer, pre/post states, action, result artifacts
5. **Index**: Update agent head and time index
6. **Emit Facts**: Append Datalog-style EDB facts for deductive reasoning

### Merkle DAG Structure

```
Artifact (file content) → State (file mapping) → Event (action)
     ↓                        ↓                      ↓
Artifact (stdout) ────────────┴──────────────────────┘
     ↓
Event links to parent Event, forming a chain
```

Any change to an object invalidates all descendant hashes, ensuring tamper-evidence.

## Documentation

- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)**: Full architecture and design document
- **[CODEBASE.md](docs/CODEBASE.md)**: Detailed codebase summary and implementation guide

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]


