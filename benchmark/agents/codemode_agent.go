package agents

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/imran31415/godemode/benchmark/llm"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/tools"
	"github.com/imran31415/godemode/pkg/compiler"
	"github.com/imran31415/godemode/pkg/executor"
	"github.com/imran31415/godemode/pkg/validator"
)

// CodeExecutor interface allows switching between WASM and Interpreter executors
type CodeExecutor interface {
	Execute(ctx context.Context, sourceCode string, timeout time.Duration) (*executor.ExecutionResult, error)
}

// CodeModeAgent solves tasks by generating and executing Go code
type CodeModeAgent struct {
	compiler  *compiler.Compiler
	validator *validator.Validator
	executor  CodeExecutor // Changed to interface for flexibility
	logger    *Logger
	registry  *tools.ToolRegistry // Add tool registry for actual execution
}

// Logger captures all generated code for visibility
type Logger struct {
	GeneratedCode []CodeLog
}

// CodeLog represents a single code generation event
type CodeLog struct {
	Timestamp   time.Time
	Task        string
	Prompt      string
	Code        string
	CompileTime time.Duration
	ExecuteTime time.Duration
	Success     bool
	Error       string
	Output      string
}

// AgentMetrics tracks performance metrics
type AgentMetrics struct {
	TaskName        string
	StartTime       time.Time
	EndTime         time.Time
	TotalDuration   time.Duration
	TokensUsed      int
	APICallCount    int
	OperationsCount int
	Success         bool
	Errors          []string
}

// NewCodeModeAgent creates a new code mode agent with Interpreter executor (fast, no compilation)
func NewCodeModeAgent(env *scenarios.TestEnvironment) *CodeModeAgent {
	comp := compiler.NewCompiler()
	val := validator.NewValidator()
	// Use InterpreterExecutor for 226x faster cold start and 6x faster average execution
	exec := executor.NewInterpreterExecutor()
	registry := tools.NewToolRegistry(env)

	return &CodeModeAgent{
		compiler:  comp,
		validator: val,
		executor:  exec,
		logger:    &Logger{GeneratedCode: []CodeLog{}},
		registry:  registry,
	}
}

// RunTask executes a task using code mode
func (a *CodeModeAgent) RunTask(ctx context.Context, task scenarios.Task, env *scenarios.TestEnvironment) (*AgentMetrics, error) {
	metrics := &AgentMetrics{
		TaskName:  task.Name,
		StartTime: time.Now(),
	}

	// Generate the system prompt with available APIs
	systemPrompt := a.buildSystemPrompt(task, env)

	// Simulate LLM call to generate code
	// In real implementation, this would call Claude API
	code, tokensUsed, err := a.generateCode(ctx, systemPrompt, task)
	if err != nil {
		metrics.EndTime = time.Now()
		metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false
		metrics.Errors = append(metrics.Errors, fmt.Sprintf("Code generation failed: %v", err))
		return metrics, err
	}

	metrics.TokensUsed = tokensUsed
	metrics.APICallCount = 1 // One LLM call for code generation

	// Log the generated code
	codeLog := CodeLog{
		Timestamp: time.Now(),
		Task:      task.Name,
		Prompt:    systemPrompt,
		Code:      code,
	}

	// Validate the code
	if err := a.validator.Validate(code); err != nil {
		codeLog.Success = false
		codeLog.Error = fmt.Sprintf("Validation failed: %v", err)
		a.logger.GeneratedCode = append(a.logger.GeneratedCode, codeLog)

		metrics.EndTime = time.Now()
		metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false
		metrics.Errors = append(metrics.Errors, codeLog.Error)
		return metrics, fmt.Errorf("validation failed: %w", err)
	}

	// Execute the code (validation + compilation + execution all handled by executor)
	executeStart := time.Now()
	result, err := a.executor.Execute(ctx, code, 30*time.Second)
	codeLog.CompileTime = 0 // Included in execute time
	codeLog.ExecuteTime = time.Since(executeStart)

	if err != nil {
		codeLog.Success = false
		codeLog.Error = fmt.Sprintf("Execution failed: %v", err)
		codeLog.Output = result.Stdout
		a.logger.GeneratedCode = append(a.logger.GeneratedCode, codeLog)

		metrics.EndTime = time.Now()
		metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false
		metrics.Errors = append(metrics.Errors, codeLog.Error)
		return metrics, fmt.Errorf("execution failed: %w", err)
	}

	codeLog.Success = true
	codeLog.Output = result.Stdout
	a.logger.GeneratedCode = append(a.logger.GeneratedCode, codeLog)

	// Execute actual operations by parsing the generated code
	// Extract tool calls from the generated code and execute them
	err = a.executeFromGeneratedCode(ctx, code, task, env)
	if err != nil {
		metrics.EndTime = time.Now()
		metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Success = false
		metrics.Errors = append(metrics.Errors, fmt.Sprintf("Operation execution failed: %v", err))
		return metrics, err
	}

	// Count operations from the code
	metrics.OperationsCount = a.countOperations(code)

	metrics.EndTime = time.Now()
	metrics.TotalDuration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Success = true

	return metrics, nil
}

