package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Config holds server configuration
type Config struct {
	DataDir    string `json:"data_dir"`
	GRPCPort   int    `json:"grpc_port"`
	RESTPort   int    `json:"rest_port"`
	WSPort     int    `json:"ws_port"`
	EnableCORS bool   `json:"enable_cors"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		DataDir:    "./gab_data",
		GRPCPort:   50051,
		RESTPort:   8080,
		WSPort:     8081,
		EnableCORS: true,
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Load from file if provided
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := json.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	if val := os.Getenv("GAB_DATA_DIR"); val != "" {
		config.DataDir = val
	}
	if val := os.Getenv("GAB_GRPC_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.GRPCPort = port
		}
	}
	if val := os.Getenv("GAB_REST_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.RESTPort = port
		}
	}
	if val := os.Getenv("GAB_WS_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.WSPort = port
		}
	}
	if val := os.Getenv("GAB_ENABLE_CORS"); val != "" {
		config.EnableCORS = val == "true" || val == "1"
	}

	return config, nil
}

