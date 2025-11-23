package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/imran31415/godemode/pkg/executor"
	githubtools "github.com/imran31415/godemode/github-mcp-benchmark/generated"
)

// API types
type ClaudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ContentBlock struct {
	Type      string      `json:"type"`
	Text      string      `json:"text,omitempty"`
	ID        string      `json:"id,omitempty"`
	Name      string      `json:"name,omitempty"`
	Input     interface{} `json:"input,omitempty"`
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   string      `json:"content,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type ClaudeResponse struct {
	Content    []ResponseContent `json:"content"`
	Usage      Usage             `json:"usage"`
	StopReason string            `json:"stop_reason"`
}

type ResponseContent struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Benchmark result
type BenchmarkResult struct {
	Approach      string
	Duration      time.Duration
	APICallCount  int
	InputTokens   int
	OutputTokens  int
	TotalTokens   int
	ToolCalls     int
	Success       bool
	Error         string
	EstimatedCost float64
	AuditLog      []AuditEntry
	GeneratedCode string
}

// AuditEntry tracks each operation
type AuditEntry struct {
	Timestamp    time.Time
	Type         string
	Details      string
	ToolName     string
	ToolArgs     string
	ToolResult   string
	InputTokens  int
	OutputTokens int
}

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		fmt.Println("Error: GITHUB_TOKEN environment variable not set")
		os.Exit(1)
	}

	// Initialize GitHub client
	githubtools.InitGitHub(githubToken)

	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println("GitHub MCP Benchmark: CodeMode vs Native Tool Calling")
	fmt.Println("=" + strings.Repeat("=", 69))

	// Run both scenarios
	runScenario(apiKey, "Simple", getSimplePrompt(), 5)
	runScenario(apiKey, "Complex", getComplexPrompt(), 12)
}

func getSimplePrompt() string {
	return `You have access to the GitHub API through the provided tools.

Task: Perform a basic issue analysis on the repository "cli/cli"

1. Get the repository details for cli/cli
2. List the 5 most recent open issues
3. Get the details of the first issue from the list
4. List the comments on that issue
5. Search for any related pull requests mentioning that issue number`
}

func getComplexPrompt() string {
	return `You have access to the GitHub API through the provided tools.

Task: Perform a comprehensive analysis across multiple repositories

1. Search for repositories with "mcp server" in their name (limit to 3 results)
2. For each of the 3 repositories found:
   a. Get the repository details
   b. List open issues (limit to 5 per repo)
   c. List open pull requests (limit to 3 per repo)
3. For the first repository, get details of the first open issue (if any)
4. Search for issues mentioning "bug" across all of GitHub (limit 5)
5. List the recent commits from the first repository found (limit 5)
6. Generate a summary of:
   - Total repositories analyzed
   - Total open issues found
   - Total open PRs found
   - Any notable patterns in issue titles`
}

func runScenario(apiKey, name, prompt string, expectedOps int) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Printf("SCENARIO: %s\n", name)
	fmt.Printf("%s\n", strings.Repeat("=", 70))
	fmt.Printf("Expected operations: ~%d tool calls\n\n", expectedOps)

	// Run CodeMode benchmark
	fmt.Println("Running CodeMode approach...")
	codeModeResult := runCodeModeBenchmark(apiKey, prompt)
	printResult(codeModeResult)

	// Run Tool Calling benchmark
	fmt.Println("\nRunning Native Tool Calling approach...")
	toolCallResult := runToolCallingBenchmark(apiKey, prompt)
	printResult(toolCallResult)

	// Compare
	printComparison(codeModeResult, toolCallResult)
}

func runCodeModeBenchmark(apiKey string, prompt string) BenchmarkResult {
	start := time.Now()
	registry := githubtools.NewRegistry()
	var auditLog []AuditEntry
	var executedToolCalls int

	// Build system prompt
	systemPrompt := buildCodeModeSystemPrompt(registry)
	fullPrompt := systemPrompt + "\n\nTask:\n" + prompt

	auditLog = append(auditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "api_call",
		Details:   "Sending prompt to Claude for code generation",
	})

	resp, err := callClaude(apiKey, fullPrompt, nil)
	if err != nil {
		auditLog = append(auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "error",
			Details:   fmt.Sprintf("API call failed: %s", err.Error()),
		})
		return BenchmarkResult{
			Approach: "CodeMode",
			Duration: time.Since(start),
			Success:  false,
			Error:    err.Error(),
			AuditLog: auditLog,
		}
	}

	auditLog = append(auditLog, AuditEntry{
		Timestamp:    time.Now(),
		Type:         "api_response",
		Details:      "Received code generation response",
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
	})

	// Extract generated code
	generatedCode := ""
	for _, block := range resp.Content {
		if block.Type == "text" {
			generatedCode = block.Text
		}
	}

	// Extract just the Go code from markdown
	goCode := extractGoCode(generatedCode)

	// Count tool calls in generated code
	toolCalls := strings.Count(goCode, "registry.Call")
	if toolCalls == 0 {
		toolCalls = strings.Count(goCode, `Call("`)
	}

	auditLog = append(auditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "code_analysis",
		Details:   fmt.Sprintf("Generated code contains %d tool calls", toolCalls),
	})

	// Execute the generated code
	auditLog = append(auditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "execution",
		Details:   "Starting code execution via Yaegi interpreter",
	})

	output, execToolCalls, execErr := executeGeneratedCode(goCode, registry, &auditLog)
	executedToolCalls = execToolCalls

	if execErr != nil {
		auditLog = append(auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "error",
			Details:   fmt.Sprintf("Execution failed: %s", execErr.Error()),
		})
	} else {
		auditLog = append(auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "execution_complete",
			Details:   fmt.Sprintf("Execution completed with %d tool calls", executedToolCalls),
		})
		if len(output) > 500 {
			output = output[:500] + "..."
		}
		if output != "" {
			auditLog = append(auditLog, AuditEntry{
				Timestamp:  time.Now(),
				Type:       "output",
				ToolResult: output,
			})
		}
	}

	// Calculate cost
	inputCost := float64(resp.Usage.InputTokens) * 0.003 / 1000
	outputCost := float64(resp.Usage.OutputTokens) * 0.015 / 1000

	return BenchmarkResult{
		Approach:      "CodeMode",
		Duration:      time.Since(start),
		APICallCount:  1,
		InputTokens:   resp.Usage.InputTokens,
		OutputTokens:  resp.Usage.OutputTokens,
		TotalTokens:   resp.Usage.InputTokens + resp.Usage.OutputTokens,
		ToolCalls:     executedToolCalls,
		Success:       execErr == nil,
		Error:         func() string { if execErr != nil { return execErr.Error() }; return "" }(),
		EstimatedCost: inputCost + outputCost,
		AuditLog:      auditLog,
		GeneratedCode: generatedCode,
	}
}

