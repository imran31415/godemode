package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type BenchmarkComparison struct {
	Task            string
	NativeResult    *NativeMCPResult
	GoDeModeResult  *GoDeModeMCPResult
}

func (bc *BenchmarkComparison) PrintComparison() {
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("📊 REAL MCP BENCHMARK RESULTS")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	fmt.Println("📋 Task:")
	fmt.Println(bc.Task)
	fmt.Println()

	// Native MCP Results
	fmt.Println("🔵 Native MCP (Sequential Tool Calling)")
	fmt.Println(strings.Repeat("-", 80))
	if bc.NativeResult.Success {
		fmt.Printf("  ✅ Success\n")
		fmt.Printf("  Total Duration:    %v\n", bc.NativeResult.TotalDuration)
		fmt.Printf("  API Calls:         %d\n", bc.NativeResult.APICallCount)
		fmt.Printf("  Tool Calls:        %d\n", bc.NativeResult.ToolCallCount)
		fmt.Printf("  Input Tokens:      %d\n", bc.NativeResult.TotalInputTokens)
		fmt.Printf("  Output Tokens:     %d\n", bc.NativeResult.TotalOutputTokens)
		fmt.Printf("  Total Tokens:      %d\n", bc.NativeResult.TotalInputTokens+bc.NativeResult.TotalOutputTokens)
		fmt.Println()
		fmt.Println("  Output:")
		fmt.Println("  " + strings.ReplaceAll(bc.NativeResult.FinalOutput, "\n", "\n  "))
	} else {
		fmt.Printf("  ❌ Failed: %v\n", bc.NativeResult.Error)
	}
	fmt.Println()

	// GoDeMode MCP Results
	fmt.Println("🟢 GoDeMode MCP (Code Generation)")
	fmt.Println(strings.Repeat("-", 80))
	if bc.GoDeModeResult.Success {
		fmt.Printf("  ✅ Success\n")
		fmt.Printf("  Total Duration:      %v\n", bc.GoDeModeResult.TotalDuration)
		fmt.Printf("  Code Gen Duration:   %v\n", bc.GoDeModeResult.CodeGenDuration)
		fmt.Printf("  Execution Duration:  %v\n", bc.GoDeModeResult.ExecutionDuration)
		fmt.Printf("  API Calls:           %d\n", bc.GoDeModeResult.APICallCount)
		fmt.Printf("  Input Tokens:        %d\n", bc.GoDeModeResult.TotalInputTokens)
		fmt.Printf("  Output Tokens:       %d\n", bc.GoDeModeResult.TotalOutputTokens)
		fmt.Printf("  Total Tokens:        %d\n", bc.GoDeModeResult.TotalInputTokens+bc.GoDeModeResult.TotalOutputTokens)
		fmt.Println()
		fmt.Println("  Generated Code:")
		fmt.Println("  " + strings.ReplaceAll(bc.GoDeModeResult.GeneratedCode, "\n", "\n  "))
		fmt.Println()
		fmt.Println("  Output:")
		fmt.Println("  " + strings.ReplaceAll(bc.GoDeModeResult.FinalOutput, "\n", "\n  "))
	} else {
		fmt.Printf("  ❌ Failed: %v\n", bc.GoDeModeResult.Error)
	}
	fmt.Println()

	// Comparison
	if bc.NativeResult.Success && bc.GoDeModeResult.Success {
		fmt.Println("📈 COMPARISON")
		fmt.Println(strings.Repeat("-", 80))

		// Calculate improvements
		durationImprovement := float64(bc.NativeResult.TotalDuration-bc.GoDeModeResult.TotalDuration) / float64(bc.NativeResult.TotalDuration) * 100
		apiCallReduction := float64(bc.NativeResult.APICallCount-bc.GoDeModeResult.APICallCount) / float64(bc.NativeResult.APICallCount) * 100
		nativeTokens := bc.NativeResult.TotalInputTokens + bc.NativeResult.TotalOutputTokens
		godemodeTokens := bc.GoDeModeResult.TotalInputTokens + bc.GoDeModeResult.TotalOutputTokens
		tokenReduction := float64(nativeTokens-godemodeTokens) / float64(nativeTokens) * 100

		fmt.Println()
		fmt.Printf("| Metric              | Native MCP | GoDeMode MCP | Improvement |\n")
		fmt.Printf("|---------------------|------------|--------------|-------------|\n")
		fmt.Printf("| API Calls           | %-10d | %-12d | %.1f%% ↓     |\n",
			bc.NativeResult.APICallCount,
			bc.GoDeModeResult.APICallCount,
			apiCallReduction)
		fmt.Printf("| Total Duration      | %-10v | %-12v | %.1f%% ↓     |\n",
			bc.NativeResult.TotalDuration,
			bc.GoDeModeResult.TotalDuration,
			durationImprovement)
		fmt.Printf("| Total Tokens        | %-10d | %-12d | %.1f%% ↓     |\n",
			nativeTokens,
			godemodeTokens,
			tokenReduction)
		fmt.Printf("| Tool Calls          | %-10d | %-12s | -           |\n",
			bc.NativeResult.ToolCallCount,
			"N/A (in code)")
		fmt.Println()

		fmt.Println("🎯 Key Insights:")
		fmt.Printf("  • GoDeMode reduced API calls by %.1f%% (%d → %d)\n",
			apiCallReduction,
			bc.NativeResult.APICallCount,
			bc.GoDeModeResult.APICallCount)
		fmt.Printf("  • GoDeMode was %.1f%% faster (%v → %v)\n",
			durationImprovement,
			bc.NativeResult.TotalDuration,
			bc.GoDeModeResult.TotalDuration)
		fmt.Printf("  • GoDeMode used %.1f%% fewer tokens (%d → %d)\n",
			tokenReduction,
			nativeTokens,
			godemodeTokens)

		// Cost estimate (Claude Sonnet pricing: $3 per 1M input tokens, $15 per 1M output tokens)
		nativeCost := (float64(bc.NativeResult.TotalInputTokens) * 3.0 / 1000000.0) +
			(float64(bc.NativeResult.TotalOutputTokens) * 15.0 / 1000000.0)
		godemodeCost := (float64(bc.GoDeModeResult.TotalInputTokens) * 3.0 / 1000000.0) +
			(float64(bc.GoDeModeResult.TotalOutputTokens) * 15.0 / 1000000.0)
		costSavings := (nativeCost - godemodeCost) / nativeCost * 100

		fmt.Println()
		fmt.Println("💰 Cost Estimate (Claude Sonnet pricing):")
		fmt.Printf("  • Native MCP:   $%.6f\n", nativeCost)
		fmt.Printf("  • GoDeMode MCP: $%.6f\n", godemodeCost)
		fmt.Printf("  • Savings:      %.1f%% ($%.6f)\n", costSavings, nativeCost-godemodeCost)
	}

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 79))
}

