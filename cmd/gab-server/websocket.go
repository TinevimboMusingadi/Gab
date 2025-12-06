package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"gab/internal/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// WebSocketServer handles WebSocket connections for real-time event streaming
type WebSocketServer struct {
	server     *Server
	config     *Config
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]bool
	clientsMu   sync.RWMutex
	broadcast   chan *EventMessage
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
}

// EventMessage represents a message to be broadcast
type EventMessage struct {
	Type    string       `json:"type"`
	AgentID string       `json:"agent_id,omitempty"`
	Event   *models.Event `json:"event,omitempty"`
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(server *Server, config *Config) *WebSocketServer {
	return &WebSocketServer{
		server:    server,
		config:    config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan *EventMessage, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Start starts the WebSocket server
func (ws *WebSocketServer) Start() error {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/stream/agents/{agent_id}/events", ws.handleStream)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", ws.config.WSPort),
		Handler: router,
	}

	// Start hub goroutine
	go ws.runHub()

	return httpServer.ListenAndServe()
}

// Stop stops the WebSocket server
func (ws *WebSocketServer) Stop() {
	ws.clientsMu.Lock()
	for client := range ws.clients {
		client.Close()
		delete(ws.clients, client)
	}
	ws.clientsMu.Unlock()
	close(ws.broadcast)
	close(ws.register)
	close(ws.unregister)
}

// handleStream handles a WebSocket connection for event streaming
func (ws *WebSocketServer) handleStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	ws.register <- conn

	// Send initial message
	msg := EventMessage{
		Type:    "connected",
		AgentID: agentID,
	}
	conn.WriteJSON(msg)

	// Keep connection alive and handle messages
	go func() {
		defer func() {
			ws.unregister <- conn
			conn.Close()
		}()

		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			// Handle client messages if needed
			// For now, just keep connection alive
		}
	}()
}

// runHub manages client connections and broadcasts messages
func (ws *WebSocketServer) runHub() {
	for {
		select {
		case conn := <-ws.register:
			ws.clientsMu.Lock()
			ws.clients[conn] = true
			ws.clientsMu.Unlock()

		case conn := <-ws.unregister:
			ws.clientsMu.Lock()
			if _, ok := ws.clients[conn]; ok {
				delete(ws.clients, conn)
				conn.Close()
			}
			ws.clientsMu.Unlock()

		case message := <-ws.broadcast:
			ws.clientsMu.RLock()
			for client := range ws.clients {
				if err := client.WriteJSON(message); err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					delete(ws.clients, client)
				}
			}
			ws.clientsMu.RUnlock()
		}
	}
}

// BroadcastEvent broadcasts an event to all connected clients
func (ws *WebSocketServer) BroadcastEvent(agentID string, event *models.Event) {
	msg := EventMessage{
		Type:    "event",
		AgentID: agentID,
		Event:   event,
	}

	select {
	case ws.broadcast <- &msg:
	default:
		log.Println("Broadcast channel full, dropping message")
	}
}

// GetClientCount returns the number of connected clients
func (ws *WebSocketServer) GetClientCount() int {
	ws.clientsMu.RLock()
	defer ws.clientsMu.RUnlock()
	return len(ws.clients)
}