// extractGoCode uses the core preprocessor to extract Go code from markdown
func extractGoCode(text string) string {
	preprocessor := executor.NewCodePreprocessor()
	return preprocessor.ExtractGoCode(text)
}

// executeGeneratedCode runs the generated Go code using the InterpreterExecutor
func executeGeneratedCode(code string, registry *githubtools.Registry, auditLog *[]AuditEntry) (string, int, error) {
	// Track tool calls
	toolCallCount := 0

	// Create a wrapper registry that logs calls
	wrappedCall := func(name string, args map[string]interface{}) (interface{}, error) {
		toolCallCount++
		argsJSON, _ := json.Marshal(args)

		*auditLog = append(*auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "tool_call",
			ToolName:  name,
			ToolArgs:  string(argsJSON),
			Details:   fmt.Sprintf("Executed tool call #%d", toolCallCount),
		})

		result, err := registry.Call(name, args)

		var resultStr string
		if err != nil {
			resultStr = fmt.Sprintf("Error: %v", err)
		} else {
			resultBytes, _ := json.Marshal(result)
			resultStr = string(resultBytes)
			if len(resultStr) > 200 {
				resultStr = resultStr[:200] + "..."
			}
		}

		*auditLog = append(*auditLog, AuditEntry{
			Timestamp:  time.Now(),
			Type:       "tool_result",
			ToolName:   name,
			ToolResult: resultStr,
		})

		return result, err
	}

	// Use the core executor's ExecuteGeneratedCode API
	exec := executor.NewInterpreterExecutor()
	result, err := exec.ExecuteGeneratedCode(context.Background(), code, 60*time.Second, wrappedCall)

	if err != nil {
		return result.Stdout, toolCallCount, fmt.Errorf("execution error: %w", err)
	}

	if !result.Success {
		return result.Stdout, toolCallCount, fmt.Errorf("execution failed: %s", result.Error)
	}

	return result.Stdout, toolCallCount, nil
}

