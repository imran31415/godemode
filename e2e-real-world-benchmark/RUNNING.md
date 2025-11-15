## Running the E2E Benchmarks

This directory contains **executable benchmarks** that perform real API calls to Claude to compare three approaches for processing an e-commerce order with 12 operations.

## Quick Start

### Prerequisites

1. **Go 1.21+** installed
2. **Anthropic API key** - Get one from [console.anthropic.com](https://console.anthropic.com)
3. **Export your API key:**
   ```bash
   export ANTHROPIC_API_KEY=sk-ant-api03-...
   ```

### Run All Benchmarks

The easiest way to run all three approaches and compare results:

```bash
chmod +x run-all.sh
./run-all.sh
```

This will:
1. Build all three benchmark programs
2. Run Code Mode benchmark
3. Run Tool Calling benchmark
4. Start MCP server and run MCP benchmark
5. Generate comparison table

**Expected output:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 COMPARISON RESULTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Approach          | Duration | API Calls | Tokens  | Cost
----------------- | -------- | --------- | ------- | --------
Code Mode         | 9.2s     | 1         | 4,140   | $0.0277
Tool Calling      | 25.1s    | 4         | 10,095  | $0.0495
Native MCP        | 21.9s    | 17        | 7,873   | $0.0356

📈 Performance vs Code Mode:
  Tool Calling: 172.8% slower, 78.8% more expensive
  Native MCP:   138.0% slower, 28.5% more expensive
```

## Run Individual Benchmarks

### 1. Code Mode Benchmark

Tests the GoDeMode pattern where Claude generates complete code in one API call:

```bash
go build -o codemode-benchmark codemode-benchmark.go
./codemode-benchmark
```

**What it does:**
- Makes 1 API call to Claude to generate complete order processing code
- Executes all 12 tool calls locally
- Measures total duration, tokens, and cost

**Expected results:**
- Duration: ~9-12 seconds
- API calls: 1
- Tokens: ~4,000-4,500
- Cost: ~$0.025-0.035

### 2. Tool Calling Benchmark

Tests native Anthropic Messages API with tool_use:

```bash
go build -o toolcalling-benchmark toolcalling-benchmark.go
./toolcalling-benchmark
```

**What it does:**
- Makes multiple sequential API calls to Claude
- Claude decides which tools to call in each iteration
- Each tool result requires a new API call

**Expected results:**
- Duration: ~20-30 seconds
- API calls: 3-5
- Tokens: ~8,000-12,000
- Cost: ~$0.040-0.060

### 3. Native MCP Benchmark

Tests Model Context Protocol with HTTP-based tool server:

**Terminal 1 - Start MCP server:**
```bash
go build -o mcp-server mcp-server.go
./mcp-server
```

**Terminal 2 - Run benchmark:**
```bash
go build -o mcp-benchmark mcp-benchmark.go
./mcp-benchmark
```

**What it does:**
- Connects to MCP server via HTTP/JSON-RPC
- Lists available tools from MCP server
- Claude calls tools through MCP protocol
- Each tool call = 1 Claude API call + 1 HTTP request to MCP server

**Expected results:**
- Duration: ~18-25 seconds
- Claude API calls: 3-5
- MCP calls: 12-15
- Total calls: 15-20
- MCP overhead: ~1-2 seconds
- Tokens: ~7,000-9,000
- Cost: ~$0.030-0.045

## Understanding the Results

### Key Metrics

**Duration:** Total time from start to finish (includes API latency + tool execution)

**API Calls:** Number of calls to Claude's API
- Code Mode: 1 call (generates all code at once)
- Tool Calling: 3-5 calls (sequential tool use)
- MCP: 3-5 Claude calls + 12-15 MCP HTTP calls

**Tokens:** Input + output tokens sent to Claude
- Lower tokens = lower cost and faster processing
- Code is more compact than tool results

**Cost:** Based on Claude Sonnet 4 pricing ($3/1M input, $15/1M output)

### Why Code Mode Wins

1. **Single API call** - No sequential latency between operations
2. **Compact representation** - Code is smaller than verbose tool results
3. **Local execution** - All 12 tools run locally without network calls
4. **Natural control flow** - Loops and conditionals work as expected

### When Each Approach Breaks

**Code Mode:**
- ✅ Scales to any complexity
- ⚠️ Very complex logic may need 2 API calls (still better than alternatives)

**Tool Calling:**
- ✅ Works for 1-5 simple operations
- ⚠️ Struggles at 10+ operations
- ❌ Breaks with loops (each iteration = API call)

**Native MCP:**
- ✅ Works for 5-15 operations
- ⚠️ Network overhead accumulates
- ❌ Same loop problem as Tool Calling

## The Scenario

All three benchmarks process the same e-commerce order:

**Order Details:**
- Customer: CUST-12345 (Gold tier)
- Items: Laptop ($1,299.99), Mouse ($29.99), Keyboard ($89.99)
- Subtotal: $1,419.97
- Discount: 20% with code SAVE20 (-$283.99)
- Shipping: $15.00 to California
- Tax: $91.08 (CA sales tax 9.25%)
- **Total: $1,242.06**

**12-Step Workflow:**
1. Validate customer
2. Check inventory availability
3. Calculate shipping cost
4. Validate discount code
5. Calculate sales tax
6. Process payment authorization
7. Reserve inventory
8. Create shipping label
9. Send order confirmation email
10. Log transaction to analytics
11. Update customer loyalty points
12. Create warehouse fulfillment task

## File Structure

```
e2e-real-world-benchmark/
├── RUNNING.md              ← You are here
├── SCENARIO.md             ← Detailed workflow description
├── RESULTS.md              ← Detailed results analysis
├── FINAL_VERDICT.md        ← Comprehensive comparison
├── tools/
│   └── registry.go         ← Shared tool implementations
├── codemode-benchmark.go   ← Code Mode implementation
├── toolcalling-benchmark.go ← Tool Calling implementation
├── mcp-benchmark.go        ← MCP client implementation
├── mcp-server.go           ← MCP server implementation
├── run-all.sh              ← Run all benchmarks script
└── go.mod                  ← Go module definition
```

## Cost Considerations

**Single run of all 3 benchmarks:**
- Code Mode: ~$0.028
- Tool Calling: ~$0.050
- MCP: ~$0.036
- **Total: ~$0.114 per complete benchmark run**

**At scale (1,000 orders/day for 1 year):**
- Code Mode: $10,110/year
- Tool Calling: $18,068/year (+$7,958)
- MCP: $12,994/year (+$2,884)

## Troubleshooting

### "ANTHROPIC_API_KEY environment variable not set"

Set your API key:
```bash
export ANTHROPIC_API_KEY=sk-ant-api03-...
```

### "Error listing tools: connection refused" (MCP benchmark)

The MCP server isn't running. Start it first:
```bash
./mcp-server  # In a separate terminal
```

### Build errors about imports

Make sure you're in the correct directory and module is initialized:
```bash
cd e2e-real-world-benchmark
go mod init github.com/yourusername/godemode/e2e-real-world-benchmark
go build ./...
```

### API rate limits

If you hit rate limits, add delays between benchmarks:
```bash
./codemode-benchmark
sleep 5
./toolcalling-benchmark
sleep 5
./mcp-benchmark
```

## Next Steps

After running the benchmarks:

1. **Review results** - Check `results-*.json` files for detailed metrics
2. **Read analysis** - See `FINAL_VERDICT.md` for comprehensive comparison
3. **Try advanced scenario** - See `ADVANCED_SCENARIO.md` for complex fraud detection
4. **Customize** - Modify tools or workflow to match your use case

## Contributing

To add new tools or scenarios:

1. Add tool implementation to `tools/registry.go`
2. Update tool definitions in all three benchmarks
3. Modify the order processing workflow
4. Run benchmarks and update documentation

---

**Questions?** See:
- [SCENARIO.md](./SCENARIO.md) - Detailed workflow
- [RESULTS.md](./RESULTS.md) - Expected results
- [FINAL_VERDICT.md](./FINAL_VERDICT.md) - Complete analysis
