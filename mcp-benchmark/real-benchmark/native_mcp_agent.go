package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func NewMCPClient(serverURL string) *MCPClient {
	return &MCPClient{
		serverURL: serverURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *MCPClient) ListTools() ([]ClaudeTool, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	body, _ := json.Marshal(payload)
	resp, err := c.httpClient.Post(c.serverURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			Tools []struct {
				Name        string                 `json:"name"`
				Description string                 `json:"description"`
				InputSchema map[string]interface{} `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode tools: %w", err)
	}

	// Convert to Claude tools format
	claudeTools := make([]ClaudeTool, len(result.Result.Tools))
	for i, t := range result.Result.Tools {
		claudeTools[i] = ClaudeTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		}
	}

	return claudeTools, nil
}

func (c *MCPClient) CallTool(toolName string, args map[string]interface{}) (interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := c.httpClient.Post(c.serverURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Result interface{}            `json:"result"`
		Error  map[string]interface{} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode result: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("MCP error: %v", result.Error)
	}

	return result.Result, nil
}

// Native MCP Agent - uses Claude to orchestrate sequential tool calls
type NativeMCPAgent struct {
	claudeAPIKey string
	mcpClient    *MCPClient
	httpClient   *http.Client
}

func NewNativeMCPAgent(apiKey string, mcpServerURL string) *NativeMCPAgent {
	return &NativeMCPAgent{
		claudeAPIKey: apiKey,
		mcpClient:    NewMCPClient(mcpServerURL),
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (a *NativeMCPAgent) callClaude(messages []ClaudeMessage, tools []ClaudeTool) (*ClaudeResponse, error) {
	req := ClaudeRequest{
		Model:       "claude-sonnet-4-20250514",
		MaxTokens:   4096,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.7,
	}

	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.claudeAPIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode Claude response: %w", err)
	}

	return &claudeResp, nil
}

// BenchmarkResult tracks metrics for the Native MCP approach
type NativeMCPResult struct {
	TotalDuration    time.Duration
	APICallCount     int
	TotalInputTokens int
	TotalOutputTokens int
	ToolCallCount    int
	FinalOutput      string
	Success          bool
	Error            error
}

func (a *NativeMCPAgent) RunTask(ctx context.Context, task string) (*NativeMCPResult, error) {
	startTime := time.Now()
	result := &NativeMCPResult{
		Success: true,
	}

	// Get available tools
	tools, err := a.mcpClient.ListTools()
	if err != nil {
		result.Success = false
		result.Error = err
		return result, err
	}

	// First API call: Ask Claude to use tools
	messages := []ClaudeMessage{
		{
			Role:    "user",
			Content: task,
		},
	}

	resp, err := a.callClaude(messages, tools)
	if err != nil {
		result.Success = false
		result.Error = err
		return result, err
	}

	result.APICallCount++
	result.TotalInputTokens += resp.Usage.InputTokens
	result.TotalOutputTokens += resp.Usage.OutputTokens

	// Execute all tool calls
	toolResults := []map[string]interface{}{}
	for _, content := range resp.Content {
		if content.Type == "tool_use" {
			result.ToolCallCount++

			// Ensure Input is not nil
			input := content.Input
			if input == nil {
				input = make(map[string]interface{})
			}

			// Call the tool via MCP server
			toolResult, err := a.mcpClient.CallTool(content.Name, input)
			if err != nil {
				toolResults = append(toolResults, map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": content.ID,
					"is_error":    true,
					"content":     err.Error(),
				})
			} else {
				// Convert result to string for Claude
				resultJSON, _ := json.Marshal(toolResult)
				toolResults = append(toolResults, map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": content.ID,
					"content":     string(resultJSON),
				})
			}
		}
	}

	// Second API call: Ask Claude to summarize results
	// Ensure all tool_use blocks have input field set
	sanitizedContent := make([]interface{}, len(resp.Content))
	for i, content := range resp.Content {
		contentMap := map[string]interface{}{
			"type": content.Type,
		}
		if content.Type == "tool_use" {
			contentMap["id"] = content.ID
			contentMap["name"] = content.Name
			// Ensure input is always set, even if empty
			if content.Input != nil {
				contentMap["input"] = content.Input
			} else {
				contentMap["input"] = map[string]interface{}{}
			}
		} else if content.Type == "text" {
			contentMap["text"] = content.Text
		}
		sanitizedContent[i] = contentMap
	}

	messages = append(messages, ClaudeMessage{
		Role:    "assistant",
		Content: sanitizedContent,
	})
	messages = append(messages, ClaudeMessage{
		Role:    "user",
		Content: toolResults,
	})

	summaryResp, err := a.callClaude(messages, nil) // No tools for summary
	if err != nil {
		result.Success = false
		result.Error = err
		return result, err
	}

	result.APICallCount++
	result.TotalInputTokens += summaryResp.Usage.InputTokens
	result.TotalOutputTokens += summaryResp.Usage.OutputTokens

	// Extract final output
	for _, content := range summaryResp.Content {
		if content.Type == "text" {
			result.FinalOutput = content.Text
		}
	}

	result.TotalDuration = time.Since(startTime)
	return result, nil
}

