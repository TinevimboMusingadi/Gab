package graph

import (
    "gab/internal/models"
)

// Sink is an interface for writing event relationships to a graph/deductive DB.
// Future implementations may target Dgraph, PostgreSQL, or other backends.
type Sink interface {
    WriteEvent(agentID string, ev *models.Event, pre *models.State, post *models.State) error
}

// NoopSink implements Sink but does nothing. Useful until a real backend is configured.
type NoopSink struct{}

func NewNoopSink() *NoopSink { return &NoopSink{} }

func (n *NoopSink) WriteEvent(agentID string, ev *models.Event, pre *models.State, post *models.State) error {
    return nil
}