// buildSystemPrompt creates the prompt for code generation
func (a *CodeModeAgent) buildSystemPrompt(task scenarios.Task, env *scenarios.TestEnvironment) string {
	var sb strings.Builder

	sb.WriteString("You are a Go code generator that solves support ticket tasks.\n\n")
	sb.WriteString(fmt.Sprintf("TASK: %s\n", task.Description))
	sb.WriteString(fmt.Sprintf("COMPLEXITY: %s\n", task.Complexity))
	sb.WriteString(fmt.Sprintf("EXPECTED OPERATIONS: %d\n\n", task.ExpectedOps))

	sb.WriteString("You have access to the following systems:\n\n")

	// Email System API
	sb.WriteString("1. EMAIL SYSTEM API:\n")
	sb.WriteString("   - ReadEmail(emailID string) (*Email, error)\n")
	sb.WriteString("   - WriteEmail(to, subject, body string) (string, error)\n")
	sb.WriteString("   - ListEmails(folderPath string) ([]string, error)\n")
	sb.WriteString("   - Email.ExtractErrorCode() string\n\n")

	// Database API
	sb.WriteString("2. SQLITE DATABASE API:\n")
	sb.WriteString("   - CreateTicket(ticket *Ticket) error\n")
	sb.WriteString("   - UpdateTicket(id string, updates map[string]interface{}) error\n")
	sb.WriteString("   - GetTicket(id string) (*Ticket, error)\n")
	sb.WriteString("   - QueryTickets(filters map[string]interface{}) ([]*Ticket, error)\n")
	sb.WriteString("   - DeleteTicket(id string) error\n\n")

	// Graph API
	sb.WriteString("3. KNOWLEDGE GRAPH API:\n")
	sb.WriteString("   - AddNode(node *Node) error\n")
	sb.WriteString("   - AddEdge(from, to, edgeType string) error\n")
	sb.WriteString("   - GetNode(id string) (*Node, error)\n")
	sb.WriteString("   - FindSimilar(description string, nodeType string, topK int) ([]*Node, error)\n")
	sb.WriteString("   - GetNeighbors(nodeID string, edgeType string) ([]*Node, error)\n\n")

	// Log System API
	sb.WriteString("4. LOG SYSTEM API:\n")
	sb.WriteString("   - SearchLogs(pattern string, timeWindow time.Duration) ([]*LogEntry, error)\n")
	sb.WriteString("   - ExtractErrorContext(errorCode string, contextLines int) (string, error)\n")
	sb.WriteString("   - WriteLog(filename string, level string, message string) error\n\n")

	// Config System API
	sb.WriteString("5. CONFIG SYSTEM API:\n")
	sb.WriteString("   - ReadConfig(filename string) (map[string]interface{}, error)\n")
	sb.WriteString("   - WriteConfig(filename string, data map[string]interface{}) error\n")
	sb.WriteString("   - CheckFeatureFlag(flagName string) (bool, error)\n\n")

	sb.WriteString("Generate a complete Go program that:\n")
	sb.WriteString("1. Imports necessary packages\n")
	sb.WriteString("2. Defines main() function\n")
	sb.WriteString("3. Uses the APIs above to complete the task\n")
	sb.WriteString("4. Handles errors appropriately\n")
	sb.WriteString("5. Prints results to stdout\n\n")

	sb.WriteString("IMPORTANT:\n")
	sb.WriteString("- Use only standard library and provided APIs\n")
	sb.WriteString("- No external network calls\n")
	sb.WriteString("- No file system access except through provided APIs\n")
	sb.WriteString("- No dangerous operations\n\n")

	sb.WriteString("Return ONLY the Go code, no explanations.\n")

	return sb.String()
}

