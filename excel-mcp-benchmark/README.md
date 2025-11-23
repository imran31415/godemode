# Excel MCP Benchmark: CodeMode vs Tool Calling

This benchmark compares **CodeMode** (having Claude generate complete Go code) vs **Native Tool Calling** (sequential API calls with tool use) using Excel file manipulation tools inspired by [haris-musa/excel-mcp-server](https://github.com/haris-musa/excel-mcp-server).

## What This Proves

The benchmark tests whether CodeMode allows the LLM to achieve **greater complexity tasks** with the same prompt/inputs by:

1. Reducing API call overhead
2. Enabling local loop execution
3. Allowing complex conditional logic
4. Providing full auditability of operations

## Tools Available

This benchmark uses 10 Excel manipulation tools implemented with the [excelize](https://github.com/xuri/excelize) Go library:

- `describe_sheets` - List all sheets and their metadata
- `read_sheet` - Read data from a sheet (optional range)
- `write_to_sheet` - Write data to a sheet
- `create_sheet` - Create a new sheet
- `delete_sheet` - Delete a sheet
- `copy_sheet` - Copy a sheet to a new sheet
- `get_cell` - Get a single cell value
- `set_cell` - Set a single cell value
- `set_formula` - Set a formula in a cell
- `create_workbook` - Create a new Excel workbook

## Running the Benchmark

```bash
# Set your API key
export ANTHROPIC_API_KEY=your-api-key

# Build and run
go build -o excel-benchmark ./simple-benchmark.go
./excel-benchmark
```

## Scenarios

### Simple CRUD Operations (~5 tool calls)
1. List all sheets in the workbook
2. Read the data from the "Sales" sheet
3. Get the value of cell B2
4. Set a formula to calculate total
5. Read back to verify

### Complex Multi-Sheet Analysis (~12 tool calls)
1. List all sheets to understand structure
2. Read Sales sheet data
3. Read Products sheet data
4. Create a new Analysis sheet
5. Write headers to Analysis sheet
6. Calculate and write metrics (Total Sales, Average, Max, Product Count)
7. Read back Analysis sheet to verify
8. Copy Sales sheet to backup
9. Describe all sheets for final structure

## Architecture

```
excel-mcp-benchmark/
├── simple-benchmark.go     # Main benchmark code with audit logging
├── excel-mcp-spec.json     # MCP tool specification
├── generated/              # Generated tool implementations
│   ├── registry.go         # Tool registry
│   └── tools.go            # Excel tool implementations (using excelize)
├── go.mod                  # Go module file
└── README.md               # This file
```

## Why CodeMode Wins

### Token Efficiency
**Tool Calling**: Context accumulates with each call
```
Call 1: prompt + tools = 1,200 tokens
Call 2: prompt + tools + call1 result = 1,400 tokens
...
Total: 30,000+ input tokens for complex scenarios
```

**CodeMode**: Single comprehensive prompt
```
Call 1: prompt + tool docs = 800 tokens
Total: 800 input tokens
```

### The Loop Advantage

For complex scenarios with multiple operations:
- **Tool Calling**: Multiple API calls with growing context = expensive
- **CodeMode**: 1 API call + local execution = efficient

## Sources

- [haris-musa/excel-mcp-server](https://github.com/haris-musa/excel-mcp-server) - Original Python MCP server
- [negokaz/excel-mcp-server](https://github.com/negokaz/excel-mcp-server) - Alternative implementation
- [excelize](https://github.com/xuri/excelize) - Go library for Excel files
- [PulseMCP Excel](https://www.pulsemcp.com/servers/haris-excel) - MCP registry