func runToolCallingBenchmark(apiKey string, prompt string) BenchmarkResult {
	start := time.Now()
	registry := githubtools.NewRegistry()
	var auditLog []AuditEntry

	tools := buildTools(registry)
	var totalInputTokens, totalOutputTokens, apiCallCount, toolCallCount int

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	for {
		apiCallCount++

		auditLog = append(auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "api_call",
			Details:   fmt.Sprintf("API call #%d", apiCallCount),
		})

		resp, err := callClaudeWithTools(apiKey, messages, tools)
		if err != nil {
			auditLog = append(auditLog, AuditEntry{
				Timestamp: time.Now(),
				Type:      "error",
				Details:   fmt.Sprintf("API call failed: %s", err.Error()),
			})
			return BenchmarkResult{
				Approach:     "ToolCalling",
				Duration:     time.Since(start),
				APICallCount: apiCallCount,
				Success:      false,
				Error:        err.Error(),
				AuditLog:     auditLog,
			}
		}

		auditLog = append(auditLog, AuditEntry{
			Timestamp:    time.Now(),
			Type:         "api_response",
			Details:      fmt.Sprintf("Response #%d (stop_reason: %s)", apiCallCount, resp.StopReason),
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		})

		totalInputTokens += resp.Usage.InputTokens
		totalOutputTokens += resp.Usage.OutputTokens

		if resp.StopReason == "end_turn" {
			break
		}

		var toolResults []ContentBlock
		var assistantContent []ContentBlock

		for _, block := range resp.Content {
			if block.Type == "tool_use" {
				toolCallCount++

				var args map[string]interface{}
				json.Unmarshal(block.Input, &args)
				argsJSON, _ := json.Marshal(args)

				auditLog = append(auditLog, AuditEntry{
					Timestamp: time.Now(),
					Type:      "tool_call",
					ToolName:  block.Name,
					ToolArgs:  string(argsJSON),
					Details:   fmt.Sprintf("Tool call #%d", toolCallCount),
				})

				result, err := registry.Call(block.Name, args)

				var resultStr string
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultBytes, _ := json.Marshal(result)
					resultStr = string(resultBytes)
				}

				logResult := resultStr
				if len(logResult) > 200 {
					logResult = logResult[:200] + "..."
				}

				auditLog = append(auditLog, AuditEntry{
					Timestamp:  time.Now(),
					Type:       "tool_result",
					ToolName:   block.Name,
					ToolResult: logResult,
				})

				toolResults = append(toolResults, ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   resultStr,
				})
				assistantContent = append(assistantContent, ContentBlock{
					Type:  "tool_use",
					ID:    block.ID,
					Name:  block.Name,
					Input: args,
				})
			} else if block.Type == "text" {
				assistantContent = append(assistantContent, ContentBlock{
					Type: "text",
					Text: block.Text,
				})
			}
		}

		if len(toolResults) == 0 {
			break
		}

		messages = append(messages, Message{Role: "assistant", Content: assistantContent})
		messages = append(messages, Message{Role: "user", Content: toolResults})

		if apiCallCount > 25 {
			auditLog = append(auditLog, AuditEntry{
				Timestamp: time.Now(),
				Type:      "warning",
				Details:   "Reached safety limit of 25 API calls",
			})
			break
		}
	}

	inputCost := float64(totalInputTokens) * 0.003 / 1000
	outputCost := float64(totalOutputTokens) * 0.015 / 1000

	return BenchmarkResult{
		Approach:      "ToolCalling",
		Duration:      time.Since(start),
		APICallCount:  apiCallCount,
		InputTokens:   totalInputTokens,
		OutputTokens:  totalOutputTokens,
		TotalTokens:   totalInputTokens + totalOutputTokens,
		ToolCalls:     toolCallCount,
		Success:       true,
		EstimatedCost: inputCost + outputCost,
		AuditLog:      auditLog,
	}
}

