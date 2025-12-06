# Query Engine Architecture

## Overview

The Query Engine provides a custom query interface for traversing and analyzing the Merkle DAG structure in Gab. It enables efficient retrieval of events, state comparisons, and graph traversal operations.

## Components

### 1. Query Engine Core (`internal/query/engine.go`)

The core engine provides fundamental Merkle DAG traversal operations:

#### Key Methods

- **GetEventChain(agentID, limit)**: Retrieves the event lineage for an agent, starting from the head and traversing backwards through parent events. Uses cycle detection to prevent infinite loops.

- **GetEventsByTimeRange(agentID, start, end)**: Retrieves all events for an agent within a specified time range. Uses the time index for efficient queries.

- **GetAncestors(eventID, depth)**: Traverses the parent chain from a given event up to a specified depth. Useful for understanding event context and history.

- **FindEventsByArtifact(artifactID)**: Performs a reverse lookup to find all events that reference a specific artifact. This includes events where the artifact appears in:
  - Resulting artifacts (stdout, stderr, exit_code)
  - Pre-action state resources
  - Post-action state resources

#### Algorithm Details

**Event Chain Traversal**:
```
1. Get agent head from index
2. Load event by hash
3. Follow ParentEventHash backwards
4. Stop when:
   - ParentEventHash is empty (root event)
   - Limit reached
   - Cycle detected (visited set)
```

**Time Range Queries**:
```
1. Get all time entries for agent (from index)
2. Filter entries within time range
3. Load events for matching entries
4. Return sorted list
```

### 2. Query Builder (`internal/query/builder.go`)

Provides a fluent API for building complex queries with filters and pagination.

#### Features

- **Fluent Interface**: Chain methods to build queries
- **Filtering**: By agent, action type, tool, time range
- **Pagination**: Limit and offset support
- **Sorting**: By time or chain order

#### Example Usage

```go
events, err := engine.NewQuery().
    Agent("demo").
    ActionType("EXECUTE_COMMAND").
    Tool("gab-cli").
    TimeRange(startTime, endTime).
    Limit(50).
    Offset(0).
    Execute()
```

### 3. State Diff Engine (`internal/query/diff.go`)

Computes and formats differences between pre and post action states.

#### Capabilities

- **File Changes**: Identifies created, modified, and deleted files
- **Content Comparison**: Loads artifact content for comparison
- **Human-Readable Output**: Formats diffs in a readable format
- **Unified Diff**: Generates unified diff format for modified files

#### Diff Algorithm

```
1. Load pre-state and post-state from storage
2. Compare resource maps:
   - Created: in post but not pre
   - Modified: in both but artifact IDs differ
   - Deleted: in pre but not post
3. Load artifact content for detailed comparison
4. Generate formatted report
```

### 4. Query Store (`internal/query/store.go`)

Wrapper around storage and index with caching support.

#### Features

- **Caching**: In-memory cache for frequently accessed events
- **Batch Loading**: Efficient loading of multiple events
- **Cache Management**: Clear cache when needed

## Integration Points

### Storage Layer

The query engine integrates with:
- `storage.Store`: For loading Artifacts, States, and Events
- `index.Index`: For agent heads and time-based queries

### Index Extensions

Added `View()` method to `index.Index` to allow custom BadgerDB queries for agent enumeration.

## Performance Considerations

1. **Caching**: Frequently accessed events are cached to reduce storage lookups
2. **Cycle Detection**: Prevents infinite loops in graph traversal
3. **Batch Operations**: Future optimization for loading multiple objects
4. **Index Usage**: Leverages time index for efficient range queries

## Future Enhancements

- Reverse index for artifact → events lookup
- Graph database integration for complex queries
- Parallel traversal for large event chains
- Incremental diff computation
- Advanced diff algorithms (Myers, patience)

## Usage Example

```go
// Initialize
store, _ := storage.Open(dataDir)
idx, _ := index.Open(dataDir)
engine := query.New(store, idx)

// Get event chain
chain, _ := engine.GetEventChain("demo", 100)

// Query with filters
events, _ := engine.NewQuery().
    Agent("demo").
    Limit(50).
    Execute()

// Get state diff
diffEngine := query.NewDiffEngine(store)
diff, _ := diffEngine.GetStateDiff(preStateID, postStateID)
fmt.Println(diff.FormatDiff())
```

