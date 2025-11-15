package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/imran31415/godemode/benchmark/mcp/protocol"
)

// MCPClient handles communication with an MCP server
type MCPClient struct {
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	reader      *bufio.Reader
	requestID   int64
	pendingReqs map[int64]chan *protocol.JSONRPCMessage
	mu          sync.Mutex
	closed      bool
}

// NewMCPClient creates a new MCP client that launches a server process
func NewMCPClient(serverCommand string, args ...string) (*MCPClient, error) {
	cmd := exec.Command(serverCommand, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	client := &MCPClient{
		cmd:         cmd,
		stdin:       stdin,
		stdout:      stdout,
		stderr:      stderr,
		reader:      bufio.NewReader(stdout),
		pendingReqs: make(map[int64]chan *protocol.JSONRPCMessage),
	}

	// Start server process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	// Start reading responses in background
	go client.readLoop()

	// Forward stderr for debugging
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[MCP Server] %s\n", scanner.Text())
		}
	}()

	return client, nil
}

// Initialize sends the initialize request
func (c *MCPClient) Initialize() error {
	req := protocol.InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities: protocol.ClientCapabilities{
			Tools: &protocol.ToolsCapability{
				Supported: true,
			},
		},
		ClientInfo: protocol.ClientInfo{
			Name:    "godemode-mcp-client",
			Version: "1.0.0",
		},
	}

	var result protocol.InitializeResult
	if err := c.call("initialize", req, &result); err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	fmt.Printf("[MCP Client] Initialized with server: %s v%s\n", result.ServerInfo.Name, result.ServerInfo.Version)
	return nil
}

// ListTools retrieves the list of available tools
func (c *MCPClient) ListTools() ([]protocol.Tool, error) {
	req := protocol.ListToolsRequest{}

	var result protocol.ListToolsResult
	if err := c.call("tools/list", req, &result); err != nil {
		return nil, fmt.Errorf("list tools failed: %w", err)
	}

	fmt.Printf("[MCP Client] Retrieved %d tools\n", len(result.Tools))
	return result.Tools, nil
}

// CallTool executes a tool on the server
func (c *MCPClient) CallTool(name string, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
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

// call sends a JSON-RPC request and waits for the response
func (c *MCPClient) call(method string, params interface{}, result interface{}) error {
	if c.closed {
		return fmt.Errorf("client is closed")
	}

	// Generate request ID
	id := atomic.AddInt64(&c.requestID, 1)
	fmt.Printf("[MCP Client] Calling method: %s (ID: %d)\n", method, id)

	// Create response channel
	respChan := make(chan *protocol.JSONRPCMessage, 1)
	c.mu.Lock()
	c.pendingReqs[id] = respChan
	c.mu.Unlock()

	// Ensure cleanup
	defer func() {
		c.mu.Lock()
		delete(c.pendingReqs, id)
		c.mu.Unlock()
	}()

	// Send request
	req := protocol.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	c.mu.Lock()
	_, err = fmt.Fprintf(c.stdin, "%s\n", data)
	c.mu.Unlock()
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	fmt.Printf("[MCP Client] Request sent, waiting for response...\n")
	// Wait for response
	resp := <-respChan
	fmt.Printf("[MCP Client] Response received for ID: %d\n", id)

	// Check for error
	if resp.Error != nil {
		return fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	// Unmarshal result
	if result != nil {
		resultJSON, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		if err := json.Unmarshal(resultJSON, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// readLoop continuously reads responses from the server
func (c *MCPClient) readLoop() {
	fmt.Printf("[MCP Client] ReadLoop started\n")
	for {
		fmt.Printf("[MCP Client] ReadLoop waiting for data...\n")
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[MCP Client] Read error: %v\n", err)
			}
			fmt.Printf("[MCP Client] ReadLoop exiting\n")
			return
		}

		fmt.Printf("[MCP Client] Received response: %s\n", string(line))

		var msg protocol.JSONRPCMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			fmt.Printf("[MCP Client] Failed to parse response: %v\n", err)
			continue
		}

		// Find pending request
		if msg.ID != nil {
			var id int64
			switch v := msg.ID.(type) {
			case float64:
				id = int64(v)
			case int64:
				id = v
			case int:
				id = int64(v)
			default:
				fmt.Printf("[MCP Client] Unknown ID type: %T\n", msg.ID)
				continue
			}

			fmt.Printf("[MCP Client] Looking for pending request ID: %d\n", id)
			c.mu.Lock()
			respChan, exists := c.pendingReqs[id]
			fmt.Printf("[MCP Client] Pending request exists: %v\n", exists)
			c.mu.Unlock()

			if exists {
				fmt.Printf("[MCP Client] Sending response to channel for ID: %d\n", id)
				respChan <- &msg
			}
		} else {
			fmt.Printf("[MCP Client] Response has no ID\n")
		}
	}
}

// Close shuts down the MCP client and server
func (c *MCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Close stdin to signal server to shutdown
	c.stdin.Close()

	// Wait for server to exit
	return c.cmd.Wait()
}
