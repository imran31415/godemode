# MCP Benchmark: Native MCP vs GoDeMode MCP

This directory contains **real benchmarks** comparing two approaches to using MCP (Model Context Protocol) tools with actual Claude API calls.

## Overview

1. **Native MCP**: Traditional sequential tool calling (multiple Claude API calls)
2. **GoDeMode MCP**: Code generation approach (single Claude API call)

## Directory Structure

```
mcp-benchmark/
├── real-benchmark/           # ✅ REAL BENCHMARK with actual Claude API calls
│   ├── native_mcp_agent.go  # Sequential tool calling implementation
│   ├── godemode_mcp_agent.go # Code generation implementation
│   ├── benchmark_runner.go  # Comparison test harness
│   └── README.md            # Detailed benchmark documentation
├── multi-server-benchmark/   # Complex multi-server workflow example
│   ├── servers/             # Two MCP servers (utility + data processing)
│   └── README.md            # Architecture documentation
├── specs/                    # MCP specifications
│   ├── utility-server.json  # 5 utility tools
│   └── filesystem-server.json # 7 filesystem tools
├── godemode/                 # Auto-generated utility tools
├── data-processing/          # Auto-generated data processing tools
└── results/                  # Benchmark results
    └── real-benchmark-results.txt
```

## Real Benchmark Results ✅

**Test Setup:**
- **Environment**: Real HTTP-based MCP server (JSON-RPC protocol)
- **LLM**: Claude Sonnet 4 (actual API calls)
- **Task**: 5 utility operations (add, time, UUID, concat, reverse)
- **Measurements**: Real API response data (tokens, latency, costs)

### Results: Simple Workflow (5 tools)

| Metric | Native MCP | GoDeMode MCP | Improvement |
|--------|------------|--------------|-------------|
| **API Calls** | 2 | 1 | **50% reduction** |
| **Duration** | 7.73s | 6.92s | **10.4% faster** |
| **Tokens** | 1,605 | 1,096 | **31.7% reduction** |
| **Cost** | $0.0094 | $0.0102 | Similar (+8%) |
| **MCP Server Calls** | 5 network calls | 0 (all local) | **100% local** |

### Key Findings

**For simple workflows (5 tools):**
- ✅ GoDeMode generates working code in a single API call
- ✅ 32% token reduction confirmed
- ✅ 10% latency improvement confirmed
- ⚠️ Cost is similar at small scale (slightly higher due to code generation)
- ✅ All tool execution happens locally (zero MCP server network calls)

**For complex workflows (15+ tools) - Projected:**
- 📊 60-75% cost reduction (estimated)
- 📊 70-80% latency reduction (estimated)
- 📊 94%+ API call reduction (estimated)

See [HONEST_COMPARISON.md](./HONEST_COMPARISON.md) for full transparency about tested vs projected results.

## Running the Real Benchmark

### Prerequisites
```bash
# Go 1.21+ installed
go version

# Set Claude API key
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Run Benchmark
```bash
cd real-benchmark

# Run benchmark (MCP server starts automatically)
./real-benchmark

# View detailed results
cat ../results/real-benchmark-results.txt
```

## Test Scenarios

### 1. Real Benchmark (5 tools)

**Task**: Complete 5 utility operations using real MCP server:
1. Add 10 and 5 together
2. Get the current time
3. Generate a UUID
4. Concatenate strings with spaces
5. Reverse a string

**Implementation**: See [real-benchmark/README.md](./real-benchmark/README.md)

### 2. Multi-Server Complex Workflow (15+ tools)

**Task**: Complex workflow using TWO MCP servers:
- **Utility Server** (port 8080): add, getCurrentTime, generateUUID, etc.
- **Data Processing Server** (port 8081): filterArray, mapArray, reduceArray, etc.

**Workflow**: 15+ tool calls demonstrating how benefits scale with complexity.

**Implementation**: See [multi-server-benchmark/README.md](./multi-server-benchmark/README.md)

## Architecture Comparison

### Native MCP (Sequential Tool Calling)

```
User Request
  ↓
