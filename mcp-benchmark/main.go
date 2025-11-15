package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/imran31415/godemode/benchmark/llm"
	utilitytools "github.com/imran31415/godemode/mcp-benchmark/godemode"
	"github.com/imran31415/godemode/pkg/executor"
)

const benchmarkTask = `Complete the following utility operations:

1. Add 10 and 5 together
2. Get the current time
3. Generate a UUID
4. Concatenate the strings ["Hello", "World", "from", "MCP"] with separator " "
5. Reverse the string "GoDeMode"

Print all results.`

type BenchmarkResult struct {
	Mode       string
	Duration   time.Duration
	TokensUsed int
	APICalls   int
	Success    bool
	Output     string
	Error      error
}

func main() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("MCP BENCHMARK: Native MCP vs GoDeMode MCP")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Simulate Native MCP benchmark
	fmt.Println("NATIVE MCP (Sequential Tool Calling):")
	fmt.Println("  In a traditional MCP setup, this task would require:")
	fmt.Println("  - 5-6 API calls (one per tool + initial planning + final summary)")
	fmt.Println("  - Sequential execution with latency between each call")
	fmt.Println("  - ~3000-4000 tokens (request + response per call)")
	fmt.Println()

	nativeResult := BenchmarkResult{
		Mode:       "Native MCP (Estimated)",
		APICalls:   6,
		TokensUsed: 3500,
		Success:    true,
	}

	// Run GoDeMode MCP benchmark
	fmt.Println("Running GODEMODE MCP (Code Mode)...")
	godemodeResult := runGodeModeMCP()

	// Print results
	printResults(nativeResult, godemodeResult)
}

func runGodeModeMCP() BenchmarkResult {
	start := time.Now()
	result := BenchmarkResult{
		Mode: "GoDeMode MCP",
	}

	// Check for API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		result.Error = fmt.Errorf("ANTHROPIC_API_KEY not set")
		result.Duration = time.Since(start)
		return result
	}

	client := llm.NewClient()
	registry := utilitytools.NewRegistry()

	// Get available tools
	toolsList := registry.List()

	systemMsg := fmt.Sprintf(`You are an expert Go programmer. Write a Go program using the provided utility tools.

Available tools: %v

Each tool can be called using: registry.Call("toolName", map[string]interface{}{...})

Example:
  result, _ := registry.Call("add", map[string]interface{}{"a": 10.0, "b": 5.0})
  fmt.Println(result)

Write complete working code that calls all required tools and prints results.
Write ONLY the code inside main(), do not include package or import statements.`, toolsList)

	ctx := context.Background()

	// Single API call to generate code
	messages := []llm.Message{
		{Role: "user", Content: benchmarkTask},
	}

	response, err := client.GenerateResponse(ctx, systemMsg, messages)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	result.TokensUsed = response.TokensUsed
	result.APICalls = 1

	// Extract and clean code
	code := extractGoCode(response.Content)

	// Execute with GoDeMode
	exec := executor.NewInterpreterExecutor()

	// Wrap code with full program
	fullCode := fmt.Sprintf(`package main

import (
	"fmt"
	utilitytools "github.com/imran31415/godemode/mcp-benchmark/godemode"
)

func main() {
	registry := utilitytools.NewRegistry()

	%s
}
`, code)

	fmt.Println("\nGenerated Code:")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println(code)
	fmt.Println(strings.Repeat("-", 40))

	execResult, err := exec.Execute(ctx, fullCode, 30*time.Second)
	if err != nil {
		result.Error = fmt.Errorf("code execution failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Output = execResult.Output
	result.Success = true
	result.Duration = time.Since(start)
	return result
}

func extractGoCode(content string) string {
	// Extract code between ```go and ``` or ``` and ```
	if strings.Contains(content, "```go") {
		start := strings.Index(content, "```go")
		end := strings.Index(content[start+5:], "```")
		if start != -1 && end != -1 {
			return strings.TrimSpace(content[start+5 : start+5+end])
		}
	}

	if strings.Contains(content, "```") {
		start := strings.Index(content, "```")
		end := strings.Index(content[start+3:], "```")
		if start != -1 && end != -1 {
			code := strings.TrimSpace(content[start+3 : start+3+end])
			// Remove "go" if it's the first line
			if strings.HasPrefix(code, "go\n") {
				code = code[3:]
			}
			return code
		}
	}

	return content
}

func printResults(native, godemode BenchmarkResult) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Native MCP Results
	fmt.Println("NATIVE MCP (Sequential Tool Calling - Estimated):")
	fmt.Printf("  API Calls:  %d\n", native.APICalls)
	fmt.Printf("  Tokens:     ~%d\n", native.TokensUsed)
	fmt.Printf("  Duration:   ~15-20s (with network latency)\n")
	fmt.Println()

	// GoDeMode MCP Results
	fmt.Println("GODEMODE MCP (Code Mode - Actual):")
	fmt.Printf("  Success:    %v\n", godemode.Success)
	fmt.Printf("  Duration:   %v\n", godemode.Duration)
	fmt.Printf("  API Calls:  %d\n", godemode.APICalls)
	fmt.Printf("  Tokens:     %d\n", godemode.TokensUsed)
	if godemode.Error != nil {
		fmt.Printf("  Error:      %v\n", godemode.Error)
	}
	fmt.Println()

	if godemode.Output != "" {
		fmt.Println("OUTPUT:")
		fmt.Println(godemode.Output)
		fmt.Println()
	}

	// Comparison
	if godemode.Success {
		fmt.Println("COMPARISON:")
		tokenDiff := native.TokensUsed - godemode.TokensUsed
		apiCallDiff := native.APICalls - godemode.APICalls

		fmt.Printf("  API Calls:  GoDeMode made %d fewer calls (%.1f%% reduction)\n",
			apiCallDiff, float64(apiCallDiff)/float64(native.APICalls)*100)
		fmt.Printf("  Tokens:     GoDeMode used ~%d fewer tokens (%.1f%% reduction)\n",
			tokenDiff, float64(tokenDiff)/float64(native.TokensUsed)*100)
		fmt.Printf("  Efficiency: Single code generation vs %d sequential tool calls\n", native.APICalls)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
}
