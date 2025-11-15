package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/imran31415/godemode/benchmark/mcp/protocol"
)

// HTTPMCPClient communicates with an MCP server over HTTP
type HTTPMCPClient struct {
	baseURL    string
	httpClient *http.Client
	requestID  int64
}

// NewHTTPMCPClient creates a new HTTP-based MCP client
func NewHTTPMCPClient(baseURL string) *HTTPMCPClient {
	return &HTTPMCPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize sends the initialize request
func (c *HTTPMCPClient) Initialize() error {
	req := protocol.InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities: protocol.ClientCapabilities{
			Tools: &protocol.ToolsCapability{
				Supported: true,
			},
		},
		ClientInfo: protocol.ClientInfo{
			Name:    "godemode-mcp-http-client",
			Version: "1.0.0",
		},
	}

	var result protocol.InitializeResult
	if err := c.call("initialize", req, &result); err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	fmt.Printf("[MCP HTTP Client] Initialized with server: %s v%s\n", result.ServerInfo.Name, result.ServerInfo.Version)
	return nil
}

// ListTools retrieves the list of available tools
func (c *HTTPMCPClient) ListTools() ([]protocol.Tool, error) {
	req := protocol.ListToolsRequest{}

	var result protocol.ListToolsResult
	if err := c.call("tools/list", req, &result); err != nil {
		return nil, fmt.Errorf("list tools failed: %w", err)
	}

	fmt.Printf("[MCP HTTP Client] Retrieved %d tools\n", len(result.Tools))
	return result.Tools, nil
}

// CallTool executes a tool on the server
func (c *HTTPMCPClient) CallTool(name string, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	req := protocol.CallToolRequest{
		Name:      name,
		Arguments: arguments,
	}

	var result protocol.CallToolResult
	if err := c.call("tools/call", req, &result); err != nil {
		return nil, fmt.Errorf("call tool failed: %w", err)
	}

	return &result, nil
}

// call sends a JSON-RPC request over HTTP and waits for the response
func (c *HTTPMCPClient) call(method string, params interface{}, result interface{}) error {
	// Generate request ID
	id := atomic.AddInt64(&c.requestID, 1)

	// Create JSON-RPC request
	rpcReq := protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Marshal request
	reqBody, err := json.Marshal(rpcReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP POST request
	httpResp, err := c.httpClient.Post(c.baseURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Check HTTP status
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", httpResp.StatusCode)
	}

	// Parse JSON-RPC response
	var rpcResp protocol.JSONRPCMessage
	if err := json.NewDecoder(httpResp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// Unmarshal result
	if result != nil {
		resultJSON, err := json.Marshal(rpcResp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		if err := json.Unmarshal(resultJSON, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// Close closes the client (no-op for HTTP)
func (c *HTTPMCPClient) Close() error {
	return nil
}
