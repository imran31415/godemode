package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/imran31415/godemode/benchmark/llm"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/tools"
)

// FunctionCallingAgent solves tasks using traditional tool calling
type FunctionCallingAgent struct {
	registry      *tools.ToolRegistry
	logger        *FunctionCallLogger
	currentTicketID string // Track the most recently created ticket
}

// FunctionCallLogger tracks all function calls
type FunctionCallLogger struct {
	Calls []FunctionCall
}

// FunctionCall represents a single tool invocation
type FunctionCall struct {
	Timestamp  time.Time
	ToolName   string
	Parameters map[string]interface{}
	Result     interface{}
	Error      string
	Duration   time.Duration
}

// NewFunctionCallingAgent creates a new function calling agent
func NewFunctionCallingAgent(env *scenarios.TestEnvironment) *FunctionCallingAgent {
	registry := tools.NewToolRegistry(env)

	return &FunctionCallingAgent{
		registry: registry,
		logger:   &FunctionCallLogger{Calls: []FunctionCall{}},
	}
}

// RunTask executes a task using function calling
func (a *FunctionCallingAgent) RunTask(ctx context.Context, task scenarios.Task, env *scenarios.TestEnvironment) (*AgentMetrics, error) {
	metrics := &AgentMetrics{
		TaskName:  task.Name,
		StartTime: time.Now(),
	}

	// Generate task plan
	plan, tokensUsed, err := a.planTask(ctx, task)
	if err != nil {
		metrics.EndTime = time.Now()
		metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false
		metrics.Errors = append(metrics.Errors, fmt.Sprintf("Planning failed: %v", err))
		return metrics, err
	}

	metrics.TokensUsed = tokensUsed
	metrics.APICallCount = 1 // Initial planning call

	// Reset state for new task
	a.currentTicketID = ""

	// Execute the plan step by step
	for i, step := range plan.Steps {
		// Simulate LLM deciding which tool to call
		toolCall, tokens, err := a.decideToolCall(ctx, step, task)
		metrics.TokensUsed += tokens
		metrics.APICallCount++

		if err != nil {
			metrics.EndTime = time.Now()
			metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
			metrics.Success = false
			metrics.Errors = append(metrics.Errors, fmt.Sprintf("Step %d failed: %v", i+1, err))
			return metrics, err
		}

		// Skip duplicate createTicket calls (only create one ticket per task)
		if toolCall.ToolName == "createTicket" && a.currentTicketID != "" {
			// Already created a ticket, skip this duplicate call
			continue
		}

		// Execute the tool call
		callStart := time.Now()
		result, err := a.executeToolCall(toolCall)
		callDuration := time.Since(callStart)

		// Log the function call
		fc := FunctionCall{
			Timestamp:  callStart,
			ToolName:   toolCall.ToolName,
			Parameters: toolCall.Parameters,
			Duration:   callDuration,
		}

		if err != nil {
			fc.Error = err.Error()
			a.logger.Calls = append(a.logger.Calls, fc)

			metrics.EndTime = time.Now()
			metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
			metrics.Success = false
			metrics.Errors = append(metrics.Errors, fmt.Sprintf("Tool %s failed: %v", toolCall.ToolName, err))
			return metrics, err
		}

		fc.Result = result
		a.logger.Calls = append(a.logger.Calls, fc)
		metrics.OperationsCount++

		// Track ticket ID if we just created a ticket
		if toolCall.ToolName == "createTicket" {
			if resultMap, ok := result.(map[string]interface{}); ok {
				if ticketID, ok := resultMap["ticketID"].(string); ok {
					a.currentTicketID = ticketID
				}
			}
		}
	}

	metrics.EndTime = time.Now()
	metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Success = true

	return metrics, nil
}

// TaskPlan represents the steps to complete a task
type TaskPlan struct {
	Steps []string
}

// ToolCall represents a tool invocation decision
type ToolCall struct {
	ToolName   string
	Parameters map[string]interface{}
}

