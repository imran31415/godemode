# SQLite MCP Benchmark Results

## CodeMode vs Native Tool Calling

This benchmark compares **CodeMode** (generating complete Go code) against **Native Tool Calling** (sequential API calls) using the popular [jparkerweb/mcp-sqlite](https://github.com/jparkerweb/mcp-sqlite) MCP server.

---

## Executive Summary

### Simple Scenario (5 operations)

| Metric | CodeMode | Tool Calling | Improvement |
|--------|----------|--------------|-------------|
| **Duration** | 8.9s | 17.1s | **1.92x faster** |
| **API Calls** | 1 | 6 | **6x fewer** |
| **Total Tokens** | 1,345 | 9,416 | **85.7% fewer** |
| **Estimated Cost** | $0.013 | $0.037 | **64.8% cheaper** |

### Complex Scenario (19 operations)

| Metric | CodeMode | Tool Calling | Improvement |
|--------|----------|--------------|-------------|
| **Duration** | 37.1s | 50.3s | **1.35x faster** |
| **API Calls** | 1 | 13 | **13x fewer** |
| **Total Tokens** | 4,187 | 44,002 | **90.5% fewer** |
| **Estimated Cost** | $0.054 | $0.163 | **66.6% cheaper** |

**Key Finding**: CodeMode achieves **90% token reduction** and **67% cost savings** for complex multi-table operations.

---

## Test Scenarios

### Simple CRUD Operations

```
1. List all tables in the database
2. Get the schema of the customers table
3. Read all customers with status 'active'
4. Create a new customer named "Test User" with email "test@example.com"
5. Read all active customers again to confirm the new customer was added
```

### Complex Multi-Table Analysis

```
1. List all tables to understand the schema
2. Get the schema of all 4 tables (customers, products, orders, order_items)
3. Find all customers who have placed orders (join customers and orders)
4. Calculate total revenue per customer by joining orders with order_items and products
5. Find the top-selling product by quantity
6. Create a new order for customer "Alice Smith" with 2 units of "Widget A" and 1 unit of "Gadget B"
7. Update the stock quantity for the ordered products
8. Verify the new order was created by reading it back with all its items
9. Generate a summary report showing: total customers, total products, total orders, and total revenue
```

**Database Setup**: Pre-seeded with 4 customers, 5 products, 4 orders, and 6 order items.

---

## MCP Tools Used

Based on [jparkerweb/mcp-sqlite](https://github.com/jparkerweb/mcp-sqlite):

| Tool | Description |
|------|-------------|
| `list_tables` | List all tables in the database |
| `get_table_schema` | Get column names, types, constraints |
| `read_records` | Query records with filters, pagination |
| `create_record` | Insert a new record |
| `update_records` | Update matching records |
| `delete_records` | Delete matching records |
| `query` | Execute custom SQL |
| `db_info` | Get database information |

---

## Tool Calling Sequence (Complex Scenario)

Here's the actual sequence of 19 tool calls made during the complex scenario:

```
[13:05:05] API_CALL #1
[13:05:07] API_RESPONSE (stop_reason: tool_use) - tokens: in=1140, out=67

[13:05:07] TOOL_CALL #1: list_tables
           Args: {}
           Result: {"count":4,"tables":["customers","order_items","orders","products"]}

[13:05:07] API_CALL #2
[13:05:11] API_RESPONSE (stop_reason: tool_use) - tokens: in=1237, out=189

[13:05:11] TOOL_CALL #2: get_table_schema
           Args: {"tableName":"customers"}
           Result: {"columns":[{"name":"id","type":"INTEGER","primary_key":true},...]}

[13:05:11] TOOL_CALL #3: get_table_schema
           Args: {"tableName":"products"}
           Result: {"columns":[{"name":"id","type":"INTEGER"},{"name":"name","type":"TEXT"},...]}

[13:05:11] TOOL_CALL #4: get_table_schema
           Args: {"tableName":"orders"}
           Result: {"columns":[{"name":"id","type":"INTEGER"},{"name":"customer_id","type":"INTEGER"},...]}

[13:05:11] TOOL_CALL #5: get_table_schema
           Args: {"tableName":"order_items"}
           Result: {"columns":[{"name":"order_id","type":"INTEGER"},{"name":"product_id","type":"INTEGER"},...]}

[13:05:11] API_CALL #3
[13:05:14] API_RESPONSE (stop_reason: tool_use) - tokens: in=2138, out=108

[13:05:14] TOOL_CALL #6: query
           Args: {"sql":"SELECT DISTINCT c.id, c.name, c.email FROM customers c INNER JOIN orders o ON c.id = o.customer_id"}
           Result: {"count":3,"records":[{"id":1,"name":"Alice Smith"},{"id":3,"name":"Carol White"},...]}

[13:05:14] API_CALL #4
[13:05:17] API_RESPONSE (stop_reason: tool_use) - tokens: in=2331, out=170

[13:05:17] TOOL_CALL #7: query
           Args: {"sql":"SELECT c.name, SUM(oi.quantity * oi.unit_price) as total_revenue FROM customers c JOIN orders o ON c.id = o.customer_id JOIN order_items oi ON o.id = oi.order_id GROUP BY c.id"}
           Result: {"records":[{"name":"Carol White","total_revenue":349.98},{"name":"Alice Smith","total_revenue":289.94},...]}

[13:05:17] API_CALL #5
[13:05:20] API_RESPONSE (stop_reason: tool_use) - tokens: in=2618, out=169

[13:05:20] TOOL_CALL #8: query
           Args: {"sql":"SELECT p.name, SUM(oi.quantity) as total_sold FROM products p JOIN order_items oi ON p.id = oi.product_id GROUP BY p.id ORDER BY total_sold DESC"}
           Result: {"records":[{"name":"Widget A","total_sold":8},{"name":"Widget B","total_sold":2},...]}

[13:05:20] API_CALL #6
[13:05:24] API_RESPONSE (stop_reason: tool_use) - tokens: in=2984, out=224

[13:05:24] TOOL_CALL #9: read_records
           Args: {"table":"customers","conditions":{"name":"Alice Smith"}}
           Result: {"records":[{"id":1,"name":"Alice Smith","email":"alice@example.com"}]}

[13:05:24] TOOL_CALL #10: read_records
           Args: {"table":"products","conditions":{"name":"Widget A"}}
           Result: {"records":[{"id":1,"name":"Widget A","price":29.99,"stock_quantity":100}]}

[13:05:24] TOOL_CALL #11: read_records
           Args: {"table":"products","conditions":{"name":"Gadget B"}}
           Result: {"records":[{"id":4,"name":"Gadget B","price":149.99,"stock_quantity":20}]}

[13:05:24] API_CALL #7
[13:05:27] API_RESPONSE (stop_reason: tool_use) - tokens: in=3417, out=94

[13:05:27] TOOL_CALL #12: create_record
           Args: {"table":"orders","data":{"customer_id":1,"status":"pending"}}
           Result: {"last_id":5,"rows_affected":1,"success":true}

[13:05:27] API_CALL #8
[13:05:31] API_RESPONSE (stop_reason: tool_use) - tokens: in=3539, out=196

[13:05:31] TOOL_CALL #13: create_record
           Args: {"table":"order_items","data":{"order_id":5,"product_id":1,"quantity":2,"unit_price":29.99}}
           Result: {"last_id":7,"rows_affected":1,"success":true}

[13:05:31] TOOL_CALL #14: create_record
           Args: {"table":"order_items","data":{"order_id":5,"product_id":4,"quantity":1,"unit_price":149.99}}
           Result: {"last_id":8,"rows_affected":1,"success":true}

[13:05:31] API_CALL #9
[13:05:34] API_RESPONSE (stop_reason: tool_use) - tokens: in=3830, out=190

[13:05:34] TOOL_CALL #15: update_records
           Args: {"table":"products","conditions":{"id":1},"data":{"stock_quantity":98}}
           Result: {"rows_affected":1,"success":true}

[13:05:34] TOOL_CALL #16: update_records
           Args: {"table":"products","conditions":{"id":4},"data":{"stock_quantity":19}}
           Result: {"rows_affected":1,"success":true}

[13:05:34] API_CALL #10
[13:05:38] API_RESPONSE (stop_reason: tool_use) - tokens: in=4103, out=207

[13:05:38] TOOL_CALL #17: query
           Args: {"sql":"SELECT o.id, c.name, p.name as product, oi.quantity, oi.unit_price FROM orders o JOIN customers c ON o.customer_id = c.id JOIN order_items oi ON o.id = oi.order_id JOIN products p ON oi.product_id = p.id WHERE o.id = 5"}
           Result: {"records":[{"id":5,"name":"Alice Smith","product":"Widget A","quantity":2},{"id":5,"name":"Alice Smith","product":"Gadget B","quantity":1}]}

[13:05:38] API_CALL #11
[13:05:41] API_RESPONSE (stop_reason: tool_use) - tokens: in=4473, out=144

[13:05:41] TOOL_CALL #18: query
           Args: {"sql":"SELECT (SELECT COUNT(*) FROM customers) as total_customers, (SELECT COUNT(*) FROM products) as total_products, (SELECT COUNT(*) FROM orders) as total_orders, (SELECT SUM(quantity * unit_price) FROM order_items) as total_revenue"}
           Result: {"records":[{"total_customers":5,"total_orders":5,"total_products":5,"total_revenue":999.84}]}

[13:05:41] API_CALL #12
[13:05:45] API_RESPONSE (stop_reason: tool_use) - tokens: in=4665, out=205

[13:05:45] TOOL_CALL #19: query
           Args: {"sql":"SELECT COUNT(DISTINCT o.customer_id) as active_customers, AVG(totals.total) as avg_order_value FROM orders o JOIN (SELECT order_id, SUM(quantity * unit_price) as total FROM order_items GROUP BY order_id) totals ON o.id = totals.order_id"}
           Result: {"records":[{"active_customers":3,"avg_order_value":199.97}]}

[13:05:45] API_CALL #13
[13:05:55] API_RESPONSE (stop_reason: end_turn) - tokens: in=4932, out=632
```

**Total Tool Calling Stats**: 13 API calls, 19 tool executions, 44,002 tokens

---

## CodeMode Generated Code (Complex Scenario)

Here's the actual code generated by CodeMode for the complex scenario:

```go
package main

import (
    "fmt"
    "log"
)

// Registry interface for calling tools
type Registry interface {
    Call(toolName string, args map[string]interface{}) (interface{}, error)
}

var registry Registry

func main() {
    // 1. List all tables to understand the schema
    fmt.Println("=== 1. LISTING ALL TABLES ===")
    tables, err := registry.Call("list_tables", nil)
    if err != nil {
        log.Fatal("Error listing tables:", err)
    }
    fmt.Println("Tables:", tables)

    // 2. Get the schema of all 4 tables
    fmt.Println("\n=== 2. TABLE SCHEMAS ===")
    tableNames := []string{"customers", "products", "orders", "order_items"}

    for _, tableName := range tableNames {
        schema, err := registry.Call("get_table_schema", map[string]interface{}{
            "tableName": tableName,
        })
        if err != nil {
            log.Printf("Error getting schema for %s: %v", tableName, err)
            continue
        }
        fmt.Printf("\n%s schema: %v\n", tableName, schema)
    }

    // 3. Find all customers who have placed orders
    fmt.Println("\n=== 3. CUSTOMERS WITH ORDERS ===")
    customersWithOrders, err := registry.Call("query", map[string]interface{}{
        "sql": `SELECT DISTINCT c.id, c.name, c.email, c.status
                FROM customers c
                INNER JOIN orders o ON c.id = o.customer_id`,
    })
    if err != nil {
        log.Fatal("Error finding customers with orders:", err)
    }
    fmt.Println("Customers with orders:", customersWithOrders)

    // 4. Calculate total revenue per customer
    fmt.Println("\n=== 4. REVENUE PER CUSTOMER ===")
    revenuePerCustomer, err := registry.Call("query", map[string]interface{}{
        "sql": `SELECT c.id, c.name, c.email,
                       COUNT(DISTINCT o.id) as total_orders,
                       SUM(oi.quantity * oi.unit_price) as total_revenue
                FROM customers c
                JOIN orders o ON c.id = o.customer_id
                JOIN order_items oi ON o.id = oi.order_id
                GROUP BY c.id, c.name, c.email
                ORDER BY total_revenue DESC`,
    })
    if err != nil {
        log.Fatal("Error calculating revenue:", err)
    }
    fmt.Println("Revenue per customer:", revenuePerCustomer)

    // 5. Find the top-selling product by quantity
    fmt.Println("\n=== 5. TOP-SELLING PRODUCTS ===")
    topProducts, err := registry.Call("query", map[string]interface{}{
        "sql": `SELECT p.id, p.name, p.price, p.category,
                       SUM(oi.quantity) as total_sold,
                       SUM(oi.quantity * oi.unit_price) as total_revenue
                FROM products p
                JOIN order_items oi ON p.id = oi.product_id
                GROUP BY p.id, p.name, p.price, p.category
                ORDER BY total_sold DESC`,
    })
    if err != nil {
        log.Fatal("Error finding top products:", err)
    }
    fmt.Println("Top-selling products:", topProducts)

    // 6. Create a new order for Alice Smith
    fmt.Println("\n=== 6. CREATING NEW ORDER ===")

    // Get Alice's customer ID
    aliceResult, err := registry.Call("read_records", map[string]interface{}{
        "table": "customers",
        "conditions": map[string]interface{}{
            "name": "Alice Smith",
        },
    })
    if err != nil {
        log.Fatal("Error finding Alice:", err)
    }
    fmt.Println("Found Alice:", aliceResult)

    // Get product IDs
    widgetA, _ := registry.Call("read_records", map[string]interface{}{
        "table": "products",
        "conditions": map[string]interface{}{"name": "Widget A"},
    })
    gadgetB, _ := registry.Call("read_records", map[string]interface{}{
        "table": "products",
        "conditions": map[string]interface{}{"name": "Gadget B"},
    })

    // Create order
    newOrder, err := registry.Call("create_record", map[string]interface{}{
        "table": "orders",
        "data": map[string]interface{}{
            "customer_id": 1,  // Alice's ID
            "status":      "pending",
        },
    })
    if err != nil {
        log.Fatal("Error creating order:", err)
    }
    fmt.Println("Created order:", newOrder)

    // Create order items (2 Widget A, 1 Gadget B)
    _, err = registry.Call("create_record", map[string]interface{}{
        "table": "order_items",
        "data": map[string]interface{}{
            "order_id":   5,  // New order ID
            "product_id": 1,  // Widget A
            "quantity":   2,
            "unit_price": 29.99,
        },
    })
    _, err = registry.Call("create_record", map[string]interface{}{
        "table": "order_items",
        "data": map[string]interface{}{
            "order_id":   5,
            "product_id": 4,  // Gadget B
            "quantity":   1,
            "unit_price": 149.99,
        },
    })
    fmt.Println("Added order items")

    // 7. Update stock quantities
    fmt.Println("\n=== 7. UPDATING STOCK ===")
    registry.Call("update_records", map[string]interface{}{
        "table":      "products",
        "conditions": map[string]interface{}{"id": 1},
        "data":       map[string]interface{}{"stock_quantity": 98},  // 100 - 2
    })
    registry.Call("update_records", map[string]interface{}{
        "table":      "products",
        "conditions": map[string]interface{}{"id": 4},
        "data":       map[string]interface{}{"stock_quantity": 19},  // 20 - 1
    })
    fmt.Println("Stock updated")

    // 8. Verify the new order
    fmt.Println("\n=== 8. VERIFYING ORDER ===")
    orderDetails, err := registry.Call("query", map[string]interface{}{
        "sql": `SELECT o.id as order_id, o.order_date, o.status,
                       c.name as customer_name, c.email,
                       p.name as product_name, oi.quantity, oi.unit_price,
                       (oi.quantity * oi.unit_price) as line_total
                FROM orders o
                JOIN customers c ON o.customer_id = c.id
                JOIN order_items oi ON o.id = oi.order_id
                JOIN products p ON oi.product_id = p.id
                WHERE o.id = 5`,
    })
    if err != nil {
        log.Fatal("Error verifying order:", err)
    }
    fmt.Println("Order details:", orderDetails)

    // 9. Generate summary report
    fmt.Println("\n=== 9. SUMMARY REPORT ===")
    summary, err := registry.Call("query", map[string]interface{}{
        "sql": `SELECT
                    (SELECT COUNT(*) FROM customers) as total_customers,
                    (SELECT COUNT(*) FROM products) as total_products,
                    (SELECT COUNT(*) FROM orders) as total_orders,
                    (SELECT SUM(quantity * unit_price) FROM order_items) as total_revenue`,
    })
    if err != nil {
        log.Fatal("Error generating summary:", err)
    }
    fmt.Println("Summary:", summary)

    // Additional metrics
    metrics, _ := registry.Call("query", map[string]interface{}{
        "sql": `SELECT
                    COUNT(DISTINCT o.customer_id) as active_customers,
                    AVG(order_totals.total) as avg_order_value,
                    MAX(order_totals.total) as max_order_value,
                    MIN(order_totals.total) as min_order_value
                FROM orders o
                JOIN (
                    SELECT order_id, SUM(quantity * unit_price) as total
                    FROM order_items
                    GROUP BY order_id
                ) order_totals ON o.id = order_totals.order_id`,
    })
    fmt.Println("Additional metrics:", metrics)
}
```

**Total CodeMode Stats**: 1 API call, 21 tool calls in generated code, 4,187 tokens

---

## How Code Gets Executed in the Interpreter

The generated code is executed using the **Yaegi Go interpreter** with the tool registry injected:

### 1. Code Generation
```go
// Claude generates the complete Go program
resp, err := callClaude(apiKey, systemPrompt + task, nil)
generatedCode := resp.Content[0].Text
```

### 2. Registry Injection
```go
// Create the tool registry with SQLite implementations
registry := sqlitetools.NewRegistry()

// The registry provides the Call interface
type Registry struct {
    tools map[string]*ToolInfo
}

func (r *Registry) Call(name string, args map[string]interface{}) (interface{}, error) {
    tool, found := r.Get(name)
    if !found {
        return nil, fmt.Errorf("tool not found: %s", name)
    }
    return tool.Function(args)  // Execute the actual SQLite operation
}
```

### 3. Interpreter Setup
```go
import "github.com/traefik/yaegi/interp"

// Create interpreter
i := interp.New(interp.Options{})

// Export the registry to the interpreter
i.Use(stdlib.Symbols)
i.Use(map[string]map[string]reflect.Value{
    "main/main": {
        "registry": reflect.ValueOf(registry),
    },
})
```

### 4. Code Execution
```go
// Evaluate the generated code
_, err := i.Eval(generatedCode)
if err != nil {
    return fmt.Errorf("execution error: %w", err)
}

// Run main()
v, err := i.Eval("main.main")
if err != nil {
    return fmt.Errorf("main error: %w", err)
}
```

### 5. Tool Execution Flow
```
Generated Code                    Interpreter                     Registry                        SQLite
      |                               |                              |                              |
      |-- registry.Call("query",..)-->|                              |                              |
      |                               |-- Call("query", args) ------>|                              |
      |                               |                              |-- Execute SQL --------------->|
      |                               |                              |<-- Results -------------------|
      |                               |<-- Return results -----------|                              |
      |<-- Assign to variable --------|                              |                              |
```

### Execution Example

When the generated code runs:
```go
tables, err := registry.Call("list_tables", nil)
```

This executes:
1. **Interpreter** evaluates the `registry.Call` expression
2. **Registry** looks up the "list_tables" tool
3. **Tool function** executes: `SELECT name FROM sqlite_master WHERE type='table'`
4. **SQLite** returns the results
5. **Registry** formats and returns: `{"count": 4, "tables": ["customers", "products", "orders", "order_items"]}`
6. **Interpreter** assigns result to `tables` variable
7. **Generated code** continues to next statement

---

## Why CodeMode Wins

### 1. Token Efficiency (90% Reduction)

**Tool Calling Problem**: Context accumulation
```
Call 1:  prompt + tools = 1,140 tokens
Call 2:  + call1 result = 1,237 tokens
Call 3:  + call2 result = 2,138 tokens
...
Call 13: + all results  = 4,932 tokens
Total input tokens: 41,407
```

**CodeMode Solution**: Single comprehensive prompt
```
Call 1: prompt + tool docs = 699 tokens
Total input tokens: 699
```

### 2. Latency Reduction (1.35x Faster)

Each API call adds:
- Network round-trip latency (~200-500ms)
- API processing time
- Token generation time

**Tool Calling**: 13 calls × ~3.5s = 50.3s
**CodeMode**: 1 call × 37.1s = 37.1s

### 3. Cost Savings (67% Cheaper)

Using Claude Sonnet pricing ($3/MTok input, $15/MTok output):

**Tool Calling**:
- Input: 41,407 × $0.003/1000 = $0.1242
- Output: 2,595 × $0.015/1000 = $0.0389
- **Total: $0.163**

**CodeMode**:
- Input: 699 × $0.003/1000 = $0.0021
- Output: 3,488 × $0.015/1000 = $0.0523
- **Total: $0.054**

---

## The Loop Advantage

CodeMode's advantage grows with iteration-heavy tasks.

### Tool Calling (updating 10 customers)
```
For each customer:
  - API call to read customer data
  - API call to read order history
  - API call to calculate tier
  - API call to update customer

Total: 40+ API calls, massive token accumulation
```

### CodeMode (updating 10 customers)
```go
// Single API call generates:
customers, _ := registry.Call("read_records", map[string]interface{}{
    "table": "customers",
})

for _, customer := range customers.([]map[string]interface{}) {
    orders, _ := registry.Call("read_records", map[string]interface{}{
        "table": "orders",
        "conditions": map[string]interface{}{
            "customer_id": customer["id"],
        },
    })

    tier := calculateTier(orders)

    registry.Call("update_records", map[string]interface{}{
        "table": "customers",
        "data": map[string]interface{}{"tier": tier},
        "conditions": map[string]interface{}{"id": customer["id"]},
    })
}

// Total: 1 API call, loop executes locally
```

---

## Conclusion

| Advantage | Simple Tasks | Complex Tasks |
|-----------|--------------|---------------|
| Speed | 1.9x faster | 1.35x faster |
| Token Usage | 86% reduction | 90% reduction |
| Cost Savings | 65% cheaper | 67% cheaper |
| API Calls | 6x fewer | 13x fewer |

**When to use CodeMode**:
- Multi-step database operations
- Tasks requiring iteration over data
- Complex conditional logic
- Cost-sensitive applications
- When you need full auditability

**Source**: [jparkerweb/mcp-sqlite](https://github.com/jparkerweb/mcp-sqlite)

---

## Running the Benchmark

```bash
cd sqlite-mcp-benchmark
export ANTHROPIC_API_KEY=your-key
./sqlite-benchmark
```

Or build manually:
```bash
go build -o sqlite-benchmark ./simple-benchmark.go
./sqlite-benchmark
```
