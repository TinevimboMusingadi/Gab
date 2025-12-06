package query

import (
	"time"

	"gab/internal/models"
)

// Query represents a query with filters and options
type Query struct {
	engine     *Engine
	agentID    string
	limit      int
	offset     int
	startTime  *time.Time
	endTime    *time.Time
	actionType string
	tool       string
	sortBy     string // "time" or "chain"
}

// NewQuery creates a new query builder
func (e *Engine) NewQuery() *Query {
	return &Query{
		engine: e,
		limit:  100, // default limit
	}
}

// Agent filters by agent ID
func (q *Query) Agent(agentID string) *Query {
	q.agentID = agentID
	return q
}

// Limit sets the maximum number of results
func (q *Query) Limit(limit int) *Query {
	q.limit = limit
	return q
}

// Offset sets the offset for pagination
func (q *Query) Offset(offset int) *Query {
	q.offset = offset
	return q
}

// TimeRange filters events by time range
func (q *Query) TimeRange(start, end time.Time) *Query {
	q.startTime = &start
	q.endTime = &end
	return q
}

// ActionType filters by action type
func (q *Query) ActionType(actionType string) *Query {
	q.actionType = actionType
	return q
}

// Tool filters by tool name
func (q *Query) Tool(tool string) *Query {
	q.tool = tool
	return q
}

// SortBy sets the sort order
func (q *Query) SortBy(sortBy string) *Query {
	q.sortBy = sortBy
	return q
}

// Execute executes the query and returns matching events
func (q *Query) Execute() ([]*models.Event, error) {
	if q.agentID == "" {
		return nil, ErrAgentIDRequired
	}

	var events []*models.Event
	var err error

	// Get base events
	if q.startTime != nil && q.endTime != nil {
		events, err = q.engine.GetEventsByTimeRange(q.agentID, *q.startTime, *q.endTime)
	} else {
		chain, err := q.engine.GetEventChain(q.agentID, q.limit+q.offset)
		if err != nil {
			return nil, err
		}
		events = chain.Events
	}

	if err != nil {
		return nil, err
	}

	// Apply filters
	filtered := make([]*models.Event, 0)
	for _, ev := range events {
		if q.actionType != "" && ev.Action.Type != q.actionType {
			continue
		}
		if q.tool != "" && ev.Action.Tool != q.tool {
			continue
		}
		filtered = append(filtered, ev)
	}

	// Apply pagination
	start := q.offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + q.limit
	if end > len(filtered) {
		end = len(filtered)
	}

	if start >= len(filtered) {
		return []*models.Event{}, nil
	}

	return filtered[start:end], nil
}

// Count returns the number of matching events without fetching them all
func (q *Query) Count() (int, error) {
	events, err := q.Execute()
	if err != nil {
		return 0, err
	}
	return len(events), nil
}

var (
	ErrAgentIDRequired = &QueryError{Message: "agent ID is required"}
)

// QueryError represents a query error
type QueryError struct {
	Message string
}

func (e *QueryError) Error() string {
	return e.Message
}
