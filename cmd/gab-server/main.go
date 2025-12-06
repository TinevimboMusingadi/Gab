package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gab/internal/index"
	"gab/internal/query"
	"gab/internal/recorder"
	"gab/internal/storage"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func main() {
	configPath := flag.String("config", "", "Path to config file (JSON)")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage and index
	store, err := storage.Open(config.DataDir)
	if err != nil {
		log.Fatalf("Failed to open storage: %v", err)
	}
	defer store.Close()

	idx, err := index.Open(config.DataDir)
	if err != nil {
		log.Fatalf("Failed to open index: %v", err)
	}
	defer idx.Close()

	// Initialize query engine
	qEngine := query.New(store, idx)
	diffEngine := query.NewDiffEngine(store)

	// Initialize recorder
	rec := recorder.New(store, idx, config.DataDir)

	// Create server instance
	server := NewServer(store, idx, qEngine, diffEngine, rec, config)

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	grpcServer := grpc.NewServer()
	// Register gRPC service (will be implemented in grpc.go)
	// RegisterGabService(grpcServer, server)

	go func() {
		log.Printf("gRPC server listening on :%d", config.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Start REST API server
	router := mux.NewRouter()
	SetupRESTRoutes(router, server, config)

	restServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.RESTPort),
		Handler: router,
	}

	go func() {
		log.Printf("REST API server listening on :%d", config.RESTPort)
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("REST server failed: %v", err)
		}
	}()

	// Start WebSocket server
	wsServer := NewWebSocketServer(server, config)
	go func() {
		log.Printf("WebSocket server listening on :%d", config.WSPort)
		if err := wsServer.Start(); err != nil {
			log.Fatalf("WebSocket server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	restServer.Shutdown(ctx)
	wsServer.Stop()

	log.Println("Servers stopped")
}

