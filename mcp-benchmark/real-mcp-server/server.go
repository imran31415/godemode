package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	utilitytools "github.com/imran31415/godemode/mcp-benchmark/godemode"
)

// MCP JSON-RPC request/response types
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP Server
type MCPServer struct {
	registry *utilitytools.Registry
}

func NewMCPServer() *MCPServer {
	return &MCPServer{
		registry: utilitytools.NewRegistry(),
	}
}

// Handle MCP requests
func (s *MCPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Add latency tracking header
	startTime := time.Now()

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[MCP Server] Received: %s with params: %v", req.Method, req.Params)

	var response MCPResponse
	response.JSONRPC = "2.0"
	response.ID = req.ID

	switch req.Method {
	case "tools/list":
		response.Result = s.listTools()

	case "tools/call":
		toolName, ok := req.Params["name"].(string)
		if !ok {
			response.Error = &MCPError{Code: -32602, Message: "Invalid params: name required"}
			break
		}

		args, ok := req.Params["arguments"].(map[string]interface{})
		if !ok {
			args = make(map[string]interface{})
		}

		result, err := s.registry.Call(toolName, args)
		if err != nil {
			response.Error = &MCPError{Code: -32000, Message: err.Error()}
		} else {
			response.Result = result
		}

	default:
		response.Error = &MCPError{Code: -32601, Message: "Method not found"}
	}

	// Add timing header
	w.Header().Set("X-MCP-Duration-Ms", fmt.Sprintf("%.2f", time.Since(startTime).Seconds()*1000))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[MCP Server] Responded in %v", time.Since(startTime))
}

func (s *MCPServer) listTools() interface{} {
	return map[string]interface{}{
		"tools": []map[string]interface{}{
			{
				"name":        "add",
				"description": "Add two numbers together",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"a": map[string]interface{}{"type": "number", "description": "First number"},
						"b": map[string]interface{}{"type": "number", "description": "Second number"},
					},
					"required": []string{"a", "b"},
				},
			},
			{
				"name":        "getCurrentTime",
				"description": "Get the current time in RFC3339 format",
				"inputSchema": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			{
				"name":        "generateUUID",
				"description": "Generate a new UUID",
				"inputSchema": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			{
				"name":        "concatenateStrings",
				"description": "Concatenate an array of strings with a separator",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"strings":   map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
						"separator": map[string]interface{}{"type": "string", "description": "Separator between strings"},
					},
					"required": []string{"strings"},
				},
			},
			{
				"name":        "reverseString",
				"description": "Reverse a string",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"text": map[string]interface{}{"type": "string", "description": "Text to reverse"},
					},
					"required": []string{"text"},
				},
			},
		},
	}
}

func main() {
	server := NewMCPServer()

	http.HandleFunc("/mcp", server.handleRequest)

	port := 8080
	log.Printf("🚀 Starting MCP Server on port %d", port)
	log.Printf("📡 Endpoint: http://localhost:%d/mcp", port)
	log.Printf("📋 Tools: add, getCurrentTime, generateUUID, concatenateStrings, reverseString")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
