package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imran31415/godemode/pkg/executor"
	browsertools "github.com/imran31415/godemode/puppeteer-mcp-benchmark/generated"
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
	Required   []string               `json:"required"`
}

type ClaudeResponse struct {
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type AuditEntry struct {
	Timestamp    time.Time
	Type         string
	ToolName     string
	ToolArgs     string
	ToolResult   string
	Details      string
	InputTokens  int
	OutputTokens int
}

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

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("ANTHROPIC_API_KEY not set")
		os.Exit(1)
	}

	fmt.Println("======================================================================")
	fmt.Println("Puppeteer/Browser MCP Benchmark: CodeMode vs Native Tool Calling")
	fmt.Println("Target: umi-app.co")
	fmt.Println("======================================================================")

	// Get results directory
	cwd, _ := os.Getwd()
	resultsDir := filepath.Join(cwd, "results")
	os.MkdirAll(resultsDir, 0755)

	// Simple Scenario
	fmt.Println("\n======================================================================")
	fmt.Println("SCENARIO: Simple - Homepage Exploration")
	fmt.Println("======================================================================")
	fmt.Println("Expected operations: ~8-10 tool calls")

	simplePrompt := fmt.Sprintf(`You have access to browser automation tools (like Puppeteer).

Your task is to explore the homepage of umi-app.co and gather basic information.

Perform these steps:
1. Navigate to https://umi-app.co
2. Wait for the page to fully load (wait for body or main content)
3. Get the page title
4. Take a screenshot and save it as "%s/umi-homepage.png"
5. Count how many links (a tags) are on the page
6. Count how many buttons are on the page
7. Check if there's a login or sign-in element (look for text containing "login", "sign in", or similar)
8. Get all navigation link texts (if any nav exists)

Output the results at each step.`, resultsDir)

	fmt.Println("\nRunning CodeMode approach...")
	codeModeResult := runCodeModeBenchmark(apiKey, simplePrompt)
	printResult(codeModeResult)

	fmt.Println("\nRunning Native Tool Calling approach...")
	toolCallingResult := runToolCallingBenchmark(apiKey, simplePrompt)
	printResult(toolCallingResult)

	printComparison(codeModeResult, toolCallingResult)

	// Complex Scenario
	fmt.Println("\n======================================================================")
	fmt.Println("SCENARIO: Complex - Multi-Page Exploration & Interaction")
	fmt.Println("======================================================================")
	fmt.Println("Expected operations: ~20-25 tool calls")

	complexPrompt := fmt.Sprintf(`You have access to browser automation tools (like Puppeteer).

Your task is to thoroughly explore the umi-app.co website and gather comprehensive information.

Perform these steps:
1. Navigate to https://umi-app.co
2. Wait for page to fully load and take a screenshot
3. Get page title and extract all text content from the hero section
4. Find and count all buttons on the page
5. Find the "Get Started" button and click it
6. Wait for any navigation or modal to appear
7. Take a screenshot of the new state
8. Extract all visible text from the current view
9. Look for any form inputs and count them
10. Check for social media links (look for app store links)
11. Navigate back to the main page if needed
12. Scroll down to see more content
13. Take a full-page screenshot
14. Count all images on the page
15. Get the URLs of the first 5 images (src attributes)
16. Find any footer content
17. Extract navigation structure (if any menus exist)
18. Check for any videos or media elements
19. Get all external links
20. Summarize findings

Save screenshots to: %s/
Output detailed results at each step.`, resultsDir)

	// Re-initialize browser for complex scenario
	browsertools.CloseBrowser()
	fmt.Println("\nRunning CodeMode approach...")
	codeModeComplex := runCodeModeBenchmark(apiKey, complexPrompt)
	printResult(codeModeComplex)

	fmt.Println("\nRunning Native Tool Calling approach...")
	browsertools.CloseBrowser()
	toolCallingComplex := runToolCallingBenchmark(apiKey, complexPrompt)
	printResult(toolCallingComplex)

	printComparison(codeModeComplex, toolCallingComplex)

	// Clean up browser
	browsertools.CloseBrowser()
}

