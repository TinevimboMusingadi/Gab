package query

import (
	"time"

	"gab/internal/index"
	"gab/internal/models"
	"gab/internal/storage"

	badger "github.com/dgraph-io/badger/v4"
)

// Engine provides query capabilities for the Merkle DAG structure
type Engine struct {
	store *storage.Store
	idx   *index.Index
}

// New creates a new query engine
func New(store *storage.Store, idx *index.Index) *Engine {
	return &Engine{
		store: store,
		idx:   idx,
	}
}

// EventChain represents a chain of events (lineage)
type EventChain struct {
	Events []*models.Event
	Count  int
}

// GetEventChain retrieves the event lineage for an agent, starting from the head
// and traversing backwards through parent events
func (e *Engine) GetEventChain(agentID string, limit int) (*EventChain, error) {
	head, err := e.idx.GetHead(agentID)
	if err != nil {
		return nil, err
	}
	if head == "" {
		return &EventChain{Events: []*models.Event{}, Count: 0}, nil
	}

	events := make([]*models.Event, 0)
	current := head
	visited := make(map[string]bool) // Prevent cycles

	for current != "" && len(events) < limit {
		if visited[current] {
			break // Cycle detected
		}
		visited[current] = true

		var ev models.Event
		if err := e.store.GetObject(current, &ev); err != nil {
			return nil, err
		}
		events = append(events, &ev)
		current = ev.ParentEventHash
	}

	return &EventChain{Events: events, Count: len(events)}, nil
}

// GetEventsByTimeRange retrieves events for an agent within a time range
func (e *Engine) GetEventsByTimeRange(agentID string, start, end time.Time) ([]*models.Event, error) {
	// Get all time entries for the agent
	entries, err := e.idx.ListByTime(agentID, 0) // 0 = no limit
	if err != nil {
		return nil, err
	}

	var events []*models.Event
	for _, entry := range entries {
		if entry.Time.After(start) && entry.Time.Before(end) {
			var ev models.Event
			if err := e.store.GetObject(entry.EventHash, &ev); err != nil {
				continue // Skip if event not found
			}
			events = append(events, &ev)
		}
	}

	return events, nil
}

// GetAncestors retrieves ancestor events up to a specified depth
func (e *Engine) GetAncestors(eventID string, depth int) ([]*models.Event, error) {
	var ev models.Event
	if err := e.store.GetObject(eventID, &ev); err != nil {
		return nil, err
	}

	ancestors := make([]*models.Event, 0)
	current := ev.ParentEventHash
	visited := make(map[string]bool)
	currentDepth := 0

	for current != "" && currentDepth < depth {
		if visited[current] {
			break // Cycle detected
		}
		visited[current] = true

		var parent models.Event
		if err := e.store.GetObject(current, &parent); err != nil {
			break // Parent not found, stop traversal
		}
		ancestors = append(ancestors, &parent)
		current = parent.ParentEventHash
		currentDepth++
	}

	return ancestors, nil
}

// FindEventsByArtifact finds all events that reference a specific artifact
// This is a reverse lookup - we need to scan events
func (e *Engine) FindEventsByArtifact(artifactID string) ([]*models.Event, error) {
	// This is expensive - we need to iterate through all events
	// For now, we'll use a simple approach: get all agents and scan their events
	// In production, we'd want an artifact->events index

	// Get all agents by scanning index keys
	agents, err := e.getAllAgents()
	if err != nil {
		return nil, err
	}

	var matchingEvents []*models.Event
	for _, agentID := range agents {
		chain, err := e.GetEventChain(agentID, 1000) // Large limit
		if err != nil {
			continue
		}

		for _, ev := range chain.Events {
			// Check if artifact is in resulting artifacts
			for _, artID := range ev.ResultingArtifactHashes {
				if artID == artifactID {
					matchingEvents = append(matchingEvents, ev)
					break
				}
			}

			// Check pre/post states
			if ev.PreActionStateHash != "" {
				var preState models.State
				if err := e.store.GetObject(ev.PreActionStateHash, &preState); err == nil {
					for _, artID := range preState.Resources {
						if artID == artifactID {
							matchingEvents = append(matchingEvents, ev)
							break
						}
					}
				}
			}

			if ev.PostActionStateHash != "" {
				var postState models.State
				if err := e.store.GetObject(ev.PostActionStateHash, &postState); err == nil {
					for _, artID := range postState.Resources {
						if artID == artifactID {
							matchingEvents = append(matchingEvents, ev)
							break
						}
					}
				}
			}
		}
	}

	return matchingEvents, nil
}

// GetAllAgents retrieves all agent IDs from the index
func (e *Engine) GetAllAgents() ([]string, error) {
	agents := make(map[string]bool)
	err := e.idx.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: []byte("head:")})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			key := string(it.Item().Key())
			// key format: head:<agentID>
			if len(key) > 5 {
				agentID := key[5:] // Skip "head:"
				agents[agentID] = true
			}
		}
		return nil
	})

	agentList := make([]string, 0, len(agents))
	for agentID := range agents {
		agentList = append(agentList, agentID)
	}
	return agentList, err
}
