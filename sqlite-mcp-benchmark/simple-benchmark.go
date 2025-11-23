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
	AuditLog      []AuditEntry
	GeneratedCode string // For CodeMode
}

// AuditEntry tracks each operation for auditability
type AuditEntry struct {
	Timestamp   time.Time
	Type        string // "api_call", "tool_call", "tool_result"
	Details     string
	ToolName    string
	ToolArgs    string
	ToolResult  string
	InputTokens  int
	OutputTokens int
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

	// Run both scenarios
	runScenario(apiKey, "Simple", getSimplePrompt(), 5)
	runScenario(apiKey, "Complex", getComplexPrompt(), 12)
}

func getSimplePrompt() string {
	return `You have a SQLite database with a customers table.
Task:
1. List all tables in the database
2. Get the schema of the customers table
3. Read all customers with status 'active'
4. Create a new customer named "Test User" with email "test@example.com" and status "active"
5. Read all active customers again to confirm the new customer was added`
}

func getComplexPrompt() string {
	return `You have a SQLite database with customers, products, orders, and order_items tables.
Perform a comprehensive business analysis:

1. List all tables to understand the schema
2. Get the schema of all 4 tables (customers, products, orders, order_items)
3. Find all customers who have placed orders (join customers and orders)
4. Calculate total revenue per customer by joining orders with order_items and products
5. Find the top-selling product by quantity
6. Create a new order for customer "Alice Smith" with 2 units of "Widget A" and 1 unit of "Gadget B"
7. Update the stock quantity for the ordered products
8. Verify the new order was created by reading it back with all its items
9. Generate a summary report showing: total customers, total products, total orders, and total revenue`
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
		return fmt.Errorf("failed to create customers table: %w", err)
	}

	// Create products table
	_, err = registry.Call("query", map[string]interface{}{
		"sql": `CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL NOT NULL,
			stock_quantity INTEGER DEFAULT 0,
			category TEXT
		)`,
	})
	if err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	// Create orders table
	_, err = registry.Call("query", map[string]interface{}{
		"sql": `CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			order_date TEXT DEFAULT CURRENT_TIMESTAMP,
			status TEXT DEFAULT 'pending',
			FOREIGN KEY (customer_id) REFERENCES customers(id)
		)`,
	})
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	// Create order_items table
	_, err = registry.Call("query", map[string]interface{}{
		"sql": `CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price REAL NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		)`,
	})
	if err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}

	// Seed customers
	customers := []map[string]interface{}{
		{"name": "Alice Smith", "email": "alice@example.com", "status": "active"},
		{"name": "Bob Johnson", "email": "bob@example.com", "status": "pending"},
		{"name": "Carol White", "email": "carol@example.com", "status": "active"},
		{"name": "David Brown", "email": "david@example.com", "status": "active"},
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

	// Seed products
	products := []map[string]interface{}{
		{"name": "Widget A", "price": 29.99, "stock_quantity": 100, "category": "widgets"},
		{"name": "Widget B", "price": 49.99, "stock_quantity": 50, "category": "widgets"},
		{"name": "Gadget A", "price": 99.99, "stock_quantity": 30, "category": "gadgets"},
		{"name": "Gadget B", "price": 149.99, "stock_quantity": 20, "category": "gadgets"},
		{"name": "Tool X", "price": 199.99, "stock_quantity": 15, "category": "tools"},
	}

	for _, p := range products {
		_, err := registry.Call("create_record", map[string]interface{}{
			"table": "products",
			"data":  p,
		})
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
	}

	// Seed orders
	orders := []map[string]interface{}{
		{"customer_id": 1, "status": "completed"},
		{"customer_id": 1, "status": "completed"},
		{"customer_id": 3, "status": "completed"},
		{"customer_id": 4, "status": "pending"},
	}

	for _, o := range orders {
		_, err := registry.Call("create_record", map[string]interface{}{
			"table": "orders",
			"data":  o,
		})
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
	}

	// Seed order_items
	orderItems := []map[string]interface{}{
		{"order_id": 1, "product_id": 1, "quantity": 3, "unit_price": 29.99},
		{"order_id": 1, "product_id": 3, "quantity": 1, "unit_price": 99.99},
		{"order_id": 2, "product_id": 2, "quantity": 2, "unit_price": 49.99},
		{"order_id": 3, "product_id": 4, "quantity": 1, "unit_price": 149.99},
		{"order_id": 3, "product_id": 5, "quantity": 1, "unit_price": 199.99},
		{"order_id": 4, "product_id": 1, "quantity": 5, "unit_price": 29.99},
	}

	for _, oi := range orderItems {
		_, err := registry.Call("create_record", map[string]interface{}{
			"table": "order_items",
			"data":  oi,
		})
		if err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	return nil
}

func runCodeModeBenchmark(apiKey string, prompt string) BenchmarkResult {
	start := time.Now()
	registry := sqlitetools.NewRegistry()
	var auditLog []AuditEntry
	var executedToolCalls int

	// Build system prompt with available tools for CodeMode
	systemPrompt := buildCodeModeSystemPrompt(registry)

	fullPrompt := systemPrompt + "\n\nTask:\n" + prompt

	// Single API call to generate code
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

	// Count registry.Call occurrences in generated code
	toolCalls := strings.Count(goCode, "registry.Call")
	if toolCalls == 0 {
		toolCalls = strings.Count(goCode, `Call("`)
	}

	// Log the tool calls found in generated code
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
func executeGeneratedCode(code string, registry *sqlitetools.Registry, auditLog *[]AuditEntry) (string, int, error) {
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
	registry := sqlitetools.NewRegistry()
	var auditLog []AuditEntry

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

				// Truncate result for logging if too long
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

		// Add assistant message and tool results
		messages = append(messages, Message{Role: "assistant", Content: assistantContent})
		messages = append(messages, Message{Role: "user", Content: toolResults})

		// Safety limit
		if apiCallCount > 20 {
			auditLog = append(auditLog, AuditEntry{
				Timestamp: time.Now(),
				Type:      "warning",
				Details:   "Reached safety limit of 20 API calls",
			})
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
Generate a complete, valid Go program that:
1. Uses the registry to call the necessary tools
2. Implements loops for iterating over results
3. Handles errors appropriately
4. Outputs results using fmt.Println

IMPORTANT: Use valid Go syntax only. Do NOT use Python-style string multiplication like "=" * 60.
Instead use strings.Repeat("=", 60) or just print a literal string.

The registry variable is already defined - do NOT redefine it. Just use registry.Call() directly.

Example usage:
  result, err := registry.Call("read_records", map[string]interface{}{
      "table": "customers",
      "conditions": map[string]interface{}{"status": "active"},
  })
  if err != nil {
      fmt.Println("Error:", err)
      return
  }

  // Type assert to access the data
  if data, ok := result.(map[string]interface{}); ok {
      if rows, ok := data["rows"].([]interface{}); ok {
          for _, row := range rows {
              fmt.Println(row)
          }
      }
  }
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
				// Format args nicely
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
		code := r.GeneratedCode
		lines := strings.Split(code, "\n")
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
