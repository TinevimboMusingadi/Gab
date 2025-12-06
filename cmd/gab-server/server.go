package main

import (
	"gab/internal/index"
	"gab/internal/query"
	"gab/internal/recorder"
	"gab/internal/storage"
)

// Server holds all server dependencies
type Server struct {
	store      *storage.Store
	idx        *index.Index
	queryEngine *query.Engine
	diffEngine  *query.DiffEngine
	recorder    *recorder.Recorder
	config      *Config
}

// NewServer creates a new server instance
func NewServer(
	store *storage.Store,
	idx *index.Index,
	qEngine *query.Engine,
	diffEngine *query.DiffEngine,
	rec *recorder.Recorder,
	config *Config,
) *Server {
	return &Server{
		store:       store,
		idx:         idx,
		queryEngine: qEngine,
		diffEngine:  diffEngine,
		recorder:    rec,
		config:      config,
	}
}