func runCodeModeBenchmark(apiKey, prompt string) BenchmarkResult {
	start := time.Now()
	result := BenchmarkResult{
		Approach: "CodeMode",
		AuditLog: []AuditEntry{},
	}

	// Initialize browser for the registry
	browsertools.InitBrowser()
	registry := browsertools.NewRegistry()

	// Create the code generation prompt with explicit examples and strict typing
	codeGenPrompt := fmt.Sprintf(`%s

%s

Generate a complete, valid Go program that accomplishes the task above.

CRITICAL REQUIREMENTS - YOU MUST FOLLOW THESE EXACTLY:

1. Package and imports:
   package main
   import "fmt"

2. Main function structure:
   func main() {
       // Your code here
   }

3. Tool calls MUST use this EXACT syntax - registry.Call():
   result, err := registry.Call("tool_name", map[string]interface{}{
       "param1": value1,
       "param2": value2,
   })

4. The variable "registry" is pre-defined - DO NOT redeclare it
5. Use ONLY fmt.Println for output - no other packages available
6. Handle all errors with if err != nil checks

COMPLETE WORKING EXAMPLES:

// Navigate to a URL
result, err := registry.Call("navigate", map[string]interface{}{
    "url": "https://example.com",
})
if err != nil {
    fmt.Println("Navigation error:", err)
    return
}
fmt.Println("Navigated to:", result)

// Take a screenshot
result, err = registry.Call("screenshot", map[string]interface{}{
    "filename":  "/path/to/screenshot.png",
    "full_page": true,
})
if err != nil {
    fmt.Println("Screenshot error:", err)
} else {
    fmt.Println("Screenshot saved:", result)
}

// Count elements
result, err = registry.Call("count_elements", map[string]interface{}{
    "selector": "a",
})
if err != nil {
    fmt.Println("Count error:", err)
} else {
    fmt.Println("Element count:", result)
}

// Get text from element
result, err = registry.Call("get_text", map[string]interface{}{
    "selector": "body",
})
if err != nil {
    fmt.Println("Get text error:", err)
} else {
    fmt.Println("Text content:", result)
}

// Click an element
result, err = registry.Call("click", map[string]interface{}{
    "selector": "button.submit",
})
if err != nil {
    fmt.Println("Click error:", err)
}

// Wait/sleep
result, err = registry.Call("sleep", map[string]interface{}{
    "milliseconds": 2000,
})

// Evaluate JavaScript
result, err = registry.Call("evaluate", map[string]interface{}{
    "script": "(function() { return document.title; })()",
})

FORBIDDEN - DO NOT USE:
- registryCall() - WRONG
- registry.call() - WRONG (lowercase)
- Call() without registry - WRONG
- Any imports except "fmt"
- Any packages like "log", "time", "strings"

Now generate the complete Go program:`, prompt, browsertools.GetToolDocumentation())

	// Log API call
	result.AuditLog = append(result.AuditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "API_CALL",
		Details:   "Sending prompt to Claude for code generation",
	})

	// Call Claude API for code generation
	code, inputTokens, outputTokens, err := callClaudeForCode(apiKey, codeGenPrompt)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	result.InputTokens = inputTokens
	result.OutputTokens = outputTokens
	result.TotalTokens = inputTokens + outputTokens
	result.APICallCount = 1
	result.GeneratedCode = code

	// Log response
	result.AuditLog = append(result.AuditLog, AuditEntry{
		Timestamp:    time.Now(),
		Type:         "API_RESPONSE",
		Details:      fmt.Sprintf("Received code generation response (tokens: in=%d, out=%d)", inputTokens, outputTokens),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	})

	// Count tool calls in generated code
	toolCallCount := strings.Count(code, "registry.Call(")
	result.AuditLog = append(result.AuditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "CODE_ANALYSIS",
		Details:   fmt.Sprintf("Generated code contains %d tool calls", toolCallCount),
	})

	// Execute the code
	result.AuditLog = append(result.AuditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "EXECUTION",
		Details:   "Starting code execution via Yaegi interpreter",
	})

	exec := executor.NewInterpreterExecutor()

	// Execute with registry callback for audit logging
	execResult, err := exec.ExecuteGeneratedCode(
		context.Background(),
		code,
		120*time.Second,
		func(toolName string, args map[string]interface{}) (interface{}, error) {
			toolResult, toolErr := registry.Call(toolName, args)

			argsJSON, _ := json.Marshal(args)
			resultJSON, _ := json.Marshal(toolResult)

			result.ToolCalls++
			result.AuditLog = append(result.AuditLog, AuditEntry{
				Timestamp:  time.Now(),
				Type:       "TOOL_CALL",
				ToolName:   toolName,
				ToolArgs:   truncateString(string(argsJSON), 100),
				ToolResult: truncateString(string(resultJSON), 200),
				Details:    fmt.Sprintf("Executed tool call #%d", result.ToolCalls),
			})

			return toolResult, toolErr
		},
	)

	if err != nil {
		result.Error = err.Error()
	} else if execResult.Error != "" {
		result.Error = execResult.Error
	} else {
		result.Success = true
		fmt.Println(execResult.Stdout)
	}

	result.AuditLog = append(result.AuditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "EXECUTION_COMPLETE",
		Details:   fmt.Sprintf("Execution completed with %d tool calls", result.ToolCalls),
	})

	result.Duration = time.Since(start)
	result.EstimatedCost = estimateCost(result.InputTokens, result.OutputTokens)

	return result
}

