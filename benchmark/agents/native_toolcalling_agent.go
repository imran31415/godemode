package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/imran31415/godemode/benchmark/llm"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/tools"
)

// NativeToolCallingAgent uses Claude's native tool calling API
type NativeToolCallingAgent struct {
	registry        *tools.ToolRegistry
	client          *llm.ClaudeClient
	logger          *FunctionCallLogger
	currentTicketID string
}

// NewNativeToolCallingAgent creates a new agent using Claude's native tool calling
func NewNativeToolCallingAgent(env *scenarios.TestEnvironment) *NativeToolCallingAgent {
	return &NativeToolCallingAgent{
		registry: tools.NewToolRegistry(env),
		client:   llm.NewClaudeClient(),
		logger:   &FunctionCallLogger{},
	}
}

// RunTask executes a task using Claude's native tool calling API
func (a *NativeToolCallingAgent) RunTask(ctx context.Context, task scenarios.Task, env *scenarios.TestEnvironment) (*AgentMetrics, error) {
	metrics := &AgentMetrics{
		TaskName:  task.Name,
		StartTime: time.Now(),
	}

	// Convert tools to Claude format
	claudeTools := a.convertToolsToClaudeFormat()

	fmt.Printf("Registered %d tools for Claude\n", len(claudeTools))

	// Debug: Print first few tools
	fmt.Println("\nFirst 3 tools:")
	for i := 0; i < 3 && i < len(claudeTools); i++ {
		fmt.Printf("  %d. %s - %s\n", i+1, claudeTools[i].Name, claudeTools[i].Description)
		fmt.Printf("     Parameters: %d properties\n", len(claudeTools[i].InputSchema.Properties))
	}
	fmt.Println()

	// Create system prompt following Anthropic best practices
	// Allow thinking and analysis before tool calls
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
	response, err := a.client.CallWithTools(ctx, systemPrompt, userMessage, claudeTools)
	if err != nil {
		metrics.EndTime = time.Now()
		metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false
		metrics.Errors = append(metrics.Errors, fmt.Sprintf("Claude API call failed: %v", err))
		return metrics, err
	}

	metrics.TokensUsed += response.Usage.InputTokens + response.Usage.OutputTokens
	metrics.APICallCount++

	// Debug: Print response content
	fmt.Printf("\n[DEBUG] Response has %d content blocks:\n", len(response.Content))
	for i, content := range response.Content {
		fmt.Printf("  Block %d: type=%s\n", i, content.Type)
		if content.Type == "text" {
			preview := content.Text
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("    Text: %s\n", preview)
		} else if content.Type == "tool_use" {
			fmt.Printf("    Tool: %s\n", content.Name)
		}
	}

	// Process the response and execute tools
	for {
		// Check stop reason
		if response.StopReason != "tool_use" {
			fmt.Printf("[DEBUG] Stopping, reason: %s\n", response.StopReason)
			break
		}

		// Execute all tool calls in this response
		toolResults := []llm.ContentBlock{}
		for _, content := range response.Content {
			if content.Type == "tool_use" {
				metrics.OperationsCount++

				fmt.Printf("[TOOL CALL %d] %s\n", metrics.OperationsCount, content.Name)

				// Execute the tool
				result, err := a.executeToolByName(content.Name, content.Input)

				// Log the call
				fc := FunctionCall{
					Timestamp:  time.Now(),
					ToolName:   content.Name,
					Parameters: content.Input,
				}

				if err != nil {
					fc.Error = err.Error()
					toolResults = append(toolResults, llm.ContentBlock{
						Type:      "tool_result",
						ToolUseID: content.ID,
						Content:   fmt.Sprintf("Error: %v", err),
					})
				} else {
					fc.Result = result
					toolResults = append(toolResults, llm.ContentBlock{
						Type:      "tool_result",
						ToolUseID: content.ID,
						Content:   fmt.Sprintf("%v", result),
					})
				}

				a.logger.Calls = append(a.logger.Calls, fc)

				// Track ticket ID
				if content.Name == "createTicket" {
					if resultMap, ok := result.(map[string]interface{}); ok {
						if ticketID, ok := resultMap["ticketID"].(string); ok {
							a.currentTicketID = ticketID
						}
					}
				}
			}
		}

		// If no tool calls, we're done
		if len(toolResults) == 0 {
			break
		}

		// Continue the conversation with tool results
		response, err = a.continueWithToolResults(ctx, response, toolResults, systemPrompt, claudeTools)
		if err != nil {
			metrics.EndTime = time.Now()
			metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
			metrics.Success = false
			metrics.Errors = append(metrics.Errors, fmt.Sprintf("Tool continuation failed: %v", err))
			return metrics, err
		}

		metrics.TokensUsed += response.Usage.InputTokens + response.Usage.OutputTokens
		metrics.APICallCount++

		// Debug: Print continuation response
		fmt.Printf("\n[DEBUG] Continuation response has %d content blocks:\n", len(response.Content))
		for i, content := range response.Content {
			fmt.Printf("  Block %d: type=%s\n", i, content.Type)
			if content.Type == "text" {
				preview := content.Text
				if len(preview) > 300 {
					preview = preview[:300] + "..."
				}
				fmt.Printf("    Text: %s\n", preview)
			} else if content.Type == "tool_use" {
				fmt.Printf("    Tool: %s\n", content.Name)
			}
		}
	}

	metrics.EndTime = time.Now()
	metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Success = true

	return metrics, nil
}