// planTask creates a plan for the task by calling the LLM
func (a *FunctionCallingAgent) planTask(ctx context.Context, task scenarios.Task) (*TaskPlan, int, error) {
	// Check if we should use real LLM or fallback to hardcoded plans
	client := llm.NewClaudeClient()

	// If API key exists, use real LLM planning
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		return a.planTaskWithLLM(ctx, task, client)
	}

	// Fallback to hardcoded plans for testing without API key
	plan := &TaskPlan{}
	var tokens int

	switch task.Name {
	case "email-to-ticket":
		plan.Steps = []string{
			"Read the support email",
			"Extract key information from email",
			"Create a ticket with appropriate priority",
			"Send confirmation email",
		}
		tokens = 400

	case "investigate-with-logs":
		plan.Steps = []string{
			"Read the support email",
			"Extract error code from email",
			"Search logs for the error code",
			"Find similar issues in knowledge graph",
			"Create ticket with context",
			"Tag ticket appropriately",
			"Link ticket to similar issues in graph",
			"Send notification email",
		}
		tokens = 800

	case "auto-resolve-known-issue":
		plan.Steps = []string{
			"Read urgent email",
			"Extract error details",
			"Search logs for error patterns",
			"Analyze log context",
			"Find similar issues in knowledge graph",
			"Get solution from knowledge graph",
			"Check current feature flags",
			"Check known issues documentation",
			"Create high-priority ticket",
			"Add comprehensive tags",
			"Link ticket to known issue",
			"Add auto-suggested solution",
			"Update ticket with all findings",
			"Send detailed email with solution",
			"Log resolution for future reference",
		}
		tokens = 1500

	case "security-incident-response":
		plan.Steps = []string{
			// Phase 1: Initial Detection (steps 1-5)
			"Read security alert email",
			"Search security events for suspicious patterns",
			"Analyze suspicious activity in time window",
			"Extract IP addresses from events",
			"Check threat intelligence for IPs",

			// Phase 2: Threat Assessment (steps 6-15)
			"Search logs for failed login attempts",
			"Search logs for unauthorized access attempts",
			"Get blast radius for suspect IPs",
			"Calculate risk score for incident",
			"Identify affected user accounts",
			"Identify affected resources",
			"Search for similar past incidents in graph",
			"Get recommended solutions from graph",
			"Analyze attack patterns",
			"Determine incident severity",

			// Phase 3: Initial Response (steps 16-25)
			"Block malicious IP addresses",
			"Mark compromised user accounts",
			"Create critical security incident ticket",
			"Add security tags to ticket",
			"Link ticket to threat indicators in graph",
			"Update ticket with blast radius data",
			"Update ticket with risk assessment",
			"Create sub-ticket for IP blocking",
			"Create sub-ticket for user account security",
			"Create sub-ticket for forensic analysis",

			// Phase 4: Configuration Updates (steps 26-35)
			"Read current security configuration",
			"Check auto-response feature flags",
			"Check enhanced monitoring feature flags",
			"Update security policy - reduce login threshold",
			"Update security policy - enable MFA requirement",
			"Update security policy - enable enhanced monitoring",
			"Update security policy - reduce escalation time",
			"Write updated security configuration",
			"Log configuration changes",
			"Verify configuration updates applied",

			// Phase 5: Documentation & Notification (steps 36-50)
			"Add incident node to knowledge graph",
			"Add threat indicator nodes to graph",
			"Link incident to affected assets in graph",
			"Link incident to attack patterns",
			"Create incident timeline",
			"Generate incident report",
			"Log IP blocking action",
			"Log account security action",
			"Log monitoring enhancement action",
			"Send notification to security team",
			"Send notification to CISO",
			"Send notification to IT operations",
			"Update ticket status to investigating",
			"Update ticket with final incident report",
			"Mark incident as resolved",
		}
		tokens = 3000

	default:
		return nil, 0, fmt.Errorf("unknown task: %s", task.Name)
	}

	return plan, tokens, nil
}

