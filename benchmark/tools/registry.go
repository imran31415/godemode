package tools

import (
	"fmt"
	"sync"
	"time"

	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/imran31415/godemode/benchmark/systems/security"
)

// ToolFunc is a function signature for tools
type ToolFunc func(args map[string]interface{}) (interface{}, error)

// ToolInfo contains metadata about a tool
type ToolInfo struct {
	Name        string
	Description string
	Parameters  []ParamInfo
	Function    ToolFunc
}

// ParamInfo describes a parameter
type ParamInfo struct {
	Name     string
	Type     string
	Required bool
}

// Registry manages all available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*ToolInfo
}

// ToolRegistry manages tools for a specific test environment
type ToolRegistry struct {
	*Registry
	env *scenarios.TestEnvironment
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*ToolInfo),
	}
}

// NewToolRegistry creates a tool registry with all tools registered for an environment
func NewToolRegistry(env *scenarios.TestEnvironment) *ToolRegistry {
	tr := &ToolRegistry{
		Registry: NewRegistry(),
		env:      env,
	}

	// Register all tools
	// Import the tool packages and register them
	// For now, register placeholder tools
	tr.registerPlaceholderTools()

	return tr
}

// registerPlaceholderTools registers tools that actually call the real systems
func (tr *ToolRegistry) registerPlaceholderTools() {
	// Email tools
	tr.Register(&ToolInfo{
		Name:        "readEmail",
		Description: "Read an email by ID",
		Parameters: []ParamInfo{
			{Name: "emailID", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			emailID := args["emailID"].(string)
			email, err := tr.env.EmailSystem.ReadEmail(emailID)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"from":    email.From,
				"to":      email.To,
				"subject": email.Subject,
				"body":    email.Body,
			}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "sendEmail",
		Description: "Send an email",
		Parameters: []ParamInfo{
			{Name: "to", Type: "string", Required: true},
			{Name: "subject", Type: "string", Required: true},
			{Name: "body", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			to := args["to"].(string)
			subject := args["subject"].(string)
			body := args["body"].(string)
			emailID, err := tr.env.EmailSystem.WriteEmail(to, subject, body)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"emailID": emailID}, nil
		},
	})

	// Ticket tools
	tr.Register(&ToolInfo{
		Name:        "createTicket",
		Description: "Create a new support ticket",
		Parameters: []ParamInfo{
			{Name: "customerID", Type: "string", Required: true},
			{Name: "subject", Type: "string", Required: true},
			{Name: "description", Type: "string", Required: true},
			{Name: "priority", Type: "int", Required: false},
			{Name: "tags", Type: "[]string", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			priority := 3 // default
			if p, ok := args["priority"]; ok {
				if pInt, ok := p.(int); ok {
					priority = pInt
				} else if pFloat, ok := p.(float64); ok {
					priority = int(pFloat)
				}
			}

			tags := []string{}
			if t, ok := args["tags"]; ok {
				if tagSlice, ok := t.([]interface{}); ok {
					for _, tag := range tagSlice {
						if tagStr, ok := tag.(string); ok {
							tags = append(tags, tagStr)
						}
					}
				} else if tagSlice, ok := t.([]string); ok {
					tags = tagSlice
				}
			}

			ticket := &database.Ticket{
				CustomerID:  args["customerID"].(string),
				Subject:     args["subject"].(string),
				Description: args["description"].(string),
				Priority:    priority,
				Status:      "open",
				Tags:        tags,
			}

			err := tr.env.Database.CreateTicket(ticket)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"ticketID": ticket.ID,
				"priority": priority,
			}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "updateTicket",
		Description: "Update ticket fields",
		Parameters: []ParamInfo{
			{Name: "ticketID", Type: "string", Required: true},
			{Name: "updates", Type: "map", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			ticketID := args["ticketID"].(string)
			updates := args["updates"].(map[string]interface{})

			err := tr.env.Database.UpdateTicket(ticketID, updates)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "updated"}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "queryTickets",
		Description: "Query tickets with filters",
		Parameters: []ParamInfo{
			{Name: "filters", Type: "map", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			filters := make(map[string]interface{})
			if f, ok := args["filters"]; ok {
				if fMap, ok := f.(map[string]interface{}); ok {
					filters = fMap
				}
			}

			tickets, err := tr.env.Database.QueryTickets(filters)
			if err != nil {
				return nil, err
			}

			return tickets, nil
		},
	})

	// Graph tools
	tr.Register(&ToolInfo{
		Name:        "findSimilarIssues",
		Description: "Find similar issues in knowledge graph",
		Parameters: []ParamInfo{
			{Name: "description", Type: "string", Required: true},
			{Name: "topK", Type: "int", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			description := args["description"].(string)
			topK := 5
			if k, ok := args["topK"]; ok {
				if kInt, ok := k.(int); ok {
					topK = kInt
				} else if kFloat, ok := k.(float64); ok {
					topK = int(kFloat)
				}
			}

			nodes, err := tr.env.Graph.FindSimilar(description, "issue", topK)
			if err != nil {
				return nil, err
			}

			return nodes, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "linkIssueInGraph",
		Description: "Link ticket to issue in graph",
		Parameters: []ParamInfo{
			{Name: "ticketID", Type: "string", Required: true},
			{Name: "issueNodeID", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			ticketID := args["ticketID"].(string)
			issueNodeID := args["issueNodeID"].(string)

			// Add ticket as node if not exists
			ticketNode := &graph.Node{
				ID:   ticketID,
				Type: "ticket",
				Data: map[string]interface{}{
					"id": ticketID,
				},
			}
			tr.env.Graph.AddNode(ticketNode)

			// Create edge
			err := tr.env.Graph.AddEdge(ticketID, issueNodeID, "similar_to")
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{"status": "linked"}, nil
		},
	})

	// Log tools
	tr.Register(&ToolInfo{
		Name:        "searchLogs",
		Description: "Search application logs",
		Parameters: []ParamInfo{
			{Name: "pattern", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			pattern := args["pattern"].(string)
			entries, err := tr.env.LogSystem.SearchLogs(pattern, 0)
			if err != nil {
				return nil, err
			}
			return entries, nil
		},
	})

	// Config tools
	tr.Register(&ToolInfo{
		Name:        "readConfig",
		Description: "Read configuration file",
		Parameters: []ParamInfo{
			{Name: "filename", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			filename := args["filename"].(string)
			config, err := tr.env.ConfigSystem.ReadConfig(filename)
			if err != nil {
				return nil, err
			}
			return config, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "checkFeatureFlag",
		Description: "Check if a feature flag is enabled",
		Parameters: []ParamInfo{
			{Name: "flagName", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			flagName := args["flagName"].(string)
			enabled, err := tr.env.ConfigSystem.CheckFeatureFlag(flagName)
			if err != nil {
				return nil, err
			}
			return enabled, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "writeConfig",
		Description: "Write configuration to a file",
		Parameters: []ParamInfo{
			{Name: "filename", Type: "string", Required: true},
			{Name: "data", Type: "map", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			filename := args["filename"].(string)
			data, ok := args["data"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("data must be a map")
			}
			err := tr.env.ConfigSystem.WriteConfig(filename, data)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"success": true, "filename": filename}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "writeLog",
		Description: "Write a log entry",
		Parameters: []ParamInfo{
			{Name: "filename", Type: "string", Required: true},
			{Name: "level", Type: "string", Required: true},
			{Name: "message", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			filename := args["filename"].(string)
			level := args["level"].(string)
			message := args["message"].(string)
			err := tr.env.LogSystem.WriteLog(filename, level, message)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"success": true}, nil
		},
	})

	// Security monitoring tools
	tr.Register(&ToolInfo{
		Name:        "logSecurityEvent",
		Description: "Log a security event to the security monitor",
		Parameters: []ParamInfo{
			{Name: "eventType", Type: "string", Required: true},
			{Name: "severity", Type: "string", Required: true},
			{Name: "sourceIP", Type: "string", Required: true},
			{Name: "userID", Type: "string", Required: true},
			{Name: "resource", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			err := tr.env.SecurityMonitor.LogSecurityEvent(&security.SecurityEvent{
				EventType: args["eventType"].(string),
				Severity:  args["severity"].(string),
				SourceIP:  args["sourceIP"].(string),
				UserID:    args["userID"].(string),
				Resource:  args["resource"].(string),
			})
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"success": true}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "searchSecurityEvents",
		Description: "Search for security events matching criteria",
		Parameters: []ParamInfo{
			{Name: "eventType", Type: "string", Required: false},
			{Name: "sourceIP", Type: "string", Required: false},
			{Name: "userID", Type: "string", Required: false},
			{Name: "severity", Type: "string", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			events, err := tr.env.SecurityMonitor.SearchEvents(args)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"count":  len(events),
				"events": events,
			}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "analyzeSuspiciousActivity",
		Description: "Analyze security events for suspicious patterns",
		Parameters: []ParamInfo{
			{Name: "timeWindowMinutes", Type: "int", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			minutes := 60 // default
			if m, ok := args["timeWindowMinutes"]; ok {
				if mInt, ok := m.(int); ok {
					minutes = mInt
				} else if mFloat, ok := m.(float64); ok {
					minutes = int(mFloat)
				}
			}

			suspicious, err := tr.env.SecurityMonitor.AnalyzeSuspiciousActivity(time.Duration(minutes) * time.Minute)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"suspicious_patterns": suspicious,
				"count":              len(suspicious),
			}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "checkThreatIntel",
		Description: "Check if an IP is in threat intelligence database",
		Parameters: []ParamInfo{
			{Name: "ip", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			ip := args["ip"].(string)
			threat, exists := tr.env.SecurityMonitor.CheckThreatIntel(ip)
			if !exists {
				return map[string]interface{}{"found": false}, nil
			}
			return map[string]interface{}{
				"found":       true,
				"threatLevel": threat.ThreatLevel,
				"threatActor": threat.ThreatActor,
				"attackType":  threat.AttackType,
				"confidence":  threat.Confidence,
			}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "blockIP",
		Description: "Block an IP address",
		Parameters: []ParamInfo{
			{Name: "ip", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			ip := args["ip"].(string)
			err := tr.env.SecurityMonitor.BlockIP(ip)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"blocked": true, "ip": ip}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "blockMultipleIPs",
		Description: "Block multiple IP addresses",
		Parameters: []ParamInfo{
			{Name: "ips", Type: "[]string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			// Convert IPs
			ips := []string{}
			if ipList, ok := args["ips"].([]interface{}); ok {
				for _, ip := range ipList {
					if ipStr, ok := ip.(string); ok {
						ips = append(ips, ipStr)
					}
				}
			} else if ipList, ok := args["ips"].([]string); ok {
				ips = ipList
			}

			blocked := []string{}
			for _, ip := range ips {
				err := tr.env.SecurityMonitor.BlockIP(ip)
				if err != nil {
					return nil, fmt.Errorf("failed to block IP %s: %w", ip, err)
				}
				blocked = append(blocked, ip)
			}

			return map[string]interface{}{"blocked": true, "ips": blocked, "count": len(blocked)}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "markUserCompromised",
		Description: "Mark a user account as compromised",
		Parameters: []ParamInfo{
			{Name: "userID", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			userID := args["userID"].(string)
			err := tr.env.SecurityMonitor.MarkUserCompromised(userID)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"compromised": true, "userID": userID}, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "getBlastRadius",
		Description: "Calculate the blast radius of a security incident",
		Parameters: []ParamInfo{
			{Name: "suspectIPs", Type: "[]string", Required: true},
			{Name: "timeWindowMinutes", Type: "int", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			// Convert suspectIPs
			suspectIPs := []string{}
			if ips, ok := args["suspectIPs"].([]interface{}); ok {
				for _, ip := range ips {
					if ipStr, ok := ip.(string); ok {
						suspectIPs = append(suspectIPs, ipStr)
					}
				}
			} else if ips, ok := args["suspectIPs"].([]string); ok {
				suspectIPs = ips
			}

			minutes := 60
			if m, ok := args["timeWindowMinutes"]; ok {
				if mInt, ok := m.(int); ok {
					minutes = mInt
				} else if mFloat, ok := m.(float64); ok {
					minutes = int(mFloat)
				}
			}

			blastRadius, err := tr.env.SecurityMonitor.GetBlastRadius(suspectIPs, time.Duration(minutes)*time.Minute)
			if err != nil {
				return nil, err
			}
			return blastRadius, nil
		},
	})

	tr.Register(&ToolInfo{
		Name:        "calculateRiskScore",
		Description: "Calculate risk score for an incident",
		Parameters: []ParamInfo{
			{Name: "factors", Type: "map", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			if tr.env.SecurityMonitor == nil {
				return nil, fmt.Errorf("security monitor not initialized")
			}

			factors, ok := args["factors"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("factors must be a map")
			}

			score := tr.env.SecurityMonitor.CalculateRiskScore(factors)
			return map[string]interface{}{
				"riskScore": score,
				"severity":  getRiskSeverity(score),
			}, nil
		},
	})
}

func getRiskSeverity(score int) string {
	if score >= 75 {
		return "critical"
	} else if score >= 50 {
		return "high"
	} else if score >= 25 {
		return "medium"
	}
	return "low"
}

// Register adds a tool to the registry
func (r *Registry) Register(info *ToolInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[info.Name]; exists {
		return fmt.Errorf("tool already registered: %s", info.Name)
	}

	r.tools[info.Name] = info
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*ToolInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, found := r.tools[name]
	return tool, found
}

// Call invokes a tool by name with arguments
func (r *Registry) Call(name string, args map[string]interface{}) (interface{}, error) {
	tool, found := r.Get(name)
	if !found {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Function(args)
}

// GetTool retrieves a tool by name (for ToolRegistry)
func (tr *ToolRegistry) GetTool(name string) (*ToolInfo, bool) {
	return tr.Get(name)
}

// List returns all registered tool names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GetDocumentation returns formatted documentation for all tools
func (r *Registry) GetDocumentation() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc := "Available Tools:\n\n"

	for _, tool := range r.tools {
		doc += fmt.Sprintf("## %s\n", tool.Name)
		doc += fmt.Sprintf("%s\n\n", tool.Description)

		if len(tool.Parameters) > 0 {
			doc += "Parameters:\n"
			for _, param := range tool.Parameters {
				required := ""
				if param.Required {
					required = " (required)"
				}
				doc += fmt.Sprintf("  - %s (%s)%s\n", param.Name, param.Type, required)
			}
			doc += "\n"
		}
	}

	return doc
}

// ListTools returns a list of all registered tools
func (r *Registry) ListTools() []*ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}
