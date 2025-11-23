package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	sqlitetools "github.com/imran31415/godemode/sqlite-mcp-benchmark/generated"
)

// MCP Server - Exposes SQLite tools via JSON-RPC protocol

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

var registry *sqlitetools.Registry

func main() {
	port := ":8084"

	// Initialize database
	dbPath := "mcp-benchmark.db"
	if err := sqlitetools.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer sqlitetools.CloseDB()

	// Setup test data
	if err := setupTestData(); err != nil {
		log.Fatalf("Failed to setup test data: %v", err)
	}

	registry = sqlitetools.NewRegistry()

	http.HandleFunc("/mcp", mcpHandler)

	fmt.Printf("🚀 SQLite MCP Server starting on http://localhost%s/mcp\n", port)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📡 Exposing %d SQLite tools via MCP protocol\n", len(registry.List()))
	fmt.Println("📁 Database: mcp-benchmark.db")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	log.Fatal(http.ListenAndServe(port, nil))
}

func setupTestData() error {
	registry := sqlitetools.NewRegistry()

	// Create customers table
	_, err := registry.Call("query", map[string]interface{}{
		"sql": `CREATE TABLE IF NOT EXISTS customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			status TEXT DEFAULT 'pending',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
	})
	if err != nil {
		return fmt.Errorf("failed to create customers table: %w", err)
	}

	// Create products table
	_, err = registry.Call("query", map[string]interface{}{
		"sql": `CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL NOT NULL,
			stock_quantity INTEGER DEFAULT 0,
			category TEXT
		)`,
	})
	if err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	// Create orders table
	_, err = registry.Call("query", map[string]interface{}{
		"sql": `CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_id INTEGER NOT NULL,
			order_date TEXT DEFAULT CURRENT_TIMESTAMP,
			status TEXT DEFAULT 'pending',
			FOREIGN KEY (customer_id) REFERENCES customers(id)
		)`,
	})
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	// Create order_items table
	_, err = registry.Call("query", map[string]interface{}{
		"sql": `CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price REAL NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		)`,
	})
	if err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}

	// Seed customers
	customers := []map[string]interface{}{
		{"name": "Alice Smith", "email": "alice@example.com", "status": "active"},
		{"name": "Bob Johnson", "email": "bob@example.com", "status": "pending"},
		{"name": "Carol White", "email": "carol@example.com", "status": "active"},
		{"name": "David Brown", "email": "david@example.com", "status": "active"},
	}

	for _, c := range customers {
		registry.Call("create_record", map[string]interface{}{
			"table": "customers",
			"data":  c,
		})
	}

	// Seed products
	products := []map[string]interface{}{
		{"name": "Widget A", "price": 29.99, "stock_quantity": 100, "category": "widgets"},
		{"name": "Widget B", "price": 49.99, "stock_quantity": 50, "category": "widgets"},
		{"name": "Gadget A", "price": 99.99, "stock_quantity": 30, "category": "gadgets"},
		{"name": "Gadget B", "price": 149.99, "stock_quantity": 20, "category": "gadgets"},
		{"name": "Tool X", "price": 199.99, "stock_quantity": 15, "category": "tools"},
	}

	for _, p := range products {
		registry.Call("create_record", map[string]interface{}{
			"table": "products",
			"data":  p,
		})
	}

	// Seed orders
	orders := []map[string]interface{}{
		{"customer_id": 1, "status": "completed"},
		{"customer_id": 1, "status": "completed"},
		{"customer_id": 3, "status": "completed"},
		{"customer_id": 4, "status": "pending"},
	}

	for _, o := range orders {
		registry.Call("create_record", map[string]interface{}{
			"table": "orders",
			"data":  o,
		})
	}

	// Seed order_items
	orderItems := []map[string]interface{}{
		{"order_id": 1, "product_id": 1, "quantity": 3, "unit_price": 29.99},
		{"order_id": 1, "product_id": 3, "quantity": 1, "unit_price": 99.99},
		{"order_id": 2, "product_id": 2, "quantity": 2, "unit_price": 49.99},
		{"order_id": 3, "product_id": 4, "quantity": 1, "unit_price": 149.99},
		{"order_id": 3, "product_id": 5, "quantity": 1, "unit_price": 199.99},
		{"order_id": 4, "product_id": 1, "quantity": 5, "unit_price": 29.99},
	}

	for _, oi := range orderItems {
		registry.Call("create_record", map[string]interface{}{
			"table": "order_items",
			"data":  oi,
		})
	}

	return nil
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
	toolsList := registry.ListTools()
	mcpTools := make([]MCPTool, len(toolsList))

	for i, tool := range toolsList {
		properties := make(map[string]interface{})
		required := []string{}

		for _, param := range tool.Parameters {
			propType := "string"
			switch param.Type {
			case "float64":
				propType = "number"
			case "map[string]interface{}":
				propType = "object"
			case "[]interface{}":
				propType = "array"
			}

			properties[param.Name] = map[string]interface{}{
				"type":        propType,
				"description": param.Name,
			}

			if param.Required {
				required = append(required, param.Name)
			}
		}

		mcpTools[i] = MCPTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
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

	// Format result as MCP content
	resultJSON, _ := json.Marshal(result)
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": string(resultJSON),
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