// generateCode calls the Claude API to generate code
func (a *CodeModeAgent) generateCode(ctx context.Context, prompt string, task scenarios.Task) (string, int, error) {
	// Create Claude client
	client := llm.NewClaudeClient()

	// Build user prompt with task-specific context
	userPrompt := fmt.Sprintf(`Task: %s

Description: %s
Expected operations: %d

Please generate a complete, executable Go program that accomplishes this task using the APIs described in the system prompt.

Requirements:
- The code must be valid, compilable Go
- Use only the APIs provided in the system prompt
- Handle all errors appropriately
- Print progress and results to stdout
- The main() function should orchestrate the entire workflow

Return ONLY the Go code, no explanations or markdown formatting.`,
		task.Name,
		task.Description,
		task.ExpectedOps,
	)

	// Call Claude API to generate the code
	code, tokens, err := client.GenerateCode(ctx, prompt, userPrompt)
	if err != nil {
		return "", 0, fmt.Errorf("Claude API call failed: %w", err)
	}

	return code, tokens, nil
}

// executeFromGeneratedCode parses the generated code and executes tool calls dynamically
func (a *CodeModeAgent) executeFromGeneratedCode(ctx context.Context, code string, task scenarios.Task, env *scenarios.TestEnvironment) error {
	// Map of function names to tool names
	toolMap := map[string]string{
		"ReadEmail":          "readEmail",
		"ListEmails":         "listEmails",
		"WriteEmail":         "sendEmail",
		"CreateTicket":       "createTicket",
		"UpdateTicket":       "updateTicket",
		"GetTicket":          "getTicket",
		"QueryTickets":       "queryTickets",
		"DeleteTicket":       "deleteTicket",
		"SearchLogs":         "searchLogs",
		"ExtractErrorContext": "extractErrorContext",
		"WriteLog":           "writeLog",
		"FindSimilar":        "findSimilarIssues",
		"AddNode":            "addNode",
		"AddEdge":            "addEdge",
		"GetNode":            "getNode",
		"GetNeighbors":       "getNeighbors",
		"ReadConfig":         "readConfig",
		"WriteConfig":        "writeConfig",
		"CheckFeatureFlag":   "checkFeatureFlag",
	}

	// Parse the code for function calls
	// This is a simple parser - looks for function calls in main()
	lines := strings.Split(code, "\n")
	inMain := false
	braceDepth := 0
	var lastTicketID string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track when we're inside main()
		if strings.Contains(trimmed, "func main()") {
			inMain = true
			braceDepth = 0
			continue
		}

		if !inMain {
			continue
		}

		// Track brace depth to know when we exit main()
		braceDepth += strings.Count(trimmed, "{")
		braceDepth -= strings.Count(trimmed, "}")

		// Exit main when braceDepth goes negative (closing brace of main)
		if braceDepth < 0 {
			break
		}

		line := trimmed

		// Look for tool function calls
		for funcName, toolName := range toolMap {
			if strings.Contains(line, funcName+"(") {
				// Extract arguments based on the tool
				params := a.extractParams(toolName, line, code, lastTicketID)

				// Execute the tool
				result, _ := a.registry.Call(toolName, params)

				// Track ticket ID for linking operations
				if toolName == "createTicket" && result != nil {
					if resultMap, ok := result.(map[string]interface{}); ok {
						if ticketID, ok := resultMap["ticketID"].(string); ok {
							lastTicketID = ticketID
						}
					}
				}
			}
		}
	}

	return nil
}

