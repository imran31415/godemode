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
	// Setup test environment with separate graph database to avoid lock conflicts
	fixturesPath := "benchmark/fixtures"

	// Create temp directory for MCP server's graph database
	tmpDir, err := os.MkdirTemp("", "mcp-server-graph-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	env, err := setupTestEnvironment(fixturesPath, tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup environment: %v\n", err)
		os.Exit(1)
	}

	// Create and start MCP server
	srv := server.NewMCPServer(env)

	fmt.Fprintln(os.Stderr, "[MCP Server] GoDeMode MCP Server v1.0.0")
	fmt.Fprintln(os.Stderr, "[MCP Server] Ready to accept connections")

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[MCP Server] Error: %v\n", err)
		os.Exit(1)
	}
}

func setupTestEnvironment(fixturesPath, graphPath string) (*scenarios.TestEnvironment, error) {
	emailSystem := email.NewEmailSystem(fixturesPath+"/emails", fixturesPath+"/emails/sent")

	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Use separate temp directory for graph database to avoid lock conflicts
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