func startMCPServer() (*exec.Cmd, error) {
	fmt.Println("🚀 Starting MCP Server...")

	// Build server
	buildCmd := exec.Command("go", "build", "-o", "mcp-server", "../real-mcp-server/server.go")
	if err := buildCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to build MCP server: %w", err)
	}

	// Start server
	cmd := exec.Command("./mcp-server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Wait for server to start
	fmt.Println("⏳ Waiting for server to start...")
	time.Sleep(2 * time.Second)

	fmt.Println("✅ MCP Server started (PID: %d)\n", cmd.Process.Pid)
	return cmd, nil
}

func stopMCPServer(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		fmt.Printf("\n🛑 Stopping MCP Server (PID: %d)...\n", cmd.Process.Pid)
		cmd.Process.Signal(syscall.SIGTERM)
		cmd.Wait()
		fmt.Println("✅ MCP Server stopped")
	}
}

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY environment variable required")
		fmt.Println("   export ANTHROPIC_API_KEY=your-api-key")
		os.Exit(1)
	}

	// Start MCP server
	serverCmd, err := startMCPServer()
	if err != nil {
		fmt.Printf("❌ Failed to start MCP server: %v\n", err)
		os.Exit(1)
	}
	defer stopMCPServer(serverCmd)

	mcpServerURL := "http://localhost:8080/mcp"

	task := `Complete these 5 utility operations and return the results:
1. Add 10 and 5 together
2. Get the current time
3. Generate a UUID
4. Concatenate ["Hello", "World", "from", "MCP"] with spaces
5. Reverse the string "GoDeMode"

Provide a summary of all results.`

	comparison := &BenchmarkComparison{
		Task: task,
	}

	// Run Native MCP benchmark
	fmt.Println("🔵 Running Native MCP Benchmark...")
	fmt.Println(strings.Repeat("-", 80))
	nativeAgent := NewNativeMCPAgent(apiKey, mcpServerURL)
	nativeResult, err := nativeAgent.RunTask(context.Background(), task)
	if err != nil {
		fmt.Printf("❌ Native MCP failed: %v\n", err)
	} else {
		fmt.Println("✅ Native MCP completed successfully")
	}
	comparison.NativeResult = nativeResult
	fmt.Println()

	// Wait a bit between benchmarks
	time.Sleep(2 * time.Second)

	// Run GoDeMode MCP benchmark
	fmt.Println("🟢 Running GoDeMode MCP Benchmark...")
	fmt.Println(strings.Repeat("-", 80))
	godemodeAgent := NewGoDeModeMCPAgent(apiKey, mcpServerURL)
	godemodeResult, err := godemodeAgent.RunTask(context.Background(), task)
	if err != nil {
		fmt.Printf("❌ GoDeMode MCP failed: %v\n", err)
	} else {
		fmt.Println("✅ GoDeMode MCP completed successfully")
	}
	comparison.GoDeModeResult = godemodeResult
	fmt.Println()

	// Print comparison
	comparison.PrintComparison()

	// Save results to file
	resultsFile := "../results/real-benchmark-results.txt"
	f, err := os.Create(resultsFile)
	if err != nil {
		fmt.Printf("⚠️  Could not save results to file: %v\n", err)
	} else {
		defer f.Close()

		// Redirect stdout to file temporarily
		oldStdout := os.Stdout
		os.Stdout = f
		comparison.PrintComparison()
		os.Stdout = oldStdout

		fmt.Printf("💾 Results saved to: %s\n", resultsFile)
	}
}
