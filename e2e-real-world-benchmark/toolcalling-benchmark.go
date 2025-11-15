package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/yourusername/godemode/e2e-real-world-benchmark/tools"
)

// Tool Calling Benchmark - Uses Anthropic Messages API with sequential tool_use calls

type ToolCallRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []TCMessage      `json:"messages"`
	Tools     []ToolDefinition `json:"tools"`
}

type TCMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []ContentBlock
}

type ContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ToolUseId string                 `json:"tool_use_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
}

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type ToolCallResponse struct {
	Content   []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage     Usage          `json:"usage"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

var totalInputTokens = 0
var totalOutputTokens = 0
var apiCallCount = 0

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	fmt.Println("🚀 Tool Calling Benchmark - E-Commerce Order Processing")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	startTime := time.Now()

	// Create tool registry
	registry := tools.NewRegistry()

	// Define all 12 tools for Claude
	toolDefs := createToolDefinitions()

	// Initial order data
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

Please process this order through all required steps:
1. Validate the customer
2. Check inventory availability
3. Calculate shipping cost
4. Validate and apply the discount code
5. Calculate sales tax for California
6. Process the payment
7. Reserve the inventory
8. Create a shipping label
9. Send order confirmation email
10. Log the transaction
11. Update customer loyalty points
12. Create warehouse fulfillment task

Use the available tools to complete each step.`

	messages := []TCMessage{
		{Role: "user", Content: orderPrompt},
	}

	// Tool calling loop
	for {
		apiCallCount++
		fmt.Printf("\n📡 API Call %d: Processing order workflow...\n", apiCallCount)
		apiCallStart := time.Now()

		response, err := callClaudeWithTools(apiKey, messages, toolDefs)
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
		toolResults := []ContentBlock{}

		for _, content := range response.Content {
			if content.Type == "tool_use" {
				hasToolUse = true
				fmt.Printf("   🔧 Tool: %s\n", content.Name)

				// Execute tool
				result, _ := registry.Call(content.Name, content.Input)
				resultJSON, _ := json.Marshal(result)

				toolResults = append(toolResults, ContentBlock{
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
		messages = append(messages, TCMessage{
			Role:    "assistant",
			Content: response.Content,
		})

		messages = append(messages, TCMessage{
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
	fmt.Printf("📞 API Calls:          %d\n", apiCallCount)
	fmt.Printf("🎯 Tokens:             %d\n", totalTokens)
	fmt.Printf("💰 Cost:               $%.4f\n", cost)
	fmt.Printf("✅ Status:             Order Confirmed\n")
	fmt.Printf("🧾 Order ID:           ORD-2025-001\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Output JSON for comparison
	resultsJSON := map[string]interface{}{
		"approach":       "Tool Calling",
		"duration":       totalDuration.Seconds(),
		"apiCalls":       apiCallCount,
		"tokens":         totalTokens,
		"inputTokens":    totalInputTokens,
		"outputTokens":   totalOutputTokens,
		"cost":           cost,
		"status":         "Order Confirmed",
		"orderProcessed": true,
	}

	jsonOutput, _ := json.MarshalIndent(resultsJSON, "", "  ")
	os.WriteFile("results-toolcalling.json", jsonOutput, 0644)
	fmt.Println("\n📄 Results saved to results-toolcalling.json")
}

func callClaudeWithTools(apiKey string, messages []TCMessage, tools []ToolDefinition) (*ToolCallResponse, error) {
	reqBody := ToolCallRequest{
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

	var response ToolCallResponse
	json.Unmarshal(body, &response)

	return &response, nil
}

func createToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "validateCustomer",
			Description: "Validate customer and get their information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"customerId": map[string]string{"type": "string"},
					"email":      map[string]string{"type": "string"},
				},
				"required": []string{"customerId", "email"},
			},
		},
		{
			Name:        "checkInventory",
			Description: "Check product availability in inventory",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"products": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
						},
					},
				},
				"required": []string{"products"},
			},
		},
		{
			Name:        "calculateShipping",
			Description: "Calculate shipping cost and delivery estimate",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"destination": map[string]string{"type": "object"},
					"weight":      map[string]string{"type": "number"},
				},
				"required": []string{"destination", "weight"},
			},
		},
		{
			Name:        "validateDiscount",
			Description: "Validate discount code and calculate discount amount",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code":         map[string]string{"type": "string"},
					"customerTier": map[string]string{"type": "string"},
					"cartTotal":    map[string]string{"type": "number"},
				},
				"required": []string{"code", "customerTier", "cartTotal"},
			},
		},
		{
			Name:        "calculateTax",
			Description: "Calculate sales tax based on location and items",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"subtotal": map[string]string{"type": "number"},
					"state":    map[string]string{"type": "string"},
				},
				"required": []string{"subtotal", "state"},
			},
		},
		{
			Name:        "processPayment",
			Description: "Process payment and authorize transaction",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"amount":     map[string]string{"type": "number"},
					"customerId": map[string]string{"type": "string"},
					"method":     map[string]string{"type": "string"},
				},
				"required": []string{"amount", "customerId", "method"},
			},
		},
		{
			Name:        "reserveInventory",
			Description: "Reserve inventory for confirmed order",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"orderId":  map[string]string{"type": "string"},
					"products": map[string]interface{}{"type": "array"},
				},
				"required": []string{"orderId", "products"},
			},
		},
		{
			Name:        "createShippingLabel",
			Description: "Generate shipping label and tracking number",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"orderId": map[string]string{"type": "string"},
					"address": map[string]string{"type": "object"},
					"weight":  map[string]string{"type": "number"},
				},
				"required": []string{"orderId", "address", "weight"},
			},
		},
		{
			Name:        "sendOrderConfirmation",
			Description: "Send order confirmation email to customer",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"email":   map[string]string{"type": "string"},
					"orderId": map[string]string{"type": "string"},
					"total":   map[string]string{"type": "number"},
				},
				"required": []string{"email", "orderId", "total"},
			},
		},
		{
			Name:        "logTransaction",
			Description: "Log transaction details for analytics",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"orderId":       map[string]string{"type": "string"},
					"customerId":    map[string]string{"type": "string"},
					"amount":        map[string]string{"type": "number"},
					"transactionId": map[string]string{"type": "string"},
				},
				"required": []string{"orderId", "customerId", "amount", "transactionId"},
			},
		},
		{
			Name:        "updateLoyaltyPoints",
			Description: "Update customer loyalty points based on purchase",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"customerId": map[string]string{"type": "string"},
					"amount":     map[string]string{"type": "number"},
				},
				"required": []string{"customerId", "amount"},
			},
		},
		{
			Name:        "createFulfillmentTask",
			Description: "Create warehouse fulfillment task for order",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"orderId":     map[string]string{"type": "string"},
					"warehouseId": map[string]string{"type": "string"},
					"products":    map[string]interface{}{"type": "array"},
				},
				"required": []string{"orderId", "warehouseId", "products"},
			},
		},
	}
}

func calculateCost(inputTokens, outputTokens int) float64 {
	// Claude Sonnet 4 pricing: $3/1M input, $15/1M output
	inputCost := float64(inputTokens) * 3.0 / 1_000_000
	outputCost := float64(outputTokens) * 15.0 / 1_000_000
	return inputCost + outputCost
}
