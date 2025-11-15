package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/imran31415/godemode/benchmark/mcp/protocol"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/tools"
)

// HTTPMCPServer implements MCP over HTTP
type HTTPMCPServer struct {
	registry        *tools.ToolRegistry
	initialized     bool
	mu              sync.Mutex
	protocolVersion string
	server          *http.Server
	port            int
}

// NewHTTPMCPServer creates a new HTTP-based MCP server
func NewHTTPMCPServer(env *scenarios.TestEnvironment, port int) *HTTPMCPServer {
	return &HTTPMCPServer{
		registry:        tools.NewToolRegistry(env),
		protocolVersion: "2024-11-05",
		port:            port,
	}
}

// Start begins listening for HTTP requests
func (s *HTTPMCPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	fmt.Printf("[MCP HTTP Server] Listening on port %d\n", s.port)
	return s.server.ListenAndServe()
}

// Stop shuts down the server
func (s *HTTPMCPServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// handleRequest processes HTTP POST requests containing JSON-RPC messages
func (s *HTTPMCPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendHTTPError(w, nil, protocol.ParseError, "Failed to read request")
		return
	}
	defer r.Body.Close()

	// Parse JSON-RPC message
	var msg protocol.JSONRPCMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		s.sendHTTPError(w, nil, protocol.ParseError, "Invalid JSON-RPC")
		return
	}

	fmt.Printf("[MCP HTTP Server] Received: %s\n", msg.Method)

	// Route to handler
	var response protocol.JSONRPCMessage

	switch msg.Method {
	case "initialize":
		response = s.handleInitialize(msg.ID, msg.Params)
	case "tools/list":
		response = s.handleListTools(msg.ID, msg.Params)
	case "tools/call":
		response = s.handleCallTool(msg.ID, msg.Params)
	case "ping":
		response = s.handlePing(msg.ID)
	default:
		response = s.createErrorResponse(msg.ID, protocol.MethodNotFound, fmt.Sprintf("Method not found: %s", msg.Method))
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleInitialize processes the initialize request
func (s *HTTPMCPServer) handleInitialize(id interface{}, params interface{}) protocol.JSONRPCMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := protocol.InitializeResult{
		ProtocolVersion: s.protocolVersion,
		Capabilities: protocol.ServerCapabilities{
			Tools: &protocol.ToolsCapability{
				Supported: true,
			},
			Logging: &protocol.LoggingCapability{
				Supported: true,
			},
		},
		ServerInfo: protocol.ServerInfo{
			Name:    "godemode-mcp-http-server",
			Version: "1.0.0",
		},
	}

	s.initialized = true
	fmt.Println("[MCP HTTP Server] Initialized successfully")

	return protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// handleListTools returns the list of available tools
func (s *HTTPMCPServer) handleListTools(id interface{}, params interface{}) protocol.JSONRPCMessage {
	if !s.initialized {
		return s.createErrorResponse(id, protocol.InvalidRequest, "Server not initialized")
	}

	registryTools := s.registry.ListTools()
	mcpTools := make([]protocol.Tool, 0, len(registryTools))

	for _, tool := range registryTools {
		properties := make(map[string]protocol.Property)
		required := []string{}

		for _, param := range tool.Parameters {
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
			}

			properties[param.Name] = protocol.Property{
				Type:        jsonType,
				Description: fmt.Sprintf("%s parameter", param.Name),
			}
			if param.Required {
				required = append(required, param.Name)
			}
		}

		if len(properties) == 0 {
			properties["_unused"] = protocol.Property{
				Type:        "string",
				Description: "Unused parameter",
			}
		}

		mcpTools = append(mcpTools, protocol.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: protocol.InputSchema{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		})
	}

	result := protocol.ListToolsResult{
		Tools: mcpTools,
	}

	fmt.Printf("[MCP HTTP Server] Returning %d tools\n", len(mcpTools))

	return protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// handleCallTool executes a tool and returns the result
func (s *HTTPMCPServer) handleCallTool(id interface{}, params interface{}) protocol.JSONRPCMessage {
	if !s.initialized {
		return s.createErrorResponse(id, protocol.InvalidRequest, "Server not initialized")
	}

	// Parse call tool request
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return s.createErrorResponse(id, protocol.InvalidParams, "Invalid parameters")
	}

	var req protocol.CallToolRequest
	if err := json.Unmarshal(paramsJSON, &req); err != nil {
		return s.createErrorResponse(id, protocol.InvalidParams, "Invalid call tool request")
	}

	fmt.Printf("[MCP HTTP Server] Calling tool: %s\n", req.Name)

	// Get tool from registry
	tool, exists := s.registry.GetTool(req.Name)
	if !exists {
		return s.createErrorResponse(id, protocol.InvalidParams, fmt.Sprintf("Tool not found: %s", req.Name))
	}

	// Execute tool
	result, err := tool.Function(req.Arguments)

	var callResult protocol.CallToolResult
	if err != nil {
		callResult = protocol.CallToolResult{
			Content: []protocol.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
			IsError: true,
		}
	} else {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return s.createErrorResponse(id, protocol.InternalError, fmt.Sprintf("Failed to marshal result: %v", err))
		}

		callResult = protocol.CallToolResult{
			Content: []protocol.Content{
				{
					Type:     "text",
					Text:     string(resultJSON),
					MimeType: "application/json",
				},
			},
			IsError: false,
		}
	}

	return protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  callResult,
	}
}

// handlePing responds to ping requests
func (s *HTTPMCPServer) handlePing(id interface{}) protocol.JSONRPCMessage {
	return protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"status": "ok",
		},
	}
}

// createErrorResponse creates a JSON-RPC error response
func (s *HTTPMCPServer) createErrorResponse(id interface{}, code int, message string) protocol.JSONRPCMessage {
	return protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &protocol.RPCError{
			Code:    code,
			Message: message,
		},
	}
}

// sendHTTPError sends an HTTP error response
func (s *HTTPMCPServer) sendHTTPError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := s.createErrorResponse(id, code, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors still return 200
	json.NewEncoder(w).Encode(response)
}