func runToolCallingBenchmark(apiKey, prompt string) BenchmarkResult {
	start := time.Now()
	result := BenchmarkResult{
		Approach: "ToolCalling",
		AuditLog: []AuditEntry{},
	}

	// Initialize browser
	browsertools.InitBrowser()
	registry := browsertools.NewRegistry()

	// Define tools for the API
	tools := getBrowserTools()

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	for {
		result.APICallCount++

		result.AuditLog = append(result.AuditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "API_CALL",
			Details:   fmt.Sprintf("API call #%d", result.APICallCount),
		})

		response, err := callClaudeWithTools(apiKey, messages, tools)
		if err != nil {
			result.Error = err.Error()
			result.Duration = time.Since(start)
			return result
		}

		result.InputTokens += response.Usage.InputTokens
		result.OutputTokens += response.Usage.OutputTokens

		result.AuditLog = append(result.AuditLog, AuditEntry{
			Timestamp:    time.Now(),
			Type:         "API_RESPONSE",
			Details:      fmt.Sprintf("Response #%d (stop_reason: %s) (tokens: in=%d, out=%d)", result.APICallCount, response.StopReason, response.Usage.InputTokens, response.Usage.OutputTokens),
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
		})

		// Process response
		var toolResults []ContentBlock
		for _, block := range response.Content {
			if block.Type == "tool_use" {
				result.ToolCalls++

				args, _ := block.Input.(map[string]interface{})
				argsJSON, _ := json.Marshal(args)

				result.AuditLog = append(result.AuditLog, AuditEntry{
					Timestamp: time.Now(),
					Type:      "TOOL_CALL",
					ToolName:  block.Name,
					ToolArgs:  truncateString(string(argsJSON), 100),
					Details:   fmt.Sprintf("Tool call #%d", result.ToolCalls),
				})

				// Execute the tool
				toolResult, toolErr := registry.Call(block.Name, args)

				var resultContent string
				if toolErr != nil {
					resultContent = fmt.Sprintf("Error: %v", toolErr)
				} else {
					resultJSON, _ := json.Marshal(toolResult)
					resultContent = string(resultJSON)
				}

				result.AuditLog = append(result.AuditLog, AuditEntry{
					Timestamp:  time.Now(),
					Type:       "TOOL_RESULT",
					ToolName:   block.Name,
					ToolResult: truncateString(resultContent, 200),
				})

				toolResults = append(toolResults, ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   resultContent,
				})
			}
		}

		if response.StopReason == "end_turn" {
			result.Success = true
			break
		}

		if len(toolResults) == 0 {
			break
		}

		// Add assistant response and tool results to messages
		messages = append(messages, Message{
			Role:    "assistant",
			Content: response.Content,
		})
		messages = append(messages, Message{
			Role:    "user",
			Content: toolResults,
		})

		// Safety limit
		if result.APICallCount > 50 {
			result.Error = "Too many API calls"
			break
		}
	}

	result.TotalTokens = result.InputTokens + result.OutputTokens
	result.Duration = time.Since(start)
	result.EstimatedCost = estimateCost(result.InputTokens, result.OutputTokens)

	return result
}

