# Multi-Server MCP Benchmark

This benchmark demonstrates GoDeMode's advantages with a **complex workflow** using **multiple MCP servers**.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    COMPLEX WORKFLOW                          │
│                                                              │
│  1. Generate 3 UUIDs (Utility Server)                       │
│  2. Get current timestamp (Utility Server)                  │
│  3. Create test data array [1-20] (Data Server - map)       │
│  4. Filter numbers > 10 (Data Server - filter)              │
│  5. Double the filtered numbers (Data Server - map)         │
│  6. Calculate sum (Data Server - reduce)                    │
│  7. Calculate average (Data Server - reduce)                │
│  8. Calculate max (Data Server - reduce)                    │
│  9. Sort descending (Data Server - sort)                    │
│  10. Get unique values [20,18,16,14,12] (Data Server)       │
│  11. Create result ID (Utility Server - UUID)               │
│  12. Format timestamp (Utility Server - concat)             │
│  13. Create summary string (Utility Server - concat)        │
│  14. Reverse for checksum (Utility Server - reverse)        │
│  15. Add final values (Utility Server - add)                │
│                                                              │
│  TOTAL: ~15 tool calls across 2 MCP servers                 │
└─────────────────────────────────────────────────────────────┘
```

## Servers

### 1. Utility MCP Server (Port 8080)
- **Tools**: add, getCurrentTime, generateUUID, concatenateStrings, reverseString
- **Location**: `servers/utility_server.go`

### 2. Data Processing MCP Server (Port 8081)
- **Tools**: filterArray, mapArray, reduceArray, sortArray, mergeArrays, uniqueValues
- **Location**: `servers/data_server.go`

## Expected Results

### Native MCP Approach
```
User sends task → Claude plans workflow
   ↓
Claude calls utility.generateUUID()         [API call 1]
   ↓
Claude calls utility.getCurrentTime()       [API call 2]
   ↓
Claude calls data.mapArray()                [API call 3]
   ↓
Claude calls data.filterArray()             [API call 4]
   ↓
Claude calls data.mapArray()                [API call 5]
   ↓
Claude calls data.reduceArray(sum)          [API call 6]
   ↓
Claude calls data.reduceArray(avg)          [API call 7]
   ↓
Claude calls data.reduceArray(max)          [API call 8]
   ↓
Claude calls data.sortArray()               [API call 9]
   ↓
Claude calls data.uniqueValues()            [API call 10]
   ↓
Claude calls utility.generateUUID()         [API call 11]
   ↓
Claude calls utility.concatenateStrings()   [API call 12]
   ↓
Claude calls utility.concatenateStrings()   [API call 13]
   ↓
Claude calls utility.reverseString()        [API call 14]
   ↓
Claude calls utility.add()                  [API call 15]
   ↓
Claude summarizes results                   [API call 16]

TOTAL: ~16 API calls, ~12-20 seconds, ~5000-6000 tokens
```

### GoDeMode MCP Approach
```
User sends task → Claude generates complete Go code
   ↓
Generated code executes all tools locally:
   - utilityRegistry.Call("generateUUID", ...)
   - utilityRegistry.Call("getCurrentTime", ...)
   - dataRegistry.Call("mapArray", ...)
   - dataRegistry.Call("filterArray", ...)
   - dataRegistry.Call("mapArray", ...)
   - dataRegistry.Call("reduceArray", ...)
   - ... (all 15 tool calls)
   - Final summary generation

All executed in ~10ms

TOTAL: 1 API call, ~3-5 seconds, ~2000-2500 tokens
```

## Projected Improvements

| Metric | Native MCP | GoDeMode MCP | Improvement |
|--------|------------|--------------|-------------|
| API Calls | ~16 | 1 | **94% ↓** |
| Duration | ~12-20s | ~3-5s | **75-85% faster** |
| Tokens | ~5000-6000 | ~2000-2500 | **58-62% ↓** |
| Cost | ~$0.08-0.12 | ~$0.03-0.05 | **62-75% ↓** |

## Running the Benchmark

### Build Servers
```bash
cd servers

# Build utility server
go build -o utility-server utility_server.go

# Build data server
go build -o data-server data_server.go
```

### Run Servers (in separate terminals)
```bash
# Terminal 1 - Utility Server
cd servers
./utility-server

# Terminal 2 - Data Server
cd servers
./data-server
```

### Run Benchmark
```bash
cd ..
export ANTHROPIC_API_KEY=your-key
go build -o complex-benchmark main.go
./complex-benchmark
```

## Code Organization

```
multi-server-benchmark/
├── servers/
│   ├── utility_server.go          # Utility MCP server (port 8080)
│   ├── data_server.go              # Data processing MCP server (port 8081)
│   ├── utility-server              # Compiled binary
│   └── data-server                 # Compiled binary
├── agents/
│   ├── types.go                    # Shared types
│   ├── native_mcp.go               # Native MCP agent (sequential calls)
│   ├── godemode_mcp.go             # GoDeMode agent (code generation)
│   └── mcp_client.go               # MCP client utilities
├── main.go                         # Benchmark runner
├── results/
│   └── complex-benchmark.txt       # Results output
└── README.md                       # This file
```

## Real-World Applications

This complex workflow simulates realistic scenarios:

1. **Data Pipeline**: ETL operations with data from multiple sources
2. **Analytics Dashboard**: Aggregate metrics from different services
3. **Report Generation**: Combine data processing with formatting
4. **Batch Processing**: Process large datasets with multiple transformations

## Key Takeaways

1. **Scalability**: GoDeMode's advantages grow with workflow complexity
2. **Multi-MCP**: Works seamlessly across multiple tool providers
3. **Cost Efficiency**: 60-75% cost reduction for complex workflows
4. **Performance**: 3-4x faster for multi-step operations
5. **Reliability**: Single code execution vs. many network roundtrips
