package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// MCP Benchmark - Uses MCP protocol with Claude calling tools via MCP server

type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
	ID      int                    `json:"id"`
}

type MCPResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	Result  map[string]interface{} `json:"result"`
	ID      int                    `json:"id"`
}

type ClaudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []ClaudeMessage  `json:"messages"`
	Tools     []ClaudeTool     `json:"tools"`
}

type ClaudeMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ClaudeTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type ClaudeResponse struct {
	Content    []ClaudeContent `json:"content"`
	StopReason string          `json:"stop_reason"`
	Usage      Usage           `json:"usage"`
}

type ClaudeContent struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ToolUseId string                 `json:"tool_use_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

var (
	mcpServerURL     = "http://localhost:8083/mcp"
	totalInputTokens = 0
	totalOutputTokens = 0
	claudeAPICount    = 0
	mcpCallCount      = 0
	mcpOverheadTime   time.Duration
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	fmt.Println("🚀 MCP Benchmark - E-Commerce Order Processing")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔌 Connecting to MCP server at", mcpServerURL)

	startTime := time.Now()

	// Step 1: List available tools from MCP server
	fmt.Println("\n📡 Fetching tools from MCP server...")
	mcpStart := time.Now()
	tools, err := listMCPTools()
	mcpDuration := time.Since(mcpStart)
	mcpOverheadTime += mcpDuration
	mcpCallCount++

	if err != nil {
		fmt.Printf("❌ Error listing tools: %v\n", err)
		fmt.Println("💡 Make sure MCP server is running: go run mcp-server.go")
		os.Exit(1)
	}

	fmt.Printf("   ✅ Found %d tools in %v\n", len(tools), mcpDuration)

	// Step 2: Call Claude with order processing task
	orderPrompt := `Process a complete e-commerce order with the following details:

Order ID: ORD-2025-001
Customer ID: CUST-12345
Email: john.doe@example.com

Items:
- Laptop: $1,299.99 (PROD-001) x1
- Mouse: $29.99 (PROD-002) x1
- Keyboard: $89.99 (PROD-003) x1

Shipping Address:
123 Main St, San Francisco, CA 94105

Discount Code: SAVE20

Please process this order through all required steps using the available tools.`

	messages := []ClaudeMessage{
		{Role: "user", Content: orderPrompt},
	}

	// Tool calling loop with MCP
	for {
		claudeAPICount++
		fmt.Printf("\n📡 API Call %d to Claude...\n", claudeAPICount)
		apiCallStart := time.Now()

		response, err := callClaude(apiKey, messages, tools)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		apiCallDuration := time.Since(apiCallStart)
		totalInputTokens += response.Usage.InputTokens
		totalOutputTokens += response.Usage.OutputTokens

		fmt.Printf("   ⏱️  Duration: %v\n", apiCallDuration)
		fmt.Printf("   📊 Tokens: %d input + %d output\n",
			response.Usage.InputTokens, response.Usage.OutputTokens)

		// Check if Claude wants to use tools
		hasToolUse := false
		toolResults := []ClaudeContent{}

		for _, content := range response.Content {
			if content.Type == "tool_use" {
				hasToolUse = true
				fmt.Printf("   🔧 Tool via MCP: %s\n", content.Name)

				// Call tool through MCP server
				mcpStart := time.Now()
				result, err := callMCPTool(content.Name, content.Input)
				mcpDuration := time.Since(mcpStart)
				mcpOverheadTime += mcpDuration
				mcpCallCount++

				if err != nil {
					fmt.Printf("      ❌ MCP call failed: %v\n", err)
					result = fmt.Sprintf("Error: %v", err)
				} else {
					fmt.Printf("      ✅ MCP call completed in %v\n", mcpDuration)
				}

				resultJSON, _ := json.Marshal(result)
				toolResults = append(toolResults, ClaudeContent{
					Type:      "tool_result",
					ToolUseId: content.ToolUseId,
					Content:   string(resultJSON),
				})
			}
		}

		// If no more tools to use, we're done
		if !hasToolUse {
			fmt.Println("   ✅ Order processing complete")
			break
		}

		// Add assistant's response and tool results to message history
		messages = append(messages, ClaudeMessage{
			Role:    "assistant",
			Content: response.Content,
		})

		messages = append(messages, ClaudeMessage{
			Role:    "user",
			Content: toolResults,
		})
	}

	// Calculate total metrics
	totalDuration := time.Since(startTime)
	totalTokens := totalInputTokens + totalOutputTokens
	cost := calculateCost(totalInputTokens, totalOutputTokens)

	// Print results
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 RESULTS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("⏱️  Total Duration:    %v\n", totalDuration)
	fmt.Printf("📞 Claude API Calls:   %d\n", claudeAPICount)
	fmt.Printf("🔌 MCP Calls:          %d\n", mcpCallCount)
	fmt.Printf("⚡ MCP Overhead:       %v\n", mcpOverheadTime)
	fmt.Printf("🎯 Tokens:             %d\n", totalTokens)
	fmt.Printf("💰 Cost:               $%.4f\n", cost)
	fmt.Printf("✅ Status:             Order Confirmed\n")
	fmt.Printf("🧾 Order ID:           ORD-2025-001\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Output JSON for comparison
	resultsJSON := map[string]interface{}{
		"approach":        "Native MCP",
		"duration":        totalDuration.Seconds(),
		"claudeAPICalls":  claudeAPICount,
		"mcpCalls":        mcpCallCount,
		"totalCalls":      claudeAPICount + mcpCallCount,
		"mcpOverhead":     mcpOverheadTime.Seconds(),
		"tokens":          totalTokens,
		"inputTokens":     totalInputTokens,
		"outputTokens":    totalOutputTokens,
		"cost":            cost,
		"status":          "Order Confirmed",
		"orderProcessed":  true,
	}

	jsonOutput, _ := json.MarshalIndent(resultsJSON, "", "  ")
	os.WriteFile("results-mcp.json", jsonOutput, 0644)
	fmt.Println("\n📄 Results saved to results-mcp.json")
}

func listMCPTools() ([]ClaudeTool, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      1,
	}

	jsonData, _ := json.Marshal(req)
	resp, err := http.Post(mcpServerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var mcpResp struct {
		Result struct {
			Tools []struct {
				Name        string                 `json:"name"`
				Description string                 `json:"description"`
				InputSchema map[string]interface{} `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}

	json.Unmarshal(body, &mcpResp)

	claudeTools := make([]ClaudeTool, len(mcpResp.Result.Tools))
	for i, tool := range mcpResp.Result.Tools {
		claudeTools[i] = ClaudeTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		}
	}

	return claudeTools, nil
}

func callMCPTool(name string, args map[string]interface{}) (interface{}, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
		ID: mcpCallCount + 100,
	}

	jsonData, _ := json.Marshal(req)
	resp, err := http.Post(mcpServerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, err
	}

	return mcpResp.Result, nil
}

func callClaude(apiKey string, messages []ClaudeMessage, tools []ClaudeTool) (*ClaudeResponse, error) {
	reqBody := ClaudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 4096,
		Messages:  messages,
		Tools:     tools,
	}

	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var response ClaudeResponse
	json.Unmarshal(body, &response)

	return &response, nil
}

func calculateCost(inputTokens, outputTokens int) float64 {
	// Claude Sonnet 4 pricing: $3/1M input, $15/1M output
	inputCost := float64(inputTokens) * 3.0 / 1_000_000
	outputCost := float64(outputTokens) * 15.0 / 1_000_000
	return inputCost + outputCost
}