// planTaskWithLLM uses real LLM to create a task plan
func (a *FunctionCallingAgent) planTaskWithLLM(ctx context.Context, task scenarios.Task, client *llm.ClaudeClient) (*TaskPlan, int, error) {
	systemPrompt := `You are a task planning agent. Create a detailed step-by-step plan.

CRITICAL: Your response must be ONLY a valid JSON array. No explanations, no markdown, no extra text.
Format: ["step1", "step2", "step3"]`

	userPrompt := fmt.Sprintf(`Task: %s
Description: %s
Expected steps: %d

Return a JSON array of %d specific, actionable steps to complete this task.
ONLY return the JSON array, nothing else.`,
		task.Name, task.Description, task.ExpectedOps, task.ExpectedOps)

	// Call Claude API
	response, tokens, err := client.GenerateCode(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, 0, fmt.Errorf("LLM planning failed: %w", err)
	}

	// Clean and parse JSON response
	cleanedJSON := cleanJSONResponse(response)

	var steps []string
	if err := json.Unmarshal([]byte(cleanedJSON), &steps); err != nil {
		return nil, tokens, fmt.Errorf("failed to parse LLM plan (response was: %s): %w", cleanedJSON[:min(200, len(cleanedJSON))], err)
	}

	return &TaskPlan{Steps: steps}, tokens, nil
}

