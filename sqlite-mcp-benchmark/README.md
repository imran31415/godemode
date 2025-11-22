# SQLite MCP Benchmark: CodeMode vs Tool Calling

This benchmark compares **CodeMode** (having Claude generate complete Go code) vs **Native Tool Calling** (sequential API calls with tool use) using the SQLite MCP server tools from [jparkerweb/mcp-sqlite](https://github.com/jparkerweb/mcp-sqlite).

## What This Proves

The benchmark tests whether CodeMode allows the LLM to achieve **greater complexity tasks** with the same prompt/inputs by:

1. Reducing API call overhead
2. Enabling local loop execution
3. Allowing complex conditional logic

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

## Expected Results

For the simple scenario (5 tool operations):

| Approach | API Calls | Expected Tokens | Expected Duration |
|----------|-----------|-----------------|-------------------|
| CodeMode | 1 | ~1,500-2,500 | ~3-5 seconds |
| Tool Calling | 3-5 | ~3,000-6,000 | ~10-20 seconds |

**Expected Improvements:**
- 3-5x faster (fewer API roundtrips)
- 40-60% fewer tokens
- 40-60% cost savings

## Scenario Details

**Simple CRUD Operations:**
1. List all tables in the database
2. Get the schema of the customers table
3. Read all customers with status 'active'
4. Create a new customer
5. Read active customers again to confirm

## Architecture

```
sqlite-mcp-benchmark/
├── simple-benchmark.go     # Main benchmark code
├── run-benchmark.sh        # Run script
├── sqlite-mcp-spec.json    # MCP specification (jparkerweb/mcp-sqlite)
├── generated/              # Generated from spec
│   ├── registry.go         # Tool registry
│   └── tools.go            # SQLite tool implementations
└── README.md               # This file
```

## Adding More Complex Scenarios

To test the **loop advantage** of CodeMode, you can modify the prompt to include tasks like:

```
"For each customer in the database, check their order history
and update their tier based on total spending"
```

This type of task requires N iterations, which means:
- **Tool Calling**: N × (API call per iteration) = expensive
- **CodeMode**: 1 API call + local loop execution = efficient

## Sources

- [jparkerweb/mcp-sqlite](https://github.com/jparkerweb/mcp-sqlite) - Original MCP server
- [PulseMCP SQLite](https://www.pulsemcp.com/servers/modelcontextprotocol-sqlite) - MCP registry
