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
	if apiKey == "" {
		fmt.Println("❌ No ANTHROPIC_API_KEY found")
		os.Exit(1)
	}

	fmt.Println("Simple Task Benchmark: MCP HTTP Test")
	fmt.Println("=====================================\n")

	// Setup environment
	fixturesPath := "benchmark/fixtures"

	// Create temp directory for shared state between server and client
	tmpDir, err := os.MkdirTemp("", "mcp-http-simple-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Create shared database file path (only SQL DB, not graph - BadgerDB doesn't support multi-process)
	dbPath := filepath.Join(tmpDir, "test.db")
	serverGraphPath := filepath.Join(tmpDir, "server_graph")
	clientGraphPath := filepath.Join(tmpDir, "client_graph")

	// Pass shared paths via environment variables (server gets its own graph path)
	os.Setenv("TEST_DB_PATH", dbPath)
	os.Setenv("TEST_GRAPH_PATH", serverGraphPath)

	testEnv, err := setupTestEnvironment(fixturesPath, dbPath, clientGraphPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup environment: %v\n", err)
		os.Exit(1)
	}

	// Store email ID for the task description
	emailID := "test_email_001"

	// Create a simple 3-step task
	task := scenarios.Task{
		Name:        "simple-email-response",
		Description: "", // Will be set after setup
		Complexity:  "simple",
		ExpectedOps: 3,
		SetupFunc: func(env *scenarios.TestEnvironment) error {
			// Create email file manually in inbox
			emailContent := `From: user@example.com
To: support@company.com
Subject: Bug Report
Date: ` + fmt.Sprintf("%s", time.Now().Format(time.RFC1123Z)) + `

Login button is not working`

			// Write directly to inbox so readEmail can find it
			inboxPath := fixturesPath + "/emails"
			emailFile := filepath.Join(inboxPath, emailID+".eml")
			return os.WriteFile(emailFile, []byte(emailContent), 0644)
		},
		VerificationFunc: func(env *scenarios.TestEnvironment) (bool, []string) {
			errors := []string{}

			// Check if ticket was created
			tickets, err := env.Database.QueryTickets(map[string]interface{}{})
			if err != nil || len(tickets) == 0 {
				errors = append(errors, "❌ No ticket created")
				return false, errors
			}

			errors = append(errors, "✅ Ticket created successfully")

			// Check if confirmation email was sent
			sent, err := env.EmailSystem.ListSentEmails()
			if err != nil || len(sent) < 2 { // original + confirmation
				errors = append(errors, "❌ No confirmation email sent")
				return false, errors
			}

			errors = append(errors, "✅ Confirmation email sent")

			return true, errors
		},
	}

	// Setup task
	if task.SetupFunc != nil {
		task.SetupFunc(testEnv)
	}

	// Set task description with the actual email ID
	task.Description = fmt.Sprintf("Read email ID '%s', create a support ticket for it, and send a confirmation email", emailID)

	// Build and start MCP HTTP server in background
	fmt.Println("Building MCP HTTP server...")
	if err := exec.Command("go", "build", "-o", "godemode-mcp-http-server", "benchmark/cmd/mcp_http_server/main.go").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build MCP HTTP server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting MCP HTTP server...")
	serverCmd := exec.Command("./godemode-mcp-http-server")
	// Pass environment variables to server process
	serverCmd.Env = append(os.Environ(),
		fmt.Sprintf("TEST_DB_PATH=%s", dbPath),
		fmt.Sprintf("TEST_GRAPH_PATH=%s", serverGraphPath),
	)
	// Capture server output for debugging
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
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

	fmt.Printf("Task: %s\n", task.Description)
	fmt.Printf("Expected operations: %d\n\n", task.ExpectedOps)

	// Run the agent
	ctx := context.Background()
	metrics, err := mcpAgent.RunTask(ctx, task, testEnv)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Verify
	success, errors := task.VerificationFunc(testEnv)

	fmt.Println("\n=== Results ===")
	fmt.Printf("Duration: %v\n", metrics.TotalDuration)
	fmt.Printf("API Calls: %d\n", metrics.APICallCount)
	fmt.Printf("Tokens: %d\n", metrics.TokensUsed)
	fmt.Printf("Operations: %d\n", metrics.OperationsCount)
	fmt.Printf("Success: %v\n\n", success)

	fmt.Println("Verification:")
	for _, err := range errors {
		fmt.Printf("  %s\n", err)
	}

	if success {
		fmt.Println("\n✅ MCP HTTP works for simple tasks!")
	} else {
		fmt.Println("\n❌ Test failed")
		os.Exit(1)
	}
}

func setupTestEnvironment(fixturesPath, dbPath, graphPath string) (*scenarios.TestEnvironment, error) {
	emailSystem := email.NewEmailSystem(fixturesPath+"/emails", fixturesPath+"/emails/sent")

	// Use shared file-based database instead of :memory:
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