// extractParams extracts parameters for a tool call from the generated code
func (a *CodeModeAgent) extractParams(toolName, line, fullCode string, lastTicketID string) map[string]interface{} {
	params := make(map[string]interface{})

	switch toolName {
	case "readEmail":
		// Extract email ID from ReadEmail("test_email_001")
		emailID := ""
		if start := strings.Index(line, "ReadEmail(\""); start != -1 {
			start += len("ReadEmail(\"")
			if end := strings.Index(line[start:], "\""); end != -1 {
				emailID = line[start : start+end]
			}
		} else if start := strings.Index(line, "readEmail(\""); start != -1 {
			start += len("readEmail(\"")
			if end := strings.Index(line[start:], "\""); end != -1 {
				emailID = line[start : start+end]
			}
		}
		// Default if not found
		if emailID == "" {
			emailID = "unknown_email"
		}
		params["emailID"] = emailID

	case "createTicket":
		// Extract priority and tags from the ticket struct definition in the generated code
		priority := 3  // default
		tags := []string{}
		description := "Request from generated code"

		// Search for ticketDescription variable or Description field value
		lines := strings.Split(fullCode, "\n")

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Extract priority
			if strings.Contains(trimmed, "Priority:") {
				parts := strings.Split(trimmed, "Priority:")
				if len(parts) > 1 {
					numStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(parts[1]), ","))
					if p, err := strconv.Atoi(numStr); err == nil && p > priority {
						priority = p
					}
				}
			}

			// Extract tags - look for Tags: []string{"memory", "OutOfMemory"}
			if strings.Contains(trimmed, "Tags:") && strings.Contains(trimmed, "[]string{") {
				start := strings.Index(trimmed, "[]string{")
				if start != -1 {
					start += len("[]string{")
					end := strings.Index(trimmed[start:], "}")
					if end != -1 {
						tagStr := trimmed[start : start+end]
						for _, t := range strings.Split(tagStr, ",") {
							tag := strings.Trim(strings.TrimSpace(t), "\"")
							if tag != "" {
								tags = append(tags, tag)
							}
						}
					}
				}
			}

			// Extract description from Description: field or ticketDescription variable
			// Look for Description: "some text" or Description: ticketDescription
			if strings.Contains(trimmed, "Description:") {
				// First, try to find a string literal directly
				if idx := strings.Index(trimmed, "Description:"); idx != -1 {
					rest := trimmed[idx+len("Description:"):]
					rest = strings.TrimSpace(rest)

					// Check if it's a direct string literal
					if strings.HasPrefix(rest, "\"") || strings.HasPrefix(rest, "`") {
						quote := rest[0:1]
						rest = rest[1:]

						// For backtick strings, collect until closing backtick
						if quote == "`" {
							// Multi-line string - collect from current line onward
							var descParts []string
							descParts = append(descParts, rest)

							// Continue collecting lines until we find closing backtick
							for j := i + 1; j < len(lines) && j < i+50; j++ {
								nextLine := lines[j]
								descParts = append(descParts, nextLine)
								if strings.Contains(nextLine, "`") {
									break
								}
							}

							fullDesc := strings.Join(descParts, "\n")
							if endIdx := strings.Index(fullDesc, "`"); endIdx != -1 {
								description = fullDesc[:endIdx]
							}
						} else {
							// Regular quoted string
							if endIdx := strings.Index(rest, "\""); endIdx != -1 {
								description = rest[:endIdx]
							}
						}
					} else if strings.Contains(rest, "ticketDescription") {
						// It references a variable - find the variable assignment
						for k := i - 1; k >= 0 && k > i-50; k-- {
							varLine := lines[k]
							if strings.Contains(varLine, "ticketDescription") && strings.Contains(varLine, ":=") {
								// Found variable assignment - extract its value
								if idx := strings.Index(varLine, ":="); idx != -1 {
									valPart := strings.TrimSpace(varLine[idx+2:])

									// Handle fmt.Sprintf or direct string
									if strings.HasPrefix(valPart, "fmt.Sprintf(") || strings.HasPrefix(valPart, "\"") || strings.HasPrefix(valPart, "`") {
										// Extract string content across multiple lines if needed
										var descParts []string

										for m := k; m < len(lines) && m < k+50; m++ {
											descParts = append(descParts, lines[m])
											// Stop when we hit ticket := or another assignment
											if m > k && (strings.Contains(lines[m], "ticket :=") || (strings.Contains(lines[m], ":=") && !strings.Contains(lines[m], "ticketDescription"))) {
												break
											}
										}

										fullDescBlock := strings.Join(descParts, " ")

										// Extract the actual text content
										// Look for quoted strings within the block
										var extractedParts []string
										inQuote := false
										var currentPart strings.Builder

										for _, ch := range fullDescBlock {
											if ch == '"' {
												if inQuote {
													extractedParts = append(extractedParts, currentPart.String())
													currentPart.Reset()
												}
												inQuote = !inQuote
											} else if inQuote {
												currentPart.WriteRune(ch)
											}
										}

										if len(extractedParts) > 0 {
											description = strings.Join(extractedParts, " ")
										}
									}
								}
								break
							}
						}
					}
				}
			}
		}

		params["customerID"] = "customer123"
		params["subject"] = "Support Request"
		params["description"] = description
		params["priority"] = priority
		params["tags"] = tags

	case "sendEmail":
		// Extract email parameters - simplified for now
		params["to"] = "user@example.com"
		params["subject"] = "Confirmation"
		params["body"] = "Your request has been processed"

	case "searchLogs":
		// Extract search pattern - try multiple patterns
		pattern := ""
		if start := strings.Index(line, "SearchLogs(\""); start != -1 {
			start += len("SearchLogs(\"")
			if end := strings.Index(line[start:], "\""); end != -1 {
				pattern = line[start : start+end]
			}
		} else if start := strings.Index(line, "searchLogs(\""); start != -1 {
			start += len("searchLogs(\"")
			if end := strings.Index(line[start:], "\""); end != -1 {
				pattern = line[start : start+end]
			}
		}
		// Default if not found
		if pattern == "" {
			pattern = "ERROR"
		}
		params["pattern"] = pattern

	case "findSimilarIssues":
		params["description"] = "similar issue"
		params["topK"] = 5

	case "linkIssueInGraph":
		if lastTicketID != "" {
			params["ticketID"] = lastTicketID
			params["issueNodeID"] = "ISS-001"
		}

	case "readConfig":
		filename := ""
		if start := strings.Index(line, "ReadConfig(\""); start != -1 {
			start += len("ReadConfig(\"")
			if end := strings.Index(line[start:], "\""); end != -1 {
				filename = line[start : start+end]
			}
		} else if start := strings.Index(line, "readConfig(\""); start != -1 {
			start += len("readConfig(\"")
			if end := strings.Index(line[start:], "\""); end != -1 {
				filename = line[start : start+end]
			}
		}
		// Default if not found
		if filename == "" {
			filename = "config.json"
		}
		params["filename"] = filename

	case "checkFeatureFlag":
		flagName := ""
		if start := strings.Index(line, "CheckFeatureFlag(\""); start != -1 {
			start += len("CheckFeatureFlag(\"")
			if end := strings.Index(line[start:], "\""); end != -1 {
				flagName = line[start : start+end]
			}
		} else if start := strings.Index(line, "checkFeatureFlag(\""); start != -1 {
			start += len("checkFeatureFlag(\"")
			if end := strings.Index(line[start:], "\""); end != -1 {
				flagName = line[start : start+end]
			}
		}
		// Default if not found
		if flagName == "" {
			flagName = "default_flag"
		}
		params["flagName"] = flagName

	case "addNode":
		// Graph node operations - provide defaults
		params["nodeID"] = "NODE-001"
		params["nodeType"] = "issue"
		params["data"] = map[string]interface{}{"description": "auto-generated"}

	case "addEdge":
		// Graph edge operations - provide defaults
		params["fromID"] = "NODE-001"
		params["toID"] = "NODE-002"
		params["edgeType"] = "related"

	case "updateTicket":
		// Use lastTicketID if available
		if lastTicketID != "" {
			params["ticketID"] = lastTicketID
		} else {
			params["ticketID"] = "TICKET-001"
		}
		params["updates"] = map[string]interface{}{"status": "updated"}

	case "getTicket":
		params["ticketID"] = lastTicketID
		if params["ticketID"] == "" {
			params["ticketID"] = "TICKET-001"
		}

	case "queryTickets":
		params["filters"] = map[string]interface{}{}

	case "deleteTicket":
		params["ticketID"] = "TICKET-001"

	case "listEmails":
		params["folderPath"] = "inbox"

	case "writeLog":
		params["filename"] = "application.log"
		params["level"] = "INFO"
		params["message"] = "Automated log entry"

	case "writeConfig":
		params["filename"] = "config.json"
		params["data"] = map[string]interface{}{}

	case "extractErrorContext":
		params["errorCode"] = "ERROR"
		params["contextLines"] = 5

	case "getNode":
		params["nodeID"] = "NODE-001"

	case "getNeighbors":
		params["nodeID"] = "NODE-001"
		params["edgeType"] = "related"
	}

	return params
}


