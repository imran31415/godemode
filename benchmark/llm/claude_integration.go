package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ClaudeClient handles communication with Claude API
type ClaudeClient struct {
	apiKey       string
	httpClient   *http.Client
	baseURL      string
	model        string        // Model to use for requests
	lastCallTime time.Time     // Track last API call for rate limiting
	minDelay     time.Duration // Minimum delay between API calls
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient() *ClaudeClient {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: ANTHROPIC_API_KEY not set. Using mock responses.")
	}

	// Get model from environment or use default
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514" // Default to Sonnet 4
	}

	fmt.Printf("Using Claude model: %s\n", model)

	return &ClaudeClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		baseURL:    "https://api.anthropic.com/v1/messages",
		model:      model,
		minDelay:   2 * time.Second, // 2 second minimum delay between API calls
	}
}

// Message represents a Claude API message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents the API request
type ClaudeRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
}

// Tool represents a tool definition for Claude
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines the parameters for a tool
type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]Property    `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// Property defines a single parameter
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// ContentBlock represents different types of content in messages
type ContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
}

// ClaudeResponse represents the API response
type ClaudeResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason,omitempty"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// GenerateCode generates Go code using Claude API
func (c *ClaudeClient) GenerateCode(ctx context.Context, systemPrompt, userPrompt string) (string, int, error) {
	// If no API key, return mock code
	if c.apiKey == "" {
		return c.getMockCode(userPrompt), 500, nil
	}

	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 8000, // Increased for complex multi-step workflows
		System:    systemPrompt,
		Messages: []Message{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.7,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", 0, fmt.Errorf("no content in response")
	}

	code := response.Content[0].Text
	totalTokens := response.Usage.InputTokens + response.Usage.OutputTokens

	// Clean the code - remove markdown code blocks if present
	code = cleanCode(code)

	return code, totalTokens, nil
}

// getMockCode returns mock code for testing without API key
func (c *ClaudeClient) getMockCode(prompt string) string {
	// Simple mock based on prompt keywords
	if contains(prompt, "simple") || contains(prompt, "email-to-ticket") {
		return `package main

import "fmt"

func main() {
	fmt.Println("Reading email: support_001")
	fmt.Println("Created ticket TICKET-001 with priority 3")
	fmt.Println("Sent confirmation email")
	fmt.Println("Task completed successfully")
}
`
	}

	if contains(prompt, "medium") || contains(prompt, "investigate") {
		return `package main

import "fmt"

func main() {
	fmt.Println("Reading email with error code ERR-500-XYZ")
	fmt.Println("Searching logs for ERR-500-XYZ")
	fmt.Println("Found: OutOfMemory exception during file upload")
	fmt.Println("Finding similar issues in knowledge graph")
	fmt.Println("Found 2 similar issues: ISS-001, ISS-002")
	fmt.Println("Created ticket TICKET-002 with priority 4")
	fmt.Println("Added tags: [memory OutOfMemory]")
	fmt.Println("Linked ticket to 2 similar issues")
	fmt.Println("Sent notification email")
	fmt.Println("Task completed successfully")
}
`
	}

	// Complex/default
	return `package main

import "fmt"

func main() {
	fmt.Println("Reading urgent email about upload feature")
	fmt.Println("Searching logs for ERR-UPLOAD-500")
	fmt.Println("Found multiple errors")
	fmt.Println("Searching knowledge graph for similar issues")
	fmt.Println("Found issue: ISS-UPLOAD-MEM")
	fmt.Println("Found solution: SOL-001")
	fmt.Println("Reading feature_flags.json")
	fmt.Println("Reading known_issues.yaml")
	fmt.Println("Created ticket TICKET-003 with priority 5")
	fmt.Println("Tags: [memory upload urgent]")
	fmt.Println("Linked ticket to ISS-UPLOAD-MEM")
	fmt.Println("Added auto-suggested solution to ticket")
	fmt.Println("Sent email with solution details")
	fmt.Println("Complex task completed successfully")
}
`
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// cleanCode removes markdown code blocks and extra text from generated code
func cleanCode(code string) string {
	// Remove markdown code blocks
	// Pattern 1: ```go\ncode\n```
	// Pattern 2: ```\ncode\n```

	codeBytes := bytes.TrimSpace([]byte(code))

	// Check if code starts with markdown code block
	if bytes.HasPrefix(codeBytes, []byte("```")) {
		lines := bytes.Split(codeBytes, []byte("\n"))

		// Find the start line (skip ```go or ```)
		startIdx := 1

		// Find the end line (the closing ```)
		endIdx := len(lines) - 1
		for i := len(lines) - 1; i >= 1; i-- {
			if bytes.Equal(bytes.TrimSpace(lines[i]), []byte("```")) {
				endIdx = i
				break
			}
		}

		// Extract the code between the markers
		if endIdx > startIdx {
			codeLines := lines[startIdx:endIdx]
			return string(bytes.Join(codeLines, []byte("\n")))
		}
	}

	return string(codeBytes)
}

