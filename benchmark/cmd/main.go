package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imran31415/godemode/benchmark/runner"
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

	fmt.Println("\nGoDeMode Benchmark: Code Mode vs Function Calling")
	fmt.Println("==================================================\n")

	// Get fixtures path
	fixturesPath := getFixturesPath()
	fmt.Printf("Using fixtures from: %s\n", fixturesPath)

	// Create benchmark runner
	benchRunner, err := runner.NewBenchmarkRunner(fixturesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create benchmark runner: %v\n", err)
		os.Exit(1)
	}

	// Run benchmark
	ctx := context.Background()
	report, err := benchRunner.RunBenchmark(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		os.Exit(1)
	}

	// Print report
	report.PrintReport()

	fmt.Println("\nBenchmark completed successfully!")
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
