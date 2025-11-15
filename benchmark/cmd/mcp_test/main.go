package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/imran31415/godemode/benchmark/mcp/client"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/systems/email"
	"github.com/imran31415/godemode/benchmark/systems/filesystem"
	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/imran31415/godemode/benchmark/systems/security"
	"github.com/imran31415/godemode/pkg/env"
)

func main() {
	// Load .env file
	if err := env.LoadDefault(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// Check API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		fmt.Println("✓ ANTHROPIC_API_KEY loaded from environment")
	} else {
		fmt.Println("❌ No ANTHROPIC_API_KEY found")
		os.Exit(1)
	}

	fmt.Println("\nGoDeMode MCP Benchmark: 50-Step Security Incident Response")
	fmt.Println("===========================================================\n")

	// Get fixtures path
	fixturesPath := getFixturesPath()
	fmt.Printf("Using fixtures from: %s\n", fixturesPath)

	// Setup test environment
	testEnv, err := setupTestEnvironment(fixturesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup environment: %v\n", err)
		os.Exit(1)
	}

	// Create security scenario
	scenario := scenarios.NewSecurityScenario()
	task := scenario.Tasks[0] // The 50-step security incident task

	fmt.Printf("\n=== Running Task: %s ===\n", task.Name)
	fmt.Printf("Description: %s\n", task.Description)
	fmt.Printf("Expected Operations: %d\n\n", task.ExpectedOps)

	// Setup the task
	if task.SetupFunc != nil {
		fmt.Println("Setting up test environment...")
		if err := task.SetupFunc(testEnv); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to setup task: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Environment setup complete.\n")
	}

	// Build MCP server first
	fmt.Println("Building MCP server...")
	if err := buildMCPServer(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build MCP server: %v\n", err)
		os.Exit(1)
	}

	// Create MCP client and connect to server
	fmt.Println("\n--- Starting MCP Agent ---\n")
	fmt.Println("Launching MCP server...")

	mcpClient, err := client.NewMCPClient("./godemode-mcp-server")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create MCP client: %v\n", err)
		os.Exit(1)
	}
	defer mcpClient.Close()

	// Give server time to start
	time.Sleep(500 * time.Millisecond)

	// Initialize MCP connection
	if err := mcpClient.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize MCP: %v\n", err)
		os.Exit(1)
	}

	// Create MCP agent
	mcpAgent, err := client.NewMCPAgent(mcpClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create MCP agent: %v\n", err)
		os.Exit(1)
	}

	// Run the task via MCP
	ctx := context.Background()
	metrics, err := mcpAgent.RunTask(ctx, task, testEnv)
	if err != nil {
		fmt.Printf("MCP Agent failed: %v\n", err)
	}

	fmt.Printf("\nMCP Agent completed %d operations\n\n", metrics.OperationsCount)

	// Verify
	verified := false
	verifyErrors := []string{}
	if task.VerificationFunc != nil {
		verified, verifyErrors = task.VerificationFunc(testEnv)
		fmt.Println("MCP Agent Verification:")
		for _, err := range verifyErrors {
			fmt.Printf("  %s\n", err)
		}
		fmt.Println()
	}

	// Print results
	fmt.Println("\n" + repeatString("=", 100))
	fmt.Println("MCP AGENT RESULTS: 50-Step Security Incident Response")
	fmt.Println(repeatString("=", 100))

	fmt.Printf("\nMCP AGENT (Tools via Model Context Protocol):\n")
	fmt.Printf("  Duration: %v\n", metrics.TotalDuration)
	fmt.Printf("  Tokens: %d\n", metrics.TokensUsed)
	fmt.Printf("  API Calls: %d\n", metrics.APICallCount)
	fmt.Printf("  Operations: %d\n", metrics.OperationsCount)
	fmt.Printf("  Success: %v\n", metrics.Success)
	fmt.Printf("  Verified: %v\n\n", verified)

	if verified {
		fmt.Println("✅ ALL VERIFICATION CHECKS PASSED!")
	} else {
		fmt.Println("❌ Some verification checks failed - see details above")
	}

	fmt.Println("\nBenchmark completed!")
}

func buildMCPServer() error {
	// Check if already built and recent
	if _, err := os.Stat("./godemode-mcp-server"); err == nil {
		return nil
	}

	fmt.Println("Building MCP server binary...")
	return nil // Server will be built by go build command
}

func setupTestEnvironment(fixturesPath string) (*scenarios.TestEnvironment, error) {
	emailSystem := email.NewEmailSystem(fixturesPath+"/emails", fixturesPath+"/emails/sent")

	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	graphDB, err := graph.NewKnowledgeGraph(fixturesPath + "/graph_data")
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

func getFixturesPath() string {
	if path := os.Getenv("FIXTURES_PATH"); path != "" {
		return path
	}

	wd, err := os.Getwd()
	if err != nil {
		return "benchmark/fixtures"
	}

	if filepath.Base(wd) == "mcp_test" {
		return filepath.Join(wd, "..", "..", "..", "benchmark", "fixtures")
	}

	if filepath.Base(wd) == "cmd" {
		return filepath.Join(wd, "..", "..", "benchmark", "fixtures")
	}

	if filepath.Base(wd) == "benchmark" {
		return filepath.Join(wd, "fixtures")
	}

	return filepath.Join(wd, "benchmark", "fixtures")
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