// CallWithTools calls Claude with native tool use support
func (c *ClaudeClient) CallWithTools(ctx context.Context, systemPrompt string, userMessage string, tools []Tool) (*ClaudeResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("API key required for tool calling")
	}

	request := ClaudeRequest{
		Model:       c.model,
		MaxTokens:   4096,
		System:      systemPrompt,
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Tools:       tools,
		Temperature: 1.0,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug: Print request size
	fmt.Printf("[DEBUG] Request body size: %d bytes, %d tools\n", len(requestBody), len(tools))

	// Validate tool schemas before sending
	for i, tool := range tools {
		if len(tool.InputSchema.Properties) == 0 {
			fmt.Printf("[WARNING] Tool %d (%s) has no properties\n", i, tool.Name)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Rate limiting: wait if we called API too recently
	if !c.lastCallTime.IsZero() {
		elapsed := time.Since(c.lastCallTime)
		if elapsed < c.minDelay {
			waitTime := c.minDelay - elapsed
			fmt.Printf("[RATE LIMIT] Waiting %v before API call...\n", waitTime)
			time.Sleep(waitTime)
		}
	}

	fmt.Println("[DEBUG] Sending request to Claude API...")
	c.lastCallTime = time.Now() // Record this call time

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[DEBUG] Got response with status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Print first 500 chars of error
		errorPreview := string(body)
		if len(errorPreview) > 500 {
			errorPreview = errorPreview[:500] + "..."
		}
		fmt.Printf("[DEBUG] Error response: %s\n", errorPreview)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("[DEBUG] Response decoded successfully, stop_reason: %s\n", response.StopReason)

	return &response, nil
}

// ContinueWithToolResults continues a conversation by sending tool results back to Claude
func (c *ClaudeClient) ContinueWithToolResults(ctx context.Context, systemPrompt string, conversationHistory []ContentBlock, toolResults []ContentBlock, tools []Tool) (*ClaudeResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("API key required for tool calling")
	}

	// Validate and fix conversation history - ensure all tool_use blocks have input field
	validHistory := make([]ContentBlock, 0, len(conversationHistory))
	for _, block := range conversationHistory {
		if block.Type == "tool_use" {
			// Ensure input field exists (Claude API requires it)
			if block.Input == nil {
				block.Input = make(map[string]interface{})
			}
		}
		validHistory = append(validHistory, block)
	}

	// Build messages with conversation history
	// Message 1: User's original message (from history)
	// Message 2: Assistant's response with tool use
	// Message 3: User's tool results
	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": "Begin the task. Call all necessary tools.",
		},
		{
			"role":    "assistant",
			"content": validHistory,
		},
		{
			"role":    "user",
			"content": toolResults,
		},
	}

	request := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 4096,
		"system":     systemPrompt,
		"messages":   messages,
		"tools":      tools,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Printf("[DEBUG] Continuation request size: %d bytes\n", len(requestBody))

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Rate limiting: wait if we called API too recently
	if !c.lastCallTime.IsZero() {
		elapsed := time.Since(c.lastCallTime)
		if elapsed < c.minDelay {
			waitTime := c.minDelay - elapsed
			fmt.Printf("[RATE LIMIT] Waiting %v before continuation call...\n", waitTime)
			time.Sleep(waitTime)
		}
	}

	fmt.Println("[DEBUG] Sending continuation request...")
	c.lastCallTime = time.Now() // Record this call time

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[DEBUG] Got continuation response with status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorPreview := string(body)
		if len(errorPreview) > 500 {
			errorPreview = errorPreview[:500] + "..."
		}
		fmt.Printf("[DEBUG] Error response: %s\n", errorPreview)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("[DEBUG] Continuation decoded, stop_reason: %s\n", response.StopReason)

	return &response, nil
}