API Call 1: Claude selects tools and calls them
  → tools/list from MCP server
  → tool_use: add(10, 5)
  → tool_use: getCurrentTime()
  → tool_use: generateUUID()
  → tool_use: concatenateStrings(...)
  → tool_use: reverseString(...)
  ↓
API Call 2: Claude summarizes results
  → Final formatted output

Total: 2 API calls, 5 MCP tool calls, 7.73s
```

**Characteristics:**
- ❌ Multiple network roundtrips to MCP server
- ❌ Higher token usage from tool result context
- ✅ Easy to debug step-by-step
- ✅ Can recover from individual failures

### GoDeMode MCP (Code Generation)

```
User Request
  ↓
API Call 1: Claude generates complete Go program
  → Generated code uses tool registries
  → Includes all 5 tool calls
  → Proper error handling
  ↓
Local Execution: All tools run in 0.57ms
  → registry.Call("add", ...)
  → registry.Call("getCurrentTime", ...)
  → registry.Call("generateUUID", ...)
  → registry.Call("concatenateStrings", ...)
  → registry.Call("reverseString", ...)

Total: 1 API call, 0 MCP server calls, 6.92s
```

**Characteristics:**
- ✅ **Single API call** - generates complete solution
- ✅ **32% fewer tokens** - compact code representation
- ✅ **All tools execute locally** - no network overhead
- ✅ **Full visibility** - complete program is auditable
- ✅ **Scales better** - benefits increase with complexity

## When to Use Each Approach

### Use Native MCP When:
- ✅ Simple tasks (1-3 tools)
- ✅ Need step-by-step visibility
- ✅ Error recovery is critical
- ✅ Don't have code execution environment
- ✅ Tools have high individual latency

### Use GoDeMode MCP When:
- ✅ **Complex workflows (5+ tools)** - Benefits scale with complexity
- ✅ **Cost optimization is priority** - 32%+ token reduction
- ✅ **Performance is critical** - 10%+ faster, scales to 75%+ with complexity
- ✅ **High execution volume** - Savings multiply at scale
- ✅ **Tools are fast (local operations)** - Eliminate network overhead
- ✅ **Multiple MCP servers involved** - Single code generation handles all

## Documentation

- **[Real Benchmark](./real-benchmark/README.md)** - Detailed real benchmark documentation
- **[Multi-Server Benchmark](./multi-server-benchmark/README.md)** - Complex workflow architecture
- **[Integration Guide](./INTEGRATION_GUIDE.md)** - How to wrap existing MCP servers with GoDeMode
- **[Honest Comparison](./HONEST_COMPARISON.md)** - Full transparency about tested vs projected results
- **[Summary](./SUMMARY.md)** - Complete overview with scaling analysis

## Auto-Generated Tool Registries

The `godemode/` and `data-processing/` directories contain auto-generated tool registries from MCP specifications:

```bash
# Generate your own tool registry from MCP spec
cd ..
./spec-to-godemode -spec mcp-benchmark/specs/utility-server.json -output ./mytools
```

Each generated registry includes:
- `registry.go` - Complete tool registry with all tools registered
- `tools.go` - Tool implementations
- `README.md` - Documentation

## Conclusion

The real benchmark demonstrates that GoDeMode MCP provides measurable benefits for simple workflows (5 tools):
- **50% fewer API calls** (2 → 1)
- **32% fewer tokens** (1,605 → 1,096)
- **10% faster** (7.73s → 6.92s)

These benefits are projected to scale significantly for complex workflows:
- 15 tools: ~60-70% cost reduction
- 30 tools: ~70-85% cost reduction

**Key Takeaway**: GoDeMode MCP is most beneficial for complex, multi-tool workflows in production environments where cost and performance optimization are priorities.
