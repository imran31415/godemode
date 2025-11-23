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
	GeneratedCode string
}

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("ANTHROPIC_API_KEY not set")
		os.Exit(1)
	}

	fmt.Println("======================================================================")
	fmt.Println("Puppeteer/Browser MCP Benchmark: Login Flow")
	fmt.Println("Target: GitHub.com")
	fmt.Println("======================================================================")

	// Get results directory
	cwd, _ := os.Getwd()
	resultsDir := filepath.Join(cwd, "results")
	os.MkdirAll(resultsDir, 0755)

	// Login Scenario
	fmt.Println("\n======================================================================")
	fmt.Println("SCENARIO: Login Flow")
	fmt.Println("======================================================================")
	fmt.Println("Expected operations: ~10-15 tool calls")

	loginPrompt := fmt.Sprintf(`You have access to browser automation tools (like Puppeteer).

Your task is to navigate to GitHub's login page and fill in the login form with the provided credentials.

Credentials:
- Username/Email: imoran21@gmail.com
- Password: WHO2like21!

Perform these steps:
1. Navigate to https://github.com/login
2. Wait for the page to fully load (wait for the login form to appear)
3. Take a screenshot of the login page
4. Find the username/email input field (id="login_field") and enter the username
5. Find the password input field (id="password") and enter the password
6. Take a screenshot showing the filled form
7. Find and click the "Sign in" button (name="commit")
8. Wait for the response (either login success or error message)
9. Take a screenshot of the result page
10. Check if login was successful or if there was an error
11. Report the final state with details

IMPORTANT: GitHub uses standard HTML form elements:
- Username input: <input id="login_field" name="login" type="text">
- Password input: <input id="password" name="password" type="password">
- Submit button: <input name="commit" type="submit" value="Sign in">

Save all screenshots to: %s/
Use descriptive filenames like "github_login_page.png", "github_form_filled.png", "github_result.png"

Output detailed results at each step.`, resultsDir)

	fmt.Println("\nRunning CodeMode approach...")
	codeModeResult := runCodeModeBenchmark(apiKey, loginPrompt)
	printResult(codeModeResult)

	// Close and reinitialize browser for Tool Calling
	browsertools.CloseBrowser()

	fmt.Println("\nRunning Native Tool Calling approach...")
	toolCallingResult := runToolCallingBenchmark(apiKey, loginPrompt)
	printResult(toolCallingResult)

	printComparison(codeModeResult, toolCallingResult)

	// Clean up browser
	browsertools.CloseBrowser()
}

func runCodeModeBenchmark(apiKey, prompt string) BenchmarkResult {
	start := time.Now()
	result := BenchmarkResult{
		Approach: "CodeMode",
	}

	// Initialize browser for the registry
	browsertools.InitBrowser()
	registry := browsertools.NewRegistry()

	// Create the code generation prompt with explicit examples and strict typing
	codeGenPrompt := fmt.Sprintf(`%s

%s

Generate a complete, valid Go program that accomplishes the login task above.

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
    "full_page": false,
})
if err != nil {
    fmt.Println("Screenshot error:", err)
} else {
    fmt.Println("Screenshot saved:", result)
}

// Click an element
result, err = registry.Call("click", map[string]interface{}{
    "selector": "button.submit",
})
if err != nil {
    fmt.Println("Click error:", err)
}

// Type text into an input field
result, err = registry.Call("type", map[string]interface{}{
    "selector": "input[type='email']",
    "text":     "user@example.com",
})
if err != nil {
    fmt.Println("Type error:", err)
}

// Wait for an element
result, err = registry.Call("wait_for_selector", map[string]interface{}{
    "selector": ".dashboard",
    "timeout":  15,
})

// Wait/sleep
result, err = registry.Call("sleep", map[string]interface{}{
    "milliseconds": 2000,
})

// Get text from element
result, err = registry.Call("get_text", map[string]interface{}{
    "selector": "body",
})

// Check if element exists
result, err = registry.Call("element_exists", map[string]interface{}{
    "selector": ".user-profile",
})

FORBIDDEN - DO NOT USE:
- registryCall() - WRONG
- registry.call() - WRONG (lowercase)
- Call() without registry - WRONG
- Any imports except "fmt"
- Any packages like "log", "time", "strings"

Now generate the complete Go program for the login flow:`, prompt, browsertools.GetToolDocumentation())

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

	// Count tool calls in generated code
	toolCallCount := strings.Count(code, "registry.Call(")
	fmt.Printf("Generated code contains %d tool calls\n", toolCallCount)

	// Execute the code
	exec := executor.NewInterpreterExecutor()

	execResult, err := exec.ExecuteGeneratedCode(
		context.Background(),
		code,
		180*time.Second, // 3 minute timeout for login flow
		func(toolName string, args map[string]interface{}) (interface{}, error) {
			toolResult, toolErr := registry.Call(toolName, args)
			result.ToolCalls++
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

	result.Duration = time.Since(start)
	result.EstimatedCost = estimateCost(result.InputTokens, result.OutputTokens)

	// Save generated code
	filename := "results/generated-login-codemode.go.txt"
	os.WriteFile(filename, []byte(result.GeneratedCode), 0644)
	fmt.Printf("Full code saved to: %s\n", filename)

	return result
}

func runToolCallingBenchmark(apiKey, prompt string) BenchmarkResult {
	start := time.Now()
	result := BenchmarkResult{
		Approach: "ToolCalling",
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

		response, err := callClaudeWithTools(apiKey, messages, tools)
		if err != nil {
			result.Error = err.Error()
			result.Duration = time.Since(start)
			return result
		}

		result.InputTokens += response.Usage.InputTokens
		result.OutputTokens += response.Usage.OutputTokens

		// Process response
		var toolResults []ContentBlock
		for _, block := range response.Content {
			if block.Type == "tool_use" {
				result.ToolCalls++

				args, _ := block.Input.(map[string]interface{})

				// Execute the tool
				toolResult, toolErr := registry.Call(block.Name, args)

				var resultContent string
				if toolErr != nil {
					resultContent = fmt.Sprintf("Error: %v", toolErr)
				} else {
					resultJSON, _ := json.Marshal(toolResult)
					resultContent = string(resultJSON)
				}

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
	for _, block := range response.Content {
		if block.Type == "text" {
			code := extractGoCode(block.Text)
			if code != "" {
				return code, response.Usage.InputTokens, response.Usage.OutputTokens, nil
			}
		}
	}

	return "", response.Usage.InputTokens, response.Usage.OutputTokens, fmt.Errorf("no code found in response")
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
		{
			Name:        "get_url",
			Description: "Get the current page URL",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
				Required:   []string{},
			},
		},
	}
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
