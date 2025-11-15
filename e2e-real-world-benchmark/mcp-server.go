package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/yourusername/godemode/e2e-real-world-benchmark/tools"
)

// MCP Server - Exposes tools via JSON-RPC protocol

type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
	ID      interface{}            `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolsList struct {
	Tools []MCPTool `json:"tools"`
}

type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

var registry *tools.Registry

func main() {
	port := ":8083"
	registry = tools.NewRegistry()

	http.HandleFunc("/mcp", mcpHandler)

	fmt.Printf("🚀 MCP Server starting on http://localhost%s/mcp\n", port)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📡 Exposing 12 e-commerce tools via MCP protocol")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	log.Fatal(http.ListenAndServe(port, nil))
}

func mcpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, -32700, "Parse error", req.ID)
		return
	}

	log.Printf("📨 RPC Request: %s", req.Method)

	var result interface{}
	var err error

	switch req.Method {
	case "tools/list":
		result = listTools()
	case "tools/call":
		result, err = callTool(req.Params)
	default:
		sendError(w, -32601, "Method not found", req.ID)
		return
	}

	if err != nil {
		sendError(w, -32000, err.Error(), req.ID)
		return
	}

	sendResponse(w, result, req.ID)
}

func listTools() ToolsList {
	toolsList := registry.List()
	mcpTools := make([]MCPTool, len(toolsList))

	for i, tool := range toolsList {
		mcpTools[i] = MCPTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": tool.Parameters,
			},
		}
	}

	return ToolsList{Tools: mcpTools}
}

func callTool(params map[string]interface{}) (interface{}, error) {
	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name required")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	log.Printf("   🔧 Executing tool: %s", toolName)
	result, err := registry.Call(toolName, arguments)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("%v", result),
			},
		},
	}, nil
}

func sendResponse(w http.ResponseWriter, result interface{}, id interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, code int, message string, id interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
