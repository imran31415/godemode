package main

import (
	"fmt"
	"os"

	"github.com/imran31415/godemode/benchmark/mcp/server"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/systems/email"
	"github.com/imran31415/godemode/benchmark/systems/filesystem"
	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/imran31415/godemode/benchmark/systems/security"
)

func main() {
	// Setup test environment with shared paths from environment
	fixturesPath := "benchmark/fixtures"

	// Check for shared database paths from environment (set by test client)
	dbPath := os.Getenv("TEST_DB_PATH")
	graphPath := os.Getenv("TEST_GRAPH_PATH")

	// If not set, create temporary paths (for standalone server)
	if dbPath == "" || graphPath == "" {
		tmpDir, err := os.MkdirTemp("", "mcp-http-server-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create temp directory: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tmpDir)
		dbPath = ":memory:"
		graphPath = tmpDir
	}

	env, err := setupTestEnvironment(fixturesPath, dbPath, graphPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup environment: %v\n", err)
		os.Exit(1)
	}

	// Create and start HTTP MCP server on port 8080
	srv := server.NewHTTPMCPServer(env, 8080)

	fmt.Println("[MCP HTTP Server] GoDeMode MCP HTTP Server v1.0.0")
	fmt.Println("[MCP HTTP Server] Starting on http://localhost:8080")

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[MCP HTTP Server] Error: %v\n", err)
		os.Exit(1)
	}
}

func setupTestEnvironment(fixturesPath, dbPath, graphPath string) (*scenarios.TestEnvironment, error) {
	emailSystem := email.NewEmailSystem(fixturesPath+"/emails", fixturesPath+"/emails/sent")

	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Use shared graph database
	graphDB, err := graph.NewKnowledgeGraph(graphPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge graph: %w", err)
	}

	logSystem := filesystem.NewLogSystem(fixturesPath + "/logs")
	configSystem := filesystem.NewConfigSystem(fixturesPath + "/configs")
	securityMonitor := security.NewSecurityMonitor()

	return &scenarios.TestEnvironment{
		EmailSystem:     emailSystem,
		Database:        db,
		Graph:           graphDB,
		LogSystem:       logSystem,
		ConfigSystem:    configSystem,
		SecurityMonitor: securityMonitor,
	}, nil
}
