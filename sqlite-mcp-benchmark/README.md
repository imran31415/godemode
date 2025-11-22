# SQLite MCP Benchmark: CodeMode vs Tool Calling

This benchmark compares **CodeMode** (having Claude generate complete Go code) vs **Native Tool Calling** (sequential API calls with tool use) using the SQLite MCP server tools from [jparkerweb/mcp-sqlite](https://github.com/jparkerweb/mcp-sqlite).

## What This Proves

The benchmark tests whether CodeMode allows the LLM to achieve **greater complexity tasks** with the same prompt/inputs by:

1. Reducing API call overhead
2. Enabling local loop execution
3. Allowing complex conditional logic
4. Providing full auditability of operations

## Latest Results

### Simple Scenario (5 tool calls)

| Metric | CodeMode | Tool Calling | Improvement |
|--------|----------|--------------|-------------|
| **Duration** | 8.9s | 17.1s | **1.92x faster** |
| **API Calls** | 1 | 6 | 6x fewer |
| **Tokens** | 1,345 | 9,416 | **85.7% fewer** |
| **Cost** | $0.013 | $0.037 | **64.8% cheaper** |

### Complex Scenario (19 tool calls)

| Metric | CodeMode | Tool Calling | Improvement |
|--------|----------|--------------|-------------|
| **Duration** | 37.1s | 50.3s | **1.35x faster** |
| **API Calls** | 1 | 13 | 13x fewer |
| **Tokens** | 4,187 | 44,002 | **90.5% fewer** |
| **Cost** | $0.054 | $0.163 | **66.6% cheaper** |

## Tools Available

Based on `jparkerweb/mcp-sqlite`, this benchmark uses 8 SQLite tools:

- `db_info` - Get database information
- `list_tables` - List all tables
- `get_table_schema` - Get table schema
- `create_record` - Insert a record
- `read_records` - Query records with filters
- `update_records` - Update matching records
- `delete_records` - Delete matching records
- `query` - Execute custom SQL

## Running the Benchmark

```bash
# Set your API key
export ANTHROPIC_API_KEY=your-api-key

# Run the benchmark
./run-benchmark.sh

# Or manually:
go build -o sqlite-benchmark ./simple-benchmark.go
./sqlite-benchmark
```

## Scenarios

### Simple CRUD Operations
1. List all tables in the database
2. Get the schema of the customers table
3. Read all customers with status 'active'
4. Create a new customer
5. Read active customers again to confirm

### Complex Multi-Table Analysis
1. List all tables to understand the schema
2. Get the schema of all 4 tables (customers, products, orders, order_items)
3. Find all customers who have placed orders (JOIN query)
4. Calculate total revenue per customer (multi-table aggregation)
5. Find the top-selling product by quantity
6. Create a new order with multiple line items
7. Update stock quantities for ordered products
8. Verify the new order was created correctly
9. Generate a summary report

## Audit Logging

The benchmark now includes detailed audit trails for full transparency:

```
--- Audit Log ---
[14:30:01.123] 1. API_CALL: API call #1
[14:30:02.456] 2. API_RESPONSE: Response #1 (stop_reason: tool_use) (tokens: in=1234, out=567)
[14:30:02.457] 3. TOOL_CALL: Tool call #1
                Tool: list_tables
                Args: {}
[14:30:02.458] 4. TOOL_RESULT: list_tables
                Result: ["customers","products","orders","order_items"]
...
```

## Architecture

```
sqlite-mcp-benchmark/
├── simple-benchmark.go     # Main benchmark code with audit logging
├── run-benchmark.sh        # Run script
├── sqlite-mcp-spec.json    # MCP specification (jparkerweb/mcp-sqlite)
├── generated/              # Generated from spec
│   ├── registry.go         # Tool registry
│   └── tools.go            # SQLite tool implementations
├── README.md               # This file
└── BENCHMARK_RESULTS.md    # Detailed results and analysis
```

## Why CodeMode Wins

### Token Efficiency
**Tool Calling**: Context accumulates with each call
```
Call 1: prompt + tools = 1,200 tokens
Call 2: prompt + tools + call1 result = 1,400 tokens
...
Total: 44,000+ input tokens for complex scenarios
```

**CodeMode**: Single comprehensive prompt
```
Call 1: prompt + tool docs = 699 tokens
Total: 699 input tokens
```

### The Loop Advantage

For the complex scenario with joins and aggregations:
- **Tool Calling**: 13 API calls with growing context = expensive
- **CodeMode**: 1 API call + local execution of 19+ operations = efficient

## Sources

- [jparkerweb/mcp-sqlite](https://github.com/jparkerweb/mcp-sqlite) - Original MCP server
- [PulseMCP SQLite](https://www.pulsemcp.com/servers/modelcontextprotocol-sqlite) - MCP registry
