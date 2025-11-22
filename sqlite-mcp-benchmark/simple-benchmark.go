package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	sqlitetools "github.com/imran31415/godemode/sqlite-mcp-benchmark/generated"
)

// API types
type ClaudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

type Message struct {
	Role    string        `json:"role"`
	Content interface{}   `json:"content"` // string or []ContentBlock
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
}

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	// Initialize database
	dbPath := "benchmark.db"
	if err := sqlitetools.InitDB(dbPath); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer sqlitetools.CloseDB()
	defer os.Remove(dbPath)

	// Setup test data
	if err := setupTestData(); err != nil {
		fmt.Printf("Failed to setup test data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println("SQLite MCP Benchmark: CodeMode vs Native Tool Calling")
	fmt.Println("=" + strings.Repeat("=", 69))

	// Simple scenario prompt
	prompt := `You have a SQLite database with a customers table.
Task:
1. List all tables in the database
2. Get the schema of the customers table
3. Read all customers with status 'active'
4. Create a new customer named "Test User" with email "test@example.com" and status "active"
5. Read all active customers again to confirm the new customer was added`

	fmt.Println("\nScenario: Simple multi-step CRUD operations")
	fmt.Println("Expected operations: ~5 tool calls\n")

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

func setupTestData() error {
	registry := sqlitetools.NewRegistry()

	// Create customers table
	_, err := registry.Call("query", map[string]interface{}{
		"sql": `CREATE TABLE IF NOT EXISTS customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			status TEXT DEFAULT 'pending',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
	})
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Seed customers
	customers := []map[string]interface{}{
		{"name": "Alice Smith", "email": "alice@example.com", "status": "active"},
		{"name": "Bob Johnson", "email": "bob@example.com", "status": "pending"},
		{"name": "Carol White", "email": "carol@example.com", "status": "active"},
	}

	for _, c := range customers {
		_, err := registry.Call("create_record", map[string]interface{}{
			"table": "customers",
			"data":  c,
		})
		if err != nil {
			return fmt.Errorf("failed to create customer: %w", err)
		}
	}

	return nil
}

func runCodeModeBenchmark(apiKey string, prompt string) BenchmarkResult {
	start := time.Now()
	registry := sqlitetools.NewRegistry()

	// Build system prompt with available tools for CodeMode
	systemPrompt := buildCodeModeSystemPrompt(registry)

	fullPrompt := systemPrompt + "\n\nTask:\n" + prompt

	// Single API call to generate code
	resp, err := callClaude(apiKey, fullPrompt, nil)
	if err != nil {
		return BenchmarkResult{
			Approach: "CodeMode",
			Duration: time.Since(start),
			Success:  false,
			Error:    err.Error(),
		}
	}

	// Extract generated code and count tool calls
	generatedCode := ""
	for _, block := range resp.Content {
		if block.Type == "text" {
			generatedCode = block.Text
		}
	}

	// Count registry.Call occurrences in generated code
	toolCalls := strings.Count(generatedCode, "registry.Call")
	if toolCalls == 0 {
		toolCalls = strings.Count(generatedCode, `Call("`)
	}

	// Calculate cost (Claude Sonnet pricing: $3/MTok in, $15/MTok out)
	inputCost := float64(resp.Usage.InputTokens) * 0.003 / 1000
	outputCost := float64(resp.Usage.OutputTokens) * 0.015 / 1000

	return BenchmarkResult{
		Approach:      "CodeMode",
		Duration:      time.Since(start),
		APICallCount:  1,
		InputTokens:   resp.Usage.InputTokens,
		OutputTokens:  resp.Usage.OutputTokens,
		TotalTokens:   resp.Usage.InputTokens + resp.Usage.OutputTokens,
		ToolCalls:     toolCalls,
		Success:       true,
		EstimatedCost: inputCost + outputCost,
	}
}

func runToolCallingBenchmark(apiKey string, prompt string) BenchmarkResult {
	start := time.Now()
	registry := sqlitetools.NewRegistry()

	// Build tools for native tool calling
	tools := buildTools(registry)

	var totalInputTokens, totalOutputTokens, apiCallCount, toolCallCount int

	// Initial messages
	messages := []Message{
		{Role: "user", Content: prompt},
	}

	// Tool calling loop
	for {
		apiCallCount++

		resp, err := callClaudeWithTools(apiKey, messages, tools)
		if err != nil {
			return BenchmarkResult{
				Approach:     "ToolCalling",
				Duration:     time.Since(start),
				APICallCount: apiCallCount,
				Success:      false,
				Error:        err.Error(),
			}
		}

		totalInputTokens += resp.Usage.InputTokens
		totalOutputTokens += resp.Usage.OutputTokens

		// Check if we're done
		if resp.StopReason == "end_turn" {
			break
		}

		// Process tool calls
		var toolResults []ContentBlock
		var assistantContent []ContentBlock

		for _, block := range resp.Content {
			if block.Type == "tool_use" {
				toolCallCount++

				// Execute the tool
				var args map[string]interface{}
				json.Unmarshal(block.Input, &args)

				result, err := registry.Call(block.Name, args)

				var resultStr string
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultBytes, _ := json.Marshal(result)
					resultStr = string(resultBytes)
				}

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

		// Add assistant message and tool results
		messages = append(messages, Message{Role: "assistant", Content: assistantContent})
		messages = append(messages, Message{Role: "user", Content: toolResults})

		// Safety limit
		if apiCallCount > 20 {
			break
		}
	}

	// Calculate cost
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
	}
}

func callClaude(apiKey, prompt string, tools []Tool) (*ClaudeResponse, error) {
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	req := ClaudeRequest{
		Model:     model,
		MaxTokens: 4096,
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
		MaxTokens: 4096,
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

func buildCodeModeSystemPrompt(registry *sqlitetools.Registry) string {
	var sb strings.Builder

	sb.WriteString(`You are a code generation assistant. Generate complete, executable Go code to accomplish the user's task.

The code will be executed in an environment with access to a SQLite database through a tool registry.

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
Generate a complete Go program that:
1. Uses the registry to call the necessary tools
2. Implements any required loops and conditional logic
3. Outputs results using fmt.Println

Example usage:
  result, err := registry.Call("read_records", map[string]interface{}{
      "table": "customers",
      "conditions": map[string]interface{}{"status": "active"},
  })
  if err != nil {
      fmt.Println("Error:", err)
  }
  fmt.Println(result)
`)

	return sb.String()
}

func buildTools(registry *sqlitetools.Registry) []Tool {
	var tools []Tool

	for _, tool := range registry.ListTools() {
		properties := make(map[string]interface{})
		var required []string

		for _, param := range tool.Parameters {
			propType := "string"
			switch param.Type {
			case "float64":
				propType = "number"
			case "map[string]interface{}":
				propType = "object"
			case "[]interface{}":
				propType = "array"
			}

			properties[param.Name] = map[string]interface{}{
				"type":        propType,
				"description": param.Name,
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

	fmt.Printf("  %s %s:\n", status, r.Approach)
	fmt.Printf("    Duration:     %v\n", r.Duration.Round(time.Millisecond))
	fmt.Printf("    API Calls:    %d\n", r.APICallCount)
	fmt.Printf("    Tool Calls:   %d\n", r.ToolCalls)
	fmt.Printf("    Tokens:       %d (in: %d, out: %d)\n", r.TotalTokens, r.InputTokens, r.OutputTokens)
	fmt.Printf("    Est. Cost:    $%.4f\n", r.EstimatedCost)
	if r.Error != "" {
		fmt.Printf("    Error:        %s\n", r.Error)
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
