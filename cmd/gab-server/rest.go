package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gab/internal/models"
	"gab/internal/query"

	"github.com/gorilla/mux"
)

// SetupRESTRoutes sets up all REST API routes
func SetupRESTRoutes(router *mux.Router, server *Server, config *Config) {
	api := router.PathPrefix("/api/v1").Subrouter()

	// CORS middleware
	if config.EnableCORS {
		api.Use(corsMiddleware)
	}

	// Agents
	api.HandleFunc("/agents", server.listAgents).Methods("GET")
	api.HandleFunc("/agents/{agent_id}/events", server.listAgentEvents).Methods("GET")

	// Events
	api.HandleFunc("/events/{event_id}", server.getEvent).Methods("GET")
	api.HandleFunc("/events/{event_id}/chain", server.getEventChain).Methods("GET")

	// States
	api.HandleFunc("/states/{pre_id}/diff/{post_id}", server.getStateDiff).Methods("GET")

	// Query
	api.HandleFunc("/query", server.queryEvents).Methods("POST")

	// Artifacts
	api.HandleFunc("/artifacts/{artifact_id}", server.getArtifact).Methods("GET")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.queryEngine.GetAllAgents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"agents": agents,
	})
}

func (s *Server) listAgentEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	chain, err := s.queryEngine.GetEventChain(agentID, limit+offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	events := chain.Events
	if offset < len(events) {
		if offset+limit > len(events) {
			events = events[offset:]
		} else {
			events = events[offset : offset+limit]
		}
	} else {
		events = []*models.Event{}
	}

	jsonResponse(w, map[string]interface{}{
		"events": events,
		"total":  chain.Count,
	})
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["event_id"]

	var ev models.Event
	if err := s.store.GetObject(eventID, &ev); err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, ev)
}

func (s *Server) getEventChain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	chain, err := s.queryEngine.GetEventChain(agentID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"events": chain.Events,
		"count":  chain.Count,
	})
}

func (s *Server) getStateDiff(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	preID := vars["pre_id"]
	postID := vars["post_id"]

	diff, err := s.diffEngine.GetStateDiff(preID, postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to API format
	created := make([]map[string]interface{}, len(diff.Created))
	for i, c := range diff.Created {
		created[i] = map[string]interface{}{
			"path":            c.Path,
			"post_artifact_id": c.PostArtifactID,
		}
	}

	modified := make([]map[string]interface{}, len(diff.Modified))
	for i, m := range diff.Modified {
		modified[i] = map[string]interface{}{
			"path":            m.Path,
			"pre_artifact_id":  m.PreArtifactID,
			"post_artifact_id": m.PostArtifactID,
		}
	}

	deleted := make([]map[string]interface{}, len(diff.Deleted))
	for i, d := range diff.Deleted {
		deleted[i] = map[string]interface{}{
			"path":           d.Path,
			"pre_artifact_id": d.PreArtifactID,
		}
	}

	jsonResponse(w, map[string]interface{}{
		"created":  created,
		"modified": modified,
		"deleted":  deleted,
	})
}

func (s *Server) queryEvents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentID    string `json:"agent_id"`
		ActionType string `json:"action_type"`
		Tool       string `json:"tool"`
		StartTime  int64  `json:"start_time"`
		EndTime    int64  `json:"end_time"`
		Limit      int    `json:"limit"`
		Offset     int    `json:"offset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	q := s.queryEngine.NewQuery()
	if req.AgentID != "" {
		q.Agent(req.AgentID)
	}
	if req.ActionType != "" {
		q.ActionType(req.ActionType)
	}
	if req.Tool != "" {
		q.Tool(req.Tool)
	}
	if req.StartTime > 0 && req.EndTime > 0 {
		q.TimeRange(time.Unix(req.StartTime, 0), time.Unix(req.EndTime, 0))
	}
	if req.Limit > 0 {
		q.Limit(req.Limit)
	}
	if req.Offset > 0 {
		q.Offset(req.Offset)
	}

	events, err := q.Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	total, _ := q.Count()

	jsonResponse(w, map[string]interface{}{
		"events": events,
		"total":  total,
	})
}

func (s *Server) getArtifact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	artifactID := vars["artifact_id"]

	var art models.Artifact
	if err := s.store.GetObject(artifactID, &art); err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, art)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

