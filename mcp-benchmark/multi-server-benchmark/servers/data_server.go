package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	dataprocessing "github.com/imran31415/godemode/mcp-benchmark/data-processing"
)

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

type MCPServer struct {
	registry *dataprocessing.Registry
}

func NewMCPServer() *MCPServer {
	return &MCPServer{
		registry: dataprocessing.NewRegistry(),
	}
}

func (s *MCPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[Data MCP Server] Received: %s with params: %v", req.Method, req.Params)

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

	w.Header().Set("X-MCP-Duration-Ms", fmt.Sprintf("%.2f", time.Since(startTime).Seconds()*1000))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[Data MCP Server] Responded in %v", time.Since(startTime))
}

func (s *MCPServer) listTools() interface{} {
	return map[string]interface{}{
		"tools": []map[string]interface{}{
			{
				"name":        "filterArray",
				"description": "Filter an array based on a condition",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data":      map[string]interface{}{"type": "array", "description": "Array of numbers to filter"},
						"operation": map[string]interface{}{"type": "string", "description": "Operation: gt, lt, eq, gte, lte"},
						"value":     map[string]interface{}{"type": "number", "description": "Value to compare against"},
					},
					"required": []string{"data", "operation", "value"},
				},
			},
			{
				"name":        "mapArray",
				"description": "Transform each element in an array",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data":      map[string]interface{}{"type": "array", "description": "Array to transform"},
						"operation": map[string]interface{}{"type": "string", "description": "Operation: double, square, negate"},
					},
					"required": []string{"data", "operation"},
				},
			},
			{
				"name":        "reduceArray",
				"description": "Reduce an array to a single value",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data":      map[string]interface{}{"type": "array", "description": "Array of numbers"},
						"operation": map[string]interface{}{"type": "string", "description": "Operation: sum, product, max, min, avg"},
					},
					"required": []string{"data", "operation"},
				},
			},
			{
				"name":        "sortArray",
				"description": "Sort an array",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data":  map[string]interface{}{"type": "array", "description": "Array to sort"},
						"order": map[string]interface{}{"type": "string", "description": "Order: asc or desc"},
					},
					"required": []string{"data", "order"},
				},
			},
			{
				"name":        "mergeArrays",
				"description": "Merge multiple arrays",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"arrays": map[string]interface{}{"type": "array", "description": "Arrays to merge"},
					},
					"required": []string{"arrays"},
				},
			},
			{
				"name":        "uniqueValues",
				"description": "Get unique values from an array",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data": map[string]interface{}{"type": "array", "description": "Array with potential duplicates"},
					},
					"required": []string{"data"},
				},
			},
		},
	}
}

func main() {
	server := NewMCPServer()

	http.HandleFunc("/mcp", server.handleRequest)

	port := 8081
	log.Printf("🚀 Starting Data Processing MCP Server on port %d", port)
	log.Printf("📡 Endpoint: http://localhost:%d/mcp", port)
	log.Printf("📋 Tools: filterArray, mapArray, reduceArray, sortArray, mergeArrays, uniqueValues")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
