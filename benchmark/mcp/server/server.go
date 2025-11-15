package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/imran31415/godemode/benchmark/mcp/protocol"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/tools"
)

// MCPServer implements the Model Context Protocol server
type MCPServer struct {
	registry        *tools.ToolRegistry
	reader          *bufio.Reader
	writer          *bufio.Writer
	initialized     bool
	mu              sync.Mutex
	protocolVersion string
}

// NewMCPServer creates a new MCP server
func NewMCPServer(env *scenarios.TestEnvironment) *MCPServer {
	return &MCPServer{
		registry:        tools.NewToolRegistry(env),
		reader:          bufio.NewReader(os.Stdin),
		writer:          bufio.NewWriter(os.Stdout),
		protocolVersion: "2024-11-05",
	}
}

// Start begins listening for JSON-RPC messages on stdin
func (s *MCPServer) Start() error {
	fmt.Fprintln(os.Stderr, "[MCP Server] Starting...")

	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read message: %w", err)
		}

		if err := s.handleMessage(line); err != nil {
			fmt.Fprintf(os.Stderr, "[MCP Server] Error handling message: %v\n", err)
			// Send error response
			s.sendError(nil, protocol.InternalError, err.Error())
		}
	}
}

// handleMessage processes a single JSON-RPC message
func (s *MCPServer) handleMessage(data []byte) error {
	var msg protocol.JSONRPCMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		s.sendError(nil, protocol.ParseError, "Parse error")
		return fmt.Errorf("failed to parse JSON-RPC message: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[MCP Server] Received method: %s\n", msg.Method)

	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg.ID, msg.Params)
	case "tools/list":
		return s.handleListTools(msg.ID, msg.Params)
	case "tools/call":
		return s.handleCallTool(msg.ID, msg.Params)
	case "ping":
		return s.handlePing(msg.ID)
	default:
		s.sendError(msg.ID, protocol.MethodNotFound, fmt.Sprintf("Method not found: %s", msg.Method))
		return fmt.Errorf("unknown method: %s", msg.Method)
	}
}

// handleInitialize processes the initialize request
func (s *MCPServer) handleInitialize(id interface{}, params interface{}) error {
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
			Name:    "godemode-mcp-server",
			Version: "1.0.0",
		},
	}

	s.initialized = true
	fmt.Fprintln(os.Stderr, "[MCP Server] Initialized successfully")

	return s.sendResult(id, result)
}

// handleListTools returns the list of available tools
func (s *MCPServer) handleListTools(id interface{}, params interface{}) error {
	if !s.initialized {
		return s.sendError(id, protocol.InvalidRequest, "Server not initialized")
	}

	registryTools := s.registry.ListTools()
	mcpTools := make([]protocol.Tool, 0, len(registryTools))

	for _, tool := range registryTools {
		properties := make(map[string]protocol.Property)
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
			}

			properties[param.Name] = protocol.Property{
				Type:        jsonType,
				Description: fmt.Sprintf("%s parameter", param.Name),
			}
			if param.Required {
				required = append(required, param.Name)
			}
		}

		// Add dummy parameter if none exist
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

	fmt.Fprintf(os.Stderr, "[MCP Server] Returning %d tools\n", len(mcpTools))
	return s.sendResult(id, result)
}

// handleCallTool executes a tool and returns the result
func (s *MCPServer) handleCallTool(id interface{}, params interface{}) error {
	if !s.initialized {
		return s.sendError(id, protocol.InvalidRequest, "Server not initialized")
	}

	// Parse call tool request
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return s.sendError(id, protocol.InvalidParams, "Invalid parameters")
	}

	var req protocol.CallToolRequest
	if err := json.Unmarshal(paramsJSON, &req); err != nil {
		return s.sendError(id, protocol.InvalidParams, "Invalid call tool request")
	}

	fmt.Fprintf(os.Stderr, "[MCP Server] Calling tool: %s\n", req.Name)

	// Get tool from registry
	tool, exists := s.registry.GetTool(req.Name)
	if !exists {
		return s.sendError(id, protocol.InvalidParams, fmt.Sprintf("Tool not found: %s", req.Name))
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
		// Convert result to JSON string
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return s.sendError(id, protocol.InternalError, fmt.Sprintf("Failed to marshal result: %v", err))
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

	return s.sendResult(id, callResult)
}

// handlePing responds to ping requests
func (s *MCPServer) handlePing(id interface{}) error {
	return s.sendResult(id, map[string]interface{}{
		"status": "ok",
	})
}

// sendResult sends a successful JSON-RPC result
func (s *MCPServer) sendResult(id interface{}, result interface{}) error {
	response := protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	return s.sendMessage(response)
}

// sendError sends a JSON-RPC error response
func (s *MCPServer) sendError(id interface{}, code int, message string) error {
	response := protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &protocol.RPCError{
			Code:    code,
			Message: message,
		},
	}

	return s.sendMessage(response)
}

// sendMessage writes a JSON-RPC message to stdout
func (s *MCPServer) sendMessage(msg protocol.JSONRPCMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = fmt.Fprintf(s.writer, "%s\n", data)
	if err != nil {
		return err
	}

	// CRITICAL: Flush immediately so client receives the response
	// Without this, responses sit in buffer and client hangs
	return s.writer.Flush()
}
