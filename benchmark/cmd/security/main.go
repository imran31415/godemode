package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imran31415/godemode/benchmark/agents"
	"github.com/imran31415/godemode/benchmark/runner"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/systems/email"
	"github.com/imran31415/godemode/benchmark/systems/filesystem"
	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/imran31415/godemode/benchmark/systems/security"
	"github.com/imran31415/godemode/pkg/env"
)

func main() {
	// Load .env file if it exists
	if err := env.LoadDefault(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// Check if API key is set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		fmt.Println("✓ ANTHROPIC_API_KEY loaded from environment")
	} else {
		fmt.Println("ℹ️  No ANTHROPIC_API_KEY found - using mock responses")
		fmt.Println("   To use real Claude API, set ANTHROPIC_API_KEY in .env file")
	}

	fmt.Println("\nGoDeMode Security Benchmark: 50-Step Incident Response")
	fmt.Println("=========================================================\n")

	// Get fixtures path
	fixturesPath := getFixturesPath()
	fmt.Printf("Using fixtures from: %s\n", fixturesPath)

	// Setup test environment
	env, err := setupTestEnvironment(fixturesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup environment: %v\n", err)
		os.Exit(1)
	}

	// Create security scenario
	scenario := scenarios.NewSecurityScenario()

	// Create agents - using NATIVE tool calling agent
	functionAgent := agents.NewNativeToolCallingAgent(env)

	// Run the security incident task
	ctx := context.Background()
	task := scenario.Tasks[0] // The 50-step security incident task

	fmt.Printf("\n=== Running Task: %s ===\n", task.Name)
	fmt.Printf("Description: %s\n", task.Description)
	fmt.Printf("Expected Operations: %d\n\n", task.ExpectedOps)

	// Setup the task
	if task.SetupFunc != nil {
		fmt.Println("Setting up test environment...")
		if err := task.SetupFunc(env); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to setup task: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Environment setup complete.\n")
	}

	// SKIP Code Mode for now - focus on Native Tool Calling
	fmt.Println("(Skipping Code Mode - focusing on Native Tool Calling Agent)\n")

	// Run Native Tool Calling Agent
	fmt.Println("--- Running NATIVE TOOL CALLING Agent ---\n")
	functionMetrics, err := functionAgent.RunTask(ctx, task, env)
	if err != nil {
		fmt.Printf("Function Calling failed: %v\n", err)
	}

	// Native Tool Calling metrics
	fmt.Printf("Native Tool Calling completed %d operations\n\n", functionMetrics.OperationsCount)

	// Verify Native Tool Calling
	functionVerified := false
	functionErrors := []string{}
	if task.VerificationFunc != nil {
		functionVerified, functionErrors = task.VerificationFunc(env)
		fmt.Println("Native Tool Calling Verification:")
		for _, err := range functionErrors {
			fmt.Printf("  %s\n", err)
		}
		fmt.Println()
	}

	// Print results
	fmt.Println("\n" + runner.RepeatString("=", 100))
	fmt.Println("NATIVE TOOL CALLING RESULTS: 50-Step Security Incident Response")
	fmt.Println(runner.RepeatString("=", 100))

	fmt.Printf("\nNATIVE TOOL CALLING (Claude's official tool use API):\n")
	fmt.Printf("  Duration: %v\n", functionMetrics.TotalDuration)
	fmt.Printf("  Tokens: %d\n", functionMetrics.TokensUsed)
	fmt.Printf("  API Calls: %d\n", functionMetrics.APICallCount)
	fmt.Printf("  Operations: %d\n", functionMetrics.OperationsCount)
	fmt.Printf("  Success: %v\n", functionMetrics.Success)
	fmt.Printf("  Verified: %v\n\n", functionVerified)

	if functionVerified {
		fmt.Println("✅ ALL VERIFICATION CHECKS PASSED!")
	} else {
		fmt.Println("❌ Some verification checks failed - see details above")
	}

	fmt.Println("\nBenchmark completed!")
}

// setupTestEnvironment initializes all systems
func setupTestEnvironment(fixturesPath string) (*scenarios.TestEnvironment, error) {
	// Email system
	emailSystem := email.NewEmailSystem(fixturesPath+"/emails", fixturesPath+"/emails/sent")

	// Database
	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Knowledge graph
	graphDB, err := graph.NewKnowledgeGraph(fixturesPath + "/graph_data")
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge graph: %w", err)
	}

	// Log system
	logSystem := filesystem.NewLogSystem(fixturesPath + "/logs")

	// Config system
	configSystem := filesystem.NewConfigSystem(fixturesPath + "/configs")

	// Security monitor
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

// getFixturesPath returns the path to the fixtures directory
func getFixturesPath() string {
	// Check if path provided via environment variable
	if path := os.Getenv("FIXTURES_PATH"); path != "" {
		return path
	}

	// Default to relative path from project root
	wd, err := os.Getwd()
	if err != nil {
		return "benchmark/fixtures"
	}

	// If we're in the security directory, go up three levels
	if filepath.Base(wd) == "security" {
		return filepath.Join(wd, "..", "..", "..", "benchmark", "fixtures")
	}

	// If we're in the cmd directory, go up two levels
	if filepath.Base(wd) == "cmd" {
		return filepath.Join(wd, "..", "..", "benchmark", "fixtures")
	}

	// If we're in the benchmark directory, fixtures is a sibling
	if filepath.Base(wd) == "benchmark" {
		return filepath.Join(wd, "fixtures")
	}

	// Default: assume we're at project root
	return filepath.Join(wd, "benchmark", "fixtures")
}
