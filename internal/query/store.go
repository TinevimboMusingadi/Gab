package query

import (
	"sync"

	"gab/internal/index"
	"gab/internal/storage"
)

// QueryStore wraps storage and index with caching for the query engine
type QueryStore struct {
	store *storage.Store
	idx   *index.Index
	cache *queryCache
}

// queryCache provides simple in-memory caching
type queryCache struct {
	mu      sync.RWMutex
	events  map[string]*cachedEvent
	maxSize int
}

type cachedEvent struct {
	event interface{}
}

// NewQueryStore creates a new query store with caching
func NewQueryStore(store *storage.Store, idx *index.Index) *QueryStore {
	return &QueryStore{
		store: store,
		idx:   idx,
		cache: &queryCache{
			events:  make(map[string]*cachedEvent),
			maxSize: 1000, // Cache up to 1000 events
		},
	}
}

// GetStore returns the underlying storage store
func (qs *QueryStore) GetStore() *storage.Store {
	return qs.store
}

// GetIndex returns the underlying index
func (qs *QueryStore) GetIndex() *index.Index {
	return qs.idx
}

// BatchLoadEvents loads multiple events efficiently
func (qs *QueryStore) BatchLoadEvents(eventIDs []string) ([]interface{}, error) {
	results := make([]interface{}, 0, len(eventIDs))

	// Check cache first
	uncached := make([]string, 0)
	qs.cache.mu.RLock()
	for _, id := range eventIDs {
		if cached, ok := qs.cache.events[id]; ok {
			results = append(results, cached.event)
		} else {
			uncached = append(uncached, id)
		}
	}
	qs.cache.mu.RUnlock()

	// Load uncached events
	if len(uncached) > 0 {
		// For now, load sequentially
		// In production, could batch load from BadgerDB
		// Note: This is a placeholder - actual implementation would load events
		_ = uncached
	}

	return results, nil
}

// ClearCache clears the query cache
func (qs *QueryStore) ClearCache() {
	qs.cache.mu.Lock()
	defer qs.cache.mu.Unlock()
	qs.cache.events = make(map[string]*cachedEvent)
}