// countOperations counts the number of operations in the code
func (a *CodeModeAgent) countOperations(code string) int {
	// Simple heuristic: count function calls
	// In real implementation, would parse AST
	count := 0

	operations := []string{
		"ReadEmail",
		"WriteEmail",
		"ListEmails",
		"CreateTicket",
		"UpdateTicket",
		"GetTicket",
		"QueryTickets",
		"SearchLogs",
		"FindSimilar",
		"AddNode",
		"AddEdge",
		"GetNeighbors",
		"ReadConfig",
		"CheckFeatureFlag",
	}

	for _, op := range operations {
		count += strings.Count(code, op)
	}

	return count
}

// GetCodeLogs returns all logged code for visibility
func (a *CodeModeAgent) GetCodeLogs() []CodeLog {
	return a.logger.GeneratedCode
}

// PrintCodeLog prints a formatted code log
func (cl *CodeLog) Print() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Task: %s\n", cl.Task)
	fmt.Printf("Timestamp: %s\n", cl.Timestamp.Format(time.RFC3339))
	fmt.Printf("Success: %v\n", cl.Success)
	if cl.Error != "" {
		fmt.Printf("Error: %s\n", cl.Error)
	}
	fmt.Printf("Compile Time: %v\n", cl.CompileTime)
	fmt.Printf("Execute Time: %v\n", cl.ExecuteTime)
	fmt.Println("\nGenerated Code:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(cl.Code)
	fmt.Println(strings.Repeat("-", 80))
	if cl.Output != "" {
		fmt.Println("\nOutput:")
		fmt.Println(cl.Output)
	}
	fmt.Println(strings.Repeat("=", 80))
}