// convertToolsToClaudeFormat converts tool registry to Claude's tool format
func (a *NativeToolCallingAgent) convertToolsToClaudeFormat() []llm.Tool {
	registryTools := a.registry.ListTools()
	claudeTools := make([]llm.Tool, 0, len(registryTools))

	for _, tool := range registryTools {
		properties := make(map[string]llm.Property)
		required := []string{}

		for _, param := range tool.Parameters {
			// Map tool types to JSON Schema types
			jsonType := "string"
			switch param.Type {
			case "string", "str":
				jsonType = "string"
			case "int", "integer", "number":
				jsonType = "number"
			case "bool", "boolean":
				jsonType = "boolean"
			case "array", "list":
				jsonType = "array"
			case "object", "map":
				jsonType = "object"
			default:
				jsonType = "string" // default to string
			}

			properties[param.Name] = llm.Property{
				Type:        jsonType,
				Description: fmt.Sprintf("%s parameter", param.Name),
			}
			if param.Required {
				required = append(required, param.Name)
			}
		}

		// Only add tool if it has valid parameters
		// Skip tools with no parameters or empty schema
		if len(properties) == 0 {
			// Add a dummy parameter for tools with no params
			properties["_unused"] = llm.Property{
				Type:        "string",
				Description: "Unused parameter",
			}
		}

		claudeTools = append(claudeTools, llm.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: llm.InputSchema{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		})
	}

	return claudeTools
}

// executeToolByName executes a tool by name with parameters
func (a *NativeToolCallingAgent) executeToolByName(toolName string, params map[string]interface{}) (interface{}, error) {
	tool, exists := a.registry.GetTool(toolName)
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	return tool.Function(params)
}

// continueWithToolResults continues the conversation with tool results
func (a *NativeToolCallingAgent) continueWithToolResults(ctx context.Context, previousResponse *llm.ClaudeResponse, toolResults []llm.ContentBlock, systemPrompt string, tools []llm.Tool) (*llm.ClaudeResponse, error) {
	// Call Claude API with tool results
	return a.client.ContinueWithToolResults(ctx, systemPrompt, previousResponse.Content, toolResults, tools)
}

// GetFunctionCalls returns all logged function calls
func (a *NativeToolCallingAgent) GetFunctionCalls() []FunctionCall {
	return a.logger.Calls
}