func callClaudeForCode(apiKey, prompt string) (string, int, int, error) {
	reqBody := ClaudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 8192,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var response ClaudeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", 0, 0, err
	}

	// Extract code from response
	var fullText string
	for _, block := range response.Content {
		if block.Type == "text" {
			fullText += block.Text
			code := extractGoCode(block.Text)
			if code != "" {
				return code, response.Usage.InputTokens, response.Usage.OutputTokens, nil
			}
		}
	}

	// Debug: show truncated response if no code found
	if len(fullText) > 200 {
		fullText = fullText[:200] + "..."
	}
	return "", response.Usage.InputTokens, response.Usage.OutputTokens, fmt.Errorf("no code found in response. Response preview: %s", fullText)
}

func callClaudeWithTools(apiKey string, messages []Message, tools []Tool) (*ClaudeResponse, error) {
	reqBody := ClaudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 4096,
		Messages:  messages,
		Tools:     tools,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var response ClaudeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func extractGoCode(text string) string {
	// Look for code blocks
	if idx := strings.Index(text, "```go"); idx != -1 {
		start := idx + 5
		if end := strings.Index(text[start:], "```"); end != -1 {
			return strings.TrimSpace(text[start : start+end])
		}
	}
	if idx := strings.Index(text, "```"); idx != -1 {
		start := idx + 3
		// Skip language identifier if present
		if newline := strings.Index(text[start:], "\n"); newline != -1 {
			start = start + newline + 1
		}
		if end := strings.Index(text[start:], "```"); end != -1 {
			return strings.TrimSpace(text[start : start+end])
		}
	}
	return ""
}

func getBrowserTools() []Tool {
	return []Tool{
		{
			Name:        "navigate",
			Description: "Navigate to a URL and optionally wait for a selector",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"url":           map[string]string{"type": "string", "description": "URL to navigate to"},
					"wait_selector": map[string]string{"type": "string", "description": "CSS selector to wait for (optional)"},
				},
				Required: []string{"url"},
			},
		},
		{
			Name:        "get_title",
			Description: "Get the page title",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
				Required:   []string{},
			},
		},
		{
			Name:        "get_url",
			Description: "Get the current page URL",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
				Required:   []string{},
			},
		},
		{
			Name:        "screenshot",
			Description: "Take a screenshot of the page",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"filename":  map[string]string{"type": "string", "description": "Filename to save screenshot"},
					"full_page": map[string]string{"type": "boolean", "description": "Take full page screenshot"},
				},
				Required: []string{},
			},
		},
		{
			Name:        "get_text",
			Description: "Get text content from an element",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"selector": map[string]string{"type": "string", "description": "CSS selector"},
				},
				Required: []string{"selector"},
			},
		},
		{
			Name:        "click",
			Description: "Click on an element",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"selector": map[string]string{"type": "string", "description": "CSS selector"},
				},
				Required: []string{"selector"},
			},
		},
		{
			Name:        "type",
			Description: "Type text into an input field",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"selector": map[string]string{"type": "string", "description": "CSS selector"},
					"text":     map[string]string{"type": "string", "description": "Text to type"},
				},
				Required: []string{"selector", "text"},
			},
		},
		{
			Name:        "wait_for_selector",
			Description: "Wait for an element to appear",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"selector": map[string]string{"type": "string", "description": "CSS selector"},
					"timeout":  map[string]string{"type": "integer", "description": "Timeout in seconds"},
				},
				Required: []string{"selector"},
			},
		},
		{
			Name:        "count_elements",
			Description: "Count elements matching a selector",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"selector": map[string]string{"type": "string", "description": "CSS selector"},
				},
				Required: []string{"selector"},
			},
		},
		{
			Name:        "get_all_text",
			Description: "Get text from all elements matching a selector",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"selector": map[string]string{"type": "string", "description": "CSS selector"},
					"limit":    map[string]string{"type": "integer", "description": "Maximum number of elements"},
				},
				Required: []string{"selector"},
			},
		},
		{
			Name:        "element_exists",
			Description: "Check if an element exists on the page",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"selector": map[string]string{"type": "string", "description": "CSS selector"},
				},
				Required: []string{"selector"},
			},
		},
		{
			Name:        "evaluate",
			Description: "Execute JavaScript in the page context",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"script": map[string]string{"type": "string", "description": "JavaScript to execute"},
				},
				Required: []string{"script"},
			},
		},
		{
			Name:        "sleep",
			Description: "Wait for a specified time",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"milliseconds": map[string]string{"type": "integer", "description": "Time to wait in milliseconds"},
				},
				Required: []string{"milliseconds"},
			},
		},
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func estimateCost(inputTokens, outputTokens int) float64 {
	// Claude Sonnet pricing: $3/1M input, $15/1M output
	inputCost := float64(inputTokens) / 1000000 * 3
	outputCost := float64(outputTokens) / 1000000 * 15
	return inputCost + outputCost
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
	fmt.Println("\n    --- Audit Log ---")
	for i, entry := range r.AuditLog {
		timestamp := entry.Timestamp.Format("15:04:05.000")
		fmt.Printf("    [%s] %d. %s: %s\n", timestamp, i+1, entry.Type, entry.Details)
		if entry.ToolName != "" {
			fmt.Printf("                    Tool: %s\n", entry.ToolName)
			if entry.ToolArgs != "" {
				fmt.Printf("                    Args: %s\n", entry.ToolArgs)
			}
		}
		if entry.ToolResult != "" {
			fmt.Printf("    %s. TOOL_RESULT: %s\n", "", entry.ToolName)
			fmt.Printf("                    Result: %s\n", entry.ToolResult)
		}
	}

	// Print generated code for CodeMode
	if r.GeneratedCode != "" {
		fmt.Printf("\n    --- Generated Code (truncated) ---\n")
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

		// Save full generated code to file (in results directory to avoid build conflicts)
		filename := fmt.Sprintf("results/generated-code-%s.go.txt", strings.ToLower(strings.ReplaceAll(r.Approach, " ", "-")))
		os.WriteFile(filename, []byte(r.GeneratedCode), 0644)
		fmt.Printf("    Full code saved to: %s\n\n", filename)
	}
}

func printComparison(codeMode, toolCalling BenchmarkResult) {
	fmt.Println("\n======================================================================")
	fmt.Println("COMPARISON")
	fmt.Println("======================================================================")

	if codeMode.Duration > 0 && toolCalling.Duration > 0 {
		speedup := float64(toolCalling.Duration) / float64(codeMode.Duration)
		tokenReduction := 100.0 * (1.0 - float64(codeMode.TotalTokens)/float64(toolCalling.TotalTokens))
		costSavings := 100.0 * (1.0 - codeMode.EstimatedCost/toolCalling.EstimatedCost)

		fmt.Printf("\nCodeMode vs Tool Calling:\n")
		fmt.Printf("  Speed:         %.2fx faster\n", speedup)
		fmt.Printf("  Tokens:        %.1f%% fewer tokens\n", tokenReduction)
		fmt.Printf("  Cost:          %.1f%% cheaper\n", costSavings)
		fmt.Printf("  API Calls:     %d vs %d\n", codeMode.APICallCount, toolCalling.APICallCount)
	}
}
