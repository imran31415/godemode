package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

	fmt.Println("\nGoDeMode MCP HTTP Benchmark: 50-Step Security Incident Response")
	fmt.Println("================================================================\n")

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

	// Build and start MCP HTTP server in background
	fmt.Println("Building MCP HTTP server...")
	if err := exec.Command("go", "build", "-o", "godemode-mcp-http-server", "benchmark/cmd/mcp_http_server/main.go").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build MCP HTTP server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting MCP HTTP server...")
	serverCmd := exec.Command("./godemode-mcp-http-server")
	if err := serverCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start MCP HTTP server: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
	}()

	// Wait for server to start
	fmt.Println("Waiting for server to start...")
	time.Sleep(2 * time.Second)

	// Create MCP HTTP client
	fmt.Println("\n--- Starting MCP HTTP Agent ---\n")
	mcpClient := client.NewHTTPMCPClient("http://localhost:8080")

	// Initialize MCP connection
	if err := mcpClient.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize MCP HTTP: %v\n", err)
		os.Exit(1)
	}

	// Create MCP HTTP agent
	mcpAgent, err := client.NewHTTPMCPAgent(mcpClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create MCP HTTP agent: %v\n", err)
		os.Exit(1)
	}

	// Run the task via MCP HTTP
	ctx := context.Background()
	metrics, err := mcpAgent.RunTask(ctx, task, testEnv)
	if err != nil {
		fmt.Printf("MCP HTTP Agent failed: %v\n", err)
	}

	fmt.Printf("\nMCP HTTP Agent completed %d operations\n\n", metrics.OperationsCount)

	// Verify
	verified := false
	verifyErrors := []string{}
	if task.VerificationFunc != nil {
		verified, verifyErrors = task.VerificationFunc(testEnv)
		fmt.Println("MCP HTTP Agent Verification:")
		for _, err := range verifyErrors {
			fmt.Printf("  %s\n", err)
		}
		fmt.Println()
	}

	// Print results
	fmt.Println("\n" + repeatString("=", 100))
	fmt.Println("MCP HTTP AGENT RESULTS: 50-Step Security Incident Response")
	fmt.Println(repeatString("=", 100))

	fmt.Printf("\nMCP HTTP AGENT (Tools via HTTP Model Context Protocol):\n")
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

	if filepath.Base(wd) == "mcp_http_test" {
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