func callClaude(apiKey, prompt string, tools []Tool) (*ClaudeResponse, error) {
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	req := ClaudeRequest{
		Model:     model,
		MaxTokens: 8192,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	if tools != nil {
		req.Tools = tools
	}

	return sendRequest(apiKey, req)
}

func callClaudeWithTools(apiKey string, messages []Message, tools []Tool) (*ClaudeResponse, error) {
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	req := ClaudeRequest{
		Model:     model,
		MaxTokens: 8192,
		Messages:  messages,
		Tools:     tools,
	}

	return sendRequest(apiKey, req)
}

func sendRequest(apiKey string, req ClaudeRequest) (*ClaudeResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", httpResp.StatusCode, string(body))
	}

	var resp ClaudeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func buildCodeModeSystemPrompt(registry *githubtools.Registry) string {
	var sb strings.Builder

	sb.WriteString(`You are a code generation assistant. Generate complete, executable Go code to accomplish the user's task.

The code will be executed in an environment with access to the GitHub API through a tool registry.

Available tools (call via registry.Call("tool_name", args)):

`)

	for _, tool := range registry.ListTools() {
		sb.WriteString(fmt.Sprintf("## %s\n%s\n", tool.Name, tool.Description))
		if len(tool.Parameters) > 0 {
			sb.WriteString("Parameters:\n")
			for _, p := range tool.Parameters {
				req := ""
				if p.Required {
					req = " (required)"
				}
				sb.WriteString(fmt.Sprintf("  - %s: %s%s\n", p.Name, p.Type, req))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`
Generate a complete, valid Go program that:
1. Uses the registry to call the necessary tools
2. Implements loops for iterating over results
3. Handles errors appropriately
4. Outputs results using fmt.Println

IMPORTANT: Use valid Go syntax only. Do NOT use Python-style string multiplication like "=" * 60.
Instead use strings.Repeat("=", 60) or just print a literal string.

The registry variable is already defined - do NOT redefine it. Just use registry.Call() directly.

Example usage:
  result, err := registry.Call("list_issues", map[string]interface{}{
      "owner": "cli",
      "repo": "cli",
      "state": "open",
      "per_page": 5,
  })
  if err != nil {
      fmt.Println("Error:", err)
      return
  }

  // Type assert to access the data
  if data, ok := result.(map[string]interface{}); ok {
      if items, ok := data["items"].([]interface{}); ok {
          for _, item := range items {
              fmt.Println(item)
          }
      }
  }
`)

	return sb.String()
}

func buildTools(registry *githubtools.Registry) []Tool {
	var tools []Tool

	for _, tool := range registry.ListTools() {
		properties := make(map[string]interface{})
		var required []string

		for _, param := range tool.Parameters {
			propType := "string"
			switch param.Type {
			case "int":
				propType = "integer"
			case "[]string":
				propType = "array"
			}

			properties[param.Name] = map[string]interface{}{
				"type":        propType,
				"description": param.Description,
			}

			if param.Required {
				required = append(required, param.Name)
			}
		}

		tools = append(tools, Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: InputSchema{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		})
	}

	return tools
}

func printResult(r BenchmarkResult) {
	status := "✅"
	if !r.Success {
		status = "❌"
	}

	fmt.Printf("\n  %s %s:\n", status, r.Approach)
	fmt.Printf("    Duration:     %v\n", r.Duration.Round(time.Millisecond))
	fmt.Printf("    API Calls:    %d\n", r.APICallCount)
	fmt.Printf("    Tool Calls:   %d\n", r.ToolCalls)
	fmt.Printf("    Tokens:       %d (in: %d, out: %d)\n", r.TotalTokens, r.InputTokens, r.OutputTokens)
	fmt.Printf("    Est. Cost:    $%.4f\n", r.EstimatedCost)
	if r.Error != "" {
		fmt.Printf("    Error:        %s\n", r.Error)
	}

	// Print audit log
	if len(r.AuditLog) > 0 {
		fmt.Printf("\n    --- Audit Log ---\n")
		for i, entry := range r.AuditLog {
			timestamp := entry.Timestamp.Format("15:04:05.000")
			switch entry.Type {
			case "api_call":
				fmt.Printf("    [%s] %d. API_CALL: %s\n", timestamp, i+1, entry.Details)
			case "api_response":
				fmt.Printf("    [%s] %d. API_RESPONSE: %s (tokens: in=%d, out=%d)\n",
					timestamp, i+1, entry.Details, entry.InputTokens, entry.OutputTokens)
			case "tool_call":
				fmt.Printf("    [%s] %d. TOOL_CALL: %s\n", timestamp, i+1, entry.Details)
				fmt.Printf("                    Tool: %s\n", entry.ToolName)
				args := entry.ToolArgs
				if len(args) > 100 {
					args = args[:100] + "..."
				}
				fmt.Printf("                    Args: %s\n", args)
			case "tool_result":
				fmt.Printf("    [%s] %d. TOOL_RESULT: %s\n", timestamp, i+1, entry.ToolName)
				fmt.Printf("                    Result: %s\n", entry.ToolResult)
			case "code_analysis":
				fmt.Printf("    [%s] %d. CODE_ANALYSIS: %s\n", timestamp, i+1, entry.Details)
			case "execution":
				fmt.Printf("    [%s] %d. EXECUTION: %s\n", timestamp, i+1, entry.Details)
			case "execution_complete":
				fmt.Printf("    [%s] %d. EXECUTION_COMPLETE: %s\n", timestamp, i+1, entry.Details)
			case "output":
				fmt.Printf("    [%s] %d. OUTPUT:\n%s\n", timestamp, i+1, entry.ToolResult)
			case "error":
				fmt.Printf("    [%s] %d. ERROR: %s\n", timestamp, i+1, entry.Details)
			case "warning":
				fmt.Printf("    [%s] %d. WARNING: %s\n", timestamp, i+1, entry.Details)
			default:
				fmt.Printf("    [%s] %d. %s: %s\n", timestamp, i+1, entry.Type, entry.Details)
			}
		}
		fmt.Println()
	}

	// Print generated code for CodeMode (truncated)
	if r.GeneratedCode != "" {
		fmt.Printf("    --- Generated Code (truncated) ---\n")
		lines := strings.Split(r.GeneratedCode, "\n")
		maxLines := 30
		if len(lines) > maxLines {
			for i := 0; i < maxLines; i++ {
				line := lines[i]
				if len(line) > 100 {
					line = line[:100] + "..."
				}
				fmt.Printf("    %s\n", line)
			}
			fmt.Printf("    ... (%d more lines)\n", len(lines)-maxLines)
		} else {
			for _, line := range lines {
				if len(line) > 100 {
					line = line[:100] + "..."
				}
				fmt.Printf("    %s\n", line)
			}
		}
		fmt.Println()
	}
}

func printComparison(codeMode, toolCalling BenchmarkResult) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("COMPARISON")
	fmt.Println(strings.Repeat("=", 70))

	if codeMode.Success && toolCalling.Success {
		speedup := float64(toolCalling.Duration) / float64(codeMode.Duration)
		tokenReduction := 100 * (1 - float64(codeMode.TotalTokens)/float64(toolCalling.TotalTokens))
		costSavings := 100 * (1 - codeMode.EstimatedCost/toolCalling.EstimatedCost)

		fmt.Printf("\nCodeMode vs Tool Calling:\n")
		fmt.Printf("  Speed:         %.2fx faster\n", speedup)
		fmt.Printf("  Tokens:        %.1f%% fewer tokens\n", tokenReduction)
		fmt.Printf("  Cost:          %.1f%% cheaper\n", costSavings)
		fmt.Printf("  API Calls:     %d vs %d\n", codeMode.APICallCount, toolCalling.APICallCount)
	}
}
