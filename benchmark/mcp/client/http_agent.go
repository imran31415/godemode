package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/imran31415/godemode/benchmark/agents"
	"github.com/imran31415/godemode/benchmark/llm"
	"github.com/imran31415/godemode/benchmark/mcp/protocol"
	"github.com/imran31415/godemode/benchmark/scenarios"
)

// HTTPMCPAgent uses MCP tools via HTTP to execute tasks
type HTTPMCPAgent struct {
	client  *HTTPMCPClient
	claude  *llm.ClaudeClient
	logger  *agents.FunctionCallLogger
	tools   []protocol.Tool
}

// NewHTTPMCPAgent creates a new agent that uses HTTP MCP tools
func NewHTTPMCPAgent(mcpClient *HTTPMCPClient) (*HTTPMCPAgent, error) {
	// Get tools from MCP server
	fmt.Println("Requesting tools list from MCP HTTP server...")
	tools, err := mcpClient.ListTools()
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	fmt.Printf("Received %d tools from MCP HTTP server\n", len(tools))

	return &HTTPMCPAgent{
		client: mcpClient,
		claude: llm.NewClaudeClient(),
		logger: &agents.FunctionCallLogger{},
		tools:  tools,
	}, nil
}

// RunTask executes a task using MCP HTTP tools
func (a *HTTPMCPAgent) RunTask(ctx context.Context, task scenarios.Task, env *scenarios.TestEnvironment) (*agents.AgentMetrics, error) {
	metrics := &agents.AgentMetrics{
		TaskName:  task.Name,
		StartTime: time.Now(),
	}

	fmt.Printf("Running MCP HTTP task with %d tools available\n", len(a.tools))

	// Convert MCP tools to Claude format
	claudeTools := a.convertToolsToClaudeFormat()

	// Create system prompt (same as native tool calling)
	systemPrompt := fmt.Sprintf(`You are an autonomous agent that completes tasks using the available tools.

Your task: %s

Instructions:
- Before calling a tool, analyze which tools are relevant and plan your approach
- Use tools to gather information and take actions as needed
- Think through the complete workflow and all required steps
- Work systematically through each phase of the task
- Continue until you have completed all necessary operations

You have access to tools for:
- Reading and sending emails
- Creating and managing support tickets
- Analyzing security events and threats
- Searching logs and configurations
- Managing user accounts and permissions
- Blocking IPs and enforcing security policies

Plan your approach and execute all necessary steps to fully complete the task.`, task.Description)

	userMessage := fmt.Sprintf(`Task: %s

Expected operations: approximately %d steps

Please analyze the task, plan your approach, and use the available tools to complete it fully. Think through each step before taking action.`, task.Name, task.ExpectedOps)

	// Call Claude with tools
	response, err := a.claude.CallWithTools(ctx, systemPrompt, userMessage, claudeTools)
	if err != nil {
		metrics.EndTime = time.Now()
		metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false
		metrics.Errors = append(metrics.Errors, fmt.Sprintf("Claude API call failed: %v", err))
		return metrics, err
	}

	metrics.TokensUsed += response.Usage.InputTokens + response.Usage.OutputTokens
	metrics.APICallCount++

	// Process the response and execute tools via MCP HTTP
	for {
		// Check stop reason
		if response.StopReason != "tool_use" {
			fmt.Printf("[MCP HTTP Agent] Stopping, reason: %s\n", response.StopReason)
			break
		}

		// Execute all tool calls via MCP HTTP
		toolResults := []llm.ContentBlock{}
		for _, content := range response.Content {
			if content.Type == "tool_use" {
				metrics.OperationsCount++

				fmt.Printf("[MCP HTTP TOOL CALL %d] %s\n", metrics.OperationsCount, content.Name)

				// Call tool via MCP HTTP
				result, err := a.client.CallTool(content.Name, content.Input)

				// Log the call
				fc := agents.FunctionCall{
					Timestamp:  time.Now(),
					ToolName:   content.Name,
					Parameters: content.Input,
				}

				if err != nil || result.IsError {
					errorMsg := ""
					if err != nil {
						errorMsg = err.Error()
					} else if len(result.Content) > 0 {
						errorMsg = result.Content[0].Text
					}

					fc.Error = errorMsg
					toolResults = append(toolResults, llm.ContentBlock{
						Type:      "tool_result",
						ToolUseID: content.ID,
						Content:   fmt.Sprintf("Error: %s", errorMsg),
					})
				} else {
					// Extract result from MCP response
					resultText := ""
					if len(result.Content) > 0 {
						resultText = result.Content[0].Text
					}

					// Try to parse as JSON for logging
					var resultObj interface{}
					if err := json.Unmarshal([]byte(resultText), &resultObj); err == nil {
						fc.Result = resultObj
					} else {
						fc.Result = resultText
					}

					toolResults = append(toolResults, llm.ContentBlock{
						Type:      "tool_result",
						ToolUseID: content.ID,
						Content:   resultText,
					})
				}

				a.logger.Calls = append(a.logger.Calls, fc)
			}
		}

		// If no tool calls, we're done
		if len(toolResults) == 0 {
			break
		}

		// Continue the conversation with tool results
		response, err = a.claude.ContinueWithToolResults(ctx, systemPrompt, response.Content, toolResults, claudeTools)
		if err != nil {
			metrics.EndTime = time.Now()
			metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
			metrics.Success = false
			metrics.Errors = append(metrics.Errors, fmt.Sprintf("Tool continuation failed: %v", err))
			return metrics, err
		}

		metrics.TokensUsed += response.Usage.InputTokens + response.Usage.OutputTokens
		metrics.APICallCount++
	}

	metrics.EndTime = time.Now()
	metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Success = true

	return metrics, nil
}

// convertToolsToClaudeFormat converts MCP tools to Claude's tool format
func (a *HTTPMCPAgent) convertToolsToClaudeFormat() []llm.Tool {
	claudeTools := make([]llm.Tool, 0, len(a.tools))

	for _, mcpTool := range a.tools {
		properties := make(map[string]llm.Property)

		for name, prop := range mcpTool.InputSchema.Properties {
			properties[name] = llm.Property{
				Type:        prop.Type,
				Description: prop.Description,
				Enum:        prop.Enum,
			}
		}

		claudeTools = append(claudeTools, llm.Tool{
			Name:        mcpTool.Name,
			Description: mcpTool.Description,
			InputSchema: llm.InputSchema{
				Type:       mcpTool.InputSchema.Type,
				Properties: properties,
				Required:   mcpTool.InputSchema.Required,
			},
		})
	}

	return claudeTools
}

// GetFunctionCalls returns all logged function calls
func (a *HTTPMCPAgent) GetFunctionCalls() []agents.FunctionCall {
	return a.logger.Calls
}
