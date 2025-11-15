package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/imran31415/godemode/benchmark/agents"
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

	fmt.Println("Simple Task Benchmark: Native Tool Calling Test\n")
	fmt.Println("=================================================\n")

	// Setup environment
	fixturesPath := "benchmark/fixtures"
	testEnv, err := setupTestEnvironment(fixturesPath)
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

	// Create native tool calling agent
	agent := agents.NewNativeToolCallingAgent(testEnv)

	fmt.Printf("Task: %s\n", task.Description)
	fmt.Printf("Expected operations: %d\n\n", task.ExpectedOps)

	// Run the agent
	ctx := context.Background()
	metrics, err := agent.RunTask(ctx, task, testEnv)
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
		fmt.Println("\n✅ Native tool calling works for simple tasks!")
	} else {
		fmt.Println("\n❌ Test failed")
	}
}

func setupTestEnvironment(fixturesPath string) (*scenarios.TestEnvironment, error) {
	emailSystem := email.NewEmailSystem(fixturesPath+"/emails", fixturesPath+"/emails/sent")
	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		return nil, err
	}
	graphDB, err := graph.NewKnowledgeGraph(fixturesPath + "/graph_data")
	if err != nil {
		return nil, err
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