// decideToolCall determines which tool to call for a step by calling the LLM
func (a *FunctionCallingAgent) decideToolCall(ctx context.Context, step string, task scenarios.Task) (*ToolCall, int, error) {
	// Check if we should use real LLM or fallback
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		client := llm.NewClaudeClient()
		return a.decideToolCallWithLLM(ctx, step, task, client)
	}

	// Fallback to hardcoded rules for testing without API key
	tokens := 150 // Simulated tokens per decision

	switch {
	case strings.Contains(step, "Read") && strings.Contains(step, "email"):
		return &ToolCall{
			ToolName: "readEmail",
			Parameters: map[string]interface{}{
				"emailID": "support_001",
			},
		}, tokens, nil

	case strings.Contains(step, "Create") && strings.Contains(step, "ticket") && task.Name != "security-incident-response":
		priority := 3
		tags := []string{"support"}
		description := "Issue from email"

		// Customize based on task
		if task.Complexity == "medium" {
			priority = 4
			tags = []string{"memory", "OutOfMemory"}
			description = "OutOfMemory exception during file upload - ERR-500-XYZ found in logs"
		} else if task.Complexity == "complex" {
			priority = 5
			tags = []string{"memory", "upload", "urgent"}
			description = "Upload feature broken - OutOfMemory exception in UploadHandler (ERR-UPLOAD-500). Found in logs: heap at 95%, failed to allocate 50MB. Similar issue ISS-UPLOAD-MEM in graph."
		}

		return &ToolCall{
			ToolName: "createTicket",
			Parameters: map[string]interface{}{
				"customerID":  "customer123",
				"subject":     "Support ticket",
				"description": description,
				"priority":    priority,
				"tags":        tags,
			},
		}, tokens, nil

	case strings.Contains(step, "Search logs"):
		return &ToolCall{
			ToolName: "searchLogs",
			Parameters: map[string]interface{}{
				"pattern": "ERR-",
			},
		}, tokens, nil

	case strings.Contains(step, "Find similar"):
		return &ToolCall{
			ToolName: "findSimilarIssues",
			Parameters: map[string]interface{}{
				"description": "OutOfMemory",
				"topK":        5,
			},
		}, tokens, nil

	case strings.Contains(step, "Link ticket"):
		ticketID := a.currentTicketID
		if ticketID == "" {
			ticketID = "TICKET-001" // fallback
		}
		return &ToolCall{
			ToolName: "linkIssueInGraph",
			Parameters: map[string]interface{}{
				"ticketID":    ticketID,
				"issueNodeID": "ISS-001",
			},
		}, tokens, nil

	case strings.Contains(step, "Check") && strings.Contains(step, "feature"):
		return &ToolCall{
			ToolName: "checkFeatureFlag",
			Parameters: map[string]interface{}{
				"flagName": "streaming_mode",
			},
		}, tokens, nil

	case strings.Contains(step, "Check") && strings.Contains(step, "known"):
		return &ToolCall{
			ToolName: "readConfig",
			Parameters: map[string]interface{}{
				"filename": "known_issues.yaml",
			},
		}, tokens, nil

	case strings.Contains(step, "Send") && strings.Contains(step, "email"):
		return &ToolCall{
			ToolName: "sendEmail",
			Parameters: map[string]interface{}{
				"to":      "user@company.com",
				"subject": "Ticket confirmation",
				"body":    "Your ticket has been created",
			},
		}, tokens, nil

	case (strings.Contains(step, "Tag ticket") || strings.Contains(step, "Update ticket") || strings.Contains(step, "Add")) && task.Name != "security-incident-response":
		ticketID := a.currentTicketID
		if ticketID == "" {
			ticketID = "TICKET-001" // fallback
		}
		return &ToolCall{
			ToolName: "updateTicket",
			Parameters: map[string]interface{}{
				"ticketID": ticketID,
				"updates": map[string]interface{}{
					"tags": []string{"memory", "urgent"},
				},
			},
		}, tokens, nil

	// Security Incident Response cases
	case task.Name == "security-incident-response" && strings.Contains(step, "Search security events"):
		return &ToolCall{
			ToolName: "searchSecurityEvents",
			Parameters: map[string]interface{}{
				"eventType": "login_failure",
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Analyze suspicious activity"):
		return &ToolCall{
			ToolName: "analyzeSuspiciousActivity",
			Parameters: map[string]interface{}{
				"timeWindowMinutes": 120,
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Check threat intelligence"):
		return &ToolCall{
			ToolName: "checkThreatIntel",
			Parameters: map[string]interface{}{
				"ip": "45.142.212.61",
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Get blast radius"):
		return &ToolCall{
			ToolName: "getBlastRadius",
			Parameters: map[string]interface{}{
				"suspectIPs":         []string{"45.142.212.61", "103.251.167.10"},
				"timeWindowMinutes":  120,
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Calculate risk score"):
		return &ToolCall{
			ToolName: "calculateRiskScore",
			Parameters: map[string]interface{}{
				"factors": map[string]interface{}{
					"failed_logins":     15,
					"affected_users":    1,
					"has_threat_intel":  true,
					"data_exfiltration": true,
					"mfa_bypassed":      true,
				},
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Block malicious IP addresses"):
		// Block both malicious IPs at once
		return &ToolCall{
			ToolName: "blockMultipleIPs",
			Parameters: map[string]interface{}{
				"ips": []string{"45.142.212.61", "103.251.167.10"},
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Mark compromised user"):
		return &ToolCall{
			ToolName: "markUserCompromised",
			Parameters: map[string]interface{}{
				"userID": "admin@company.com",
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Create critical security incident ticket"):
		return &ToolCall{
			ToolName: "createTicket",
			Parameters: map[string]interface{}{
				"customerID":  "security-team",
				"subject":     "CRITICAL: Brute Force Attack Detected from Known Threat Actor",
				"description": "Sophisticated credential stuffing attack detected from malicious IP 45.142.212.61 (known threat actor APT-29). Attack timeline: 15 failed login attempts followed by successful MFA bypass and unauthorized data access across /api/users/export, /api/customers/list, /api/financial/reports. Privilege escalation attempt detected. Blast radius: 1 compromised admin account, 4 critical resources accessed, 1000+ records potentially exfiltrated. Threat intelligence confirms high-confidence match. Immediate containment required.",
				"priority":    5,
				"tags":        []string{"security", "credential-stuffing", "critical"},
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Read current security configuration"):
		return &ToolCall{
			ToolName: "readConfig",
			Parameters: map[string]interface{}{
				"filename": "security_settings.json",
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Write updated security configuration"):
		return &ToolCall{
			ToolName: "writeConfig",
			Parameters: map[string]interface{}{
				"filename": "security_settings.json",
				"data": map[string]interface{}{
					"security": map[string]interface{}{
						"auto_block_threshold":    5,
						"session_timeout_minutes": 30,
						"mfa_required":            true,
						"ip_whitelist_enabled":    false,
						"audit_logging":           true,
					},
				},
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && (strings.Contains(step, "Log") || strings.Contains(step, "action")):
		return &ToolCall{
			ToolName: "writeLog",
			Parameters: map[string]interface{}{
				"filename": "security_incident.log",
				"level":    "INFO",
				"message":  "Security incident response action completed",
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Add") && strings.Contains(step, "node"):
		return &ToolCall{
			ToolName: "addNodeToGraph",
			Parameters: map[string]interface{}{
				"nodeID":      "SEC-INC-001",
				"nodeType":    "incident",
				"description": "Credential stuffing attack",
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Send notification to security team"):
		return &ToolCall{
			ToolName: "sendEmail",
			Parameters: map[string]interface{}{
				"to":      "security-team@company.com",
				"subject": "SECURITY ALERT: Incident Resolved",
				"body":    "Security incident has been successfully contained and mitigated. Malicious IPs blocked, compromised accounts secured, MFA enabled.",
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Send notification to CISO"):
		return &ToolCall{
			ToolName: "sendEmail",
			Parameters: map[string]interface{}{
				"to":      "ciso@company.com",
				"subject": "Security Incident Report",
				"body":    "Executive summary: Credential stuffing attack detected and mitigated. Attack vector: brute force. Response time: <30min. Impact: Minimal - no data breach.",
			},
		}, tokens, nil

	case task.Name == "security-incident-response" && strings.Contains(step, "Send notification to IT operations"):
		return &ToolCall{
			ToolName: "sendEmail",
			Parameters: map[string]interface{}{
				"to":      "it-ops@company.com",
				"subject": "URGENT: Password Reset Required",
				"body":    "Security policies updated. Please implement new firewall rules and force password resets for affected accounts.",
			},
		}, tokens, nil

	default:
		// Generic read operation
		return &ToolCall{
			ToolName: "queryTickets",
			Parameters: map[string]interface{}{
				"filters": map[string]interface{}{},
			},
		}, tokens, nil
	}
}

// decideToolCallWithLLM uses real LLM to decide which tool to call
func (a *FunctionCallingAgent) decideToolCallWithLLM(ctx context.Context, step string, task scenarios.Task, client *llm.ClaudeClient) (*ToolCall, int, error) {
	// Get available tools
	toolsList := a.registry.ListTools()

	systemPrompt := fmt.Sprintf(`You are a function calling agent that decides which tool to call.

Available tools:
%s

CRITICAL: Return ONLY valid JSON. No explanations, no markdown, no extra text.
Format: {"tool": "toolName", "parameters": {"param": "value"}}`, formatToolsList(toolsList))

	userPrompt := fmt.Sprintf(`Task: %s
Description: %s

Current Step: "%s"

IMPORTANT: Use the exact identifiers mentioned in the task description (e.g., if it says "Read email ID 'error_report_001'", use that exact email ID).

Which tool should be called for this step? Return JSON only.`, task.Name, task.Description, step)

	// Call Claude API
	response, tokens, err := client.GenerateCode(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, 0, fmt.Errorf("LLM decision failed: %w", err)
	}

	// Clean and parse JSON response
	cleanedJSON := cleanJSONResponse(response)

	var decision struct {
		Tool       string                 `json:"tool"`
		Parameters map[string]interface{} `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(cleanedJSON), &decision); err != nil {
		return nil, tokens, fmt.Errorf("failed to parse LLM decision (response was: %s): %w", cleanedJSON[:min(200, len(cleanedJSON))], err)
	}

	return &ToolCall{
		ToolName:   decision.Tool,
		Parameters: decision.Parameters,
	}, tokens, nil
}

// formatToolsList formats the tools for the LLM prompt
func formatToolsList(tools []*tools.ToolInfo) string {
	var sb strings.Builder
	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description))
		sb.WriteString("  Parameters: ")
		params := []string{}
		for _, param := range tool.Parameters {
			required := ""
			if param.Required {
				required = " (required)"
			}
			params = append(params, fmt.Sprintf("%s: %s%s", param.Name, param.Type, required))
		}
		sb.WriteString(strings.Join(params, ", "))
		sb.WriteString("\n")
	}
	return sb.String()
}

// executeToolCall executes a tool call through the registry
func (a *FunctionCallingAgent) executeToolCall(toolCall *ToolCall) (interface{}, error) {
	tool, exists := a.registry.GetTool(toolCall.ToolName)
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolCall.ToolName)
	}

	// Convert parameters to JSON for tool execution
	paramsJSON, err := json.Marshal(toolCall.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	var params map[string]interface{}
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	// Execute the tool
	result, err := tool.Function(params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetFunctionCalls returns all logged function calls
func (a *FunctionCallingAgent) GetFunctionCalls() []FunctionCall {
	return a.logger.Calls
}

// PrintFunctionCall prints a formatted function call
func (fc *FunctionCall) Print() {
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("Tool: %s\n", fc.ToolName)
	fmt.Printf("Timestamp: %s\n", fc.Timestamp.Format(time.RFC3339))
	fmt.Printf("Duration: %v\n", fc.Duration)

	if len(fc.Parameters) > 0 {
		fmt.Println("\nParameters:")
		paramsJSON, _ := json.MarshalIndent(fc.Parameters, "  ", "  ")
		fmt.Printf("  %s\n", string(paramsJSON))
	}

	if fc.Error != "" {
		fmt.Printf("\nError: %s\n", fc.Error)
	} else if fc.Result != nil {
		fmt.Println("\nResult:")
		resultJSON, _ := json.MarshalIndent(fc.Result, "  ", "  ")
		fmt.Printf("  %s\n", string(resultJSON))
	}
	fmt.Println(strings.Repeat("─", 80))
}

// cleanJSONResponse extracts JSON from LLM responses that may include explanatory text or markdown
func cleanJSONResponse(response string) string {
	// Trim whitespace
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	if strings.HasPrefix(response, "```") {
		lines := strings.Split(response, "\n")
		if len(lines) > 2 {
			// Skip first line (```json or ```) and last line (```)
			response = strings.Join(lines[1:len(lines)-1], "\n")
			response = strings.TrimSpace(response)
		}
	}

	// Try to find JSON object or array in the response
	// Look for { or [ as the start of JSON
	startIdx := -1
	endIdx := -1

	for i, ch := range response {
		if ch == '{' || ch == '[' {
			startIdx = i
			break
		}
	}

	if startIdx >= 0 {
		// Find matching closing bracket
		openChar := rune(response[startIdx])
		closeChar := '}'
		if openChar == '[' {
			closeChar = ']'
		}

		depth := 0
		for i := startIdx; i < len(response); i++ {
			ch := rune(response[i])
			if ch == openChar || (openChar == '{' && ch == '[') || (openChar == '[' && ch == '{') {
				depth++
			} else if ch == closeChar || (closeChar == '}' && ch == ']') || (closeChar == ']' && ch == '}') {
				depth--
				if depth == 0 {
					endIdx = i
					break
				}
			}
		}

		if endIdx > startIdx {
			return response[startIdx : endIdx+1]
		}
	}

	return response
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
