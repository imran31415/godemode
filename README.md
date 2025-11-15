# GoDeMode: Code Generation vs Native Tool Calling Benchmark

A comprehensive benchmark comparing **Code Mode** (LLM-generated Go code execution) vs **Native Tool Calling** (direct function invocation) for IT support automation tasks using Claude API.

[![Go 1.21+](https://img.shields.io/badge/go-1.21+-blue)]()

## 🎯 What is This?

This project benchmarks two approaches to building agentic AI systems:

1. **Code Mode**: Claude generates complete Go programs that are interpreted and executed
2. **Native Tool Calling**: Claude makes sequential tool calls using Anthropic's tool use API

Both approaches solve the same tasks using the same underlying tools, allowing direct performance comparison.

## ✨ Features

### Benchmark Framework
- ✅ **3 Complexity Levels**: Simple (3 ops) → Medium (8 ops) → Complex (15 ops)
- ✅ **5 Real Systems**: Email, SQLite, Knowledge Graph, Logs, Configs
- ✅ **21 Production Tools**: Real operations across all systems
- ✅ **Full Verification**: SQL queries, file checks, graph validation
- ✅ **Complete Metrics**: Duration, tokens, API calls, success rates
- ✅ **Side-by-Side Comparison**: Both modes pass all verifications
- ✅ **Claude API Integration**: Uses claude-sonnet-4-20250514

### Code Mode Implementation
- ✅ **yaegi Interpreter**: Fast Go code interpretation without compilation
- ✅ **Source Validation**: Blocks dangerous imports and operations
- ✅ **Execution Timeouts**: Context-based cancellation (30s default)
- ✅ **Parameter Extraction**: Intelligent parsing of generated code for actual tool execution

## 🚀 Quick Start

Get started with GoDeMode in 5 minutes - Benchmark or integrate Code Mode into your application.

### Step 1: Prerequisites

```bash
# Check Go version (1.21+ required)
go version

# Set Claude API key
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Step 2: Clone and Run Benchmark

```bash
# Clone repository
git clone https://github.com/imran31415/godemode.git
cd godemode

# Build and run benchmark
go build -o godemode-benchmark benchmark/cmd/main.go
./godemode-benchmark

# Or run specific complexity
TASK_FILTER=simple ./godemode-benchmark   # 3 operations
TASK_FILTER=medium ./godemode-benchmark   # 8 operations
TASK_FILTER=complex ./godemode-benchmark  # 15 operations
```

**Expected Output:**

```
=== Running Task: email-to-ticket ===

--- Running CODE MODE Agent ---
Generated code solves task in single API call...

--- Running FUNCTION CALLING Agent ---
Step-by-step tool calls...

====================================================================================================
BENCHMARK REPORT
====================================================================================================
1. email-to-ticket (simple, 3 operations)
   CODE MODE:         ✓ All checks passed (11s, 1,448 tokens, 1 API call)
   FUNCTION CALLING:  ✓ All checks passed (13s, 2,764 tokens, 4 API calls)
   COMPARISON: Code Mode 19% faster, used 1,316 fewer tokens, made 3 fewer API calls
```

### Step 3: Integrate Code Mode

Use GoDeMode in your own application for safe LLM code execution:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/imran31415/godemode/pkg/executor"
)

func main() {
    // 1. Create executor with Yaegi interpreter
    exec := executor.NewInterpreterExecutor()

    // 2. Get Go code from your LLM (Claude, GPT, etc.)
    sourceCode := `package main
import "fmt"

func main() {
    fmt.Println("Hello from Code Mode!")
}
`

    // 3. Execute safely with timeout
    ctx := context.Background()
    result, err := exec.Execute(ctx, sourceCode, 30*time.Second)

    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Output: %s\n", result.Output)
    fmt.Printf("Duration: %v\n", result.Duration)
}
```

**What's Happening?**
- **Yaegi Interpreter**: Code is interpreted directly (~15ms) instead of compiled to WASM (2-3s)
- **Source Validation**: Automatically blocks 8 forbidden imports (os/exec, syscall, unsafe, etc.)
- **Execution Timeout**: 30-second timeout prevents infinite loops
- **Pool of 5 Interpreters**: Pre-initialized interpreters enable instant execution

### Step 4: Register Custom Tools

Create a tool registry to give your LLM-generated code access to your systems:

```go
package main

import (
    "github.com/imran31415/godemode/benchmark/tools"
)

func main() {
    // Create tool registry
    registry := tools.NewRegistry()

    // Register custom tools
    registry.Register(&tools.ToolInfo{
        Name:        "sendEmail",
        Description: "Send an email to a recipient",
        Parameters: []tools.ParamInfo{
            {Name: "to", Type: "string", Required: true},
            {Name: "subject", Type: "string", Required: true},
            {Name: "body", Type: "string", Required: true},
        },
        Function: func(args map[string]interface{}) (interface{}, error) {
            // Your email sending logic here
            return "Email sent successfully", nil
        },
    })

    // Now LLM-generated code can call your tools!
}
```

**Available Tool Categories:**
- **Email** (2 tools): `readEmail`, `sendEmail`
- **Database/Tickets** (3 tools): `createTicket`, `updateTicket`, `queryTickets`
- **Knowledge Graph** (2 tools): `findSimilarIssues`, `linkIssueInGraph`
- **Logs/Config** (5 tools): `searchLogs`, `readConfig`, `checkFeatureFlag`, `writeConfig`, `writeLog`
- **Security** (9 tools): `logSecurityEvent`, `searchSecurityEvents`, `analyzeSuspiciousActivity`, and more

See `benchmark/tools/registry.go` for full implementation details.

## 📊 Latest Benchmark Results

All 3 tasks pass verification for both approaches ✅

| Task | Complexity | Code Mode | Function Calling | Advantage |
|------|------------|-----------|------------------|-----------|
| Email to Ticket | Simple (3 ops) | ✅ 11s, 1.4K tokens, 1 call | ✅ 13s, 2.8K tokens, 4 calls | Code Mode |
| Investigate Logs | Medium (8 ops) | ✅ 33s, 3.1K tokens, 1 call | ✅ 28s, 6.7K tokens, 8 calls | Function Calling (speed) / Code Mode (efficiency) |
| Auto-Resolution | Complex (15 ops) | ✅ 36s, 4.0K tokens, 1 call | ✅ 51s, 13.4K tokens, 15 calls | Code Mode |

### Key Insights

**Code Mode Advantages:**
- 📉 **50-70% fewer tokens** - Single LLM call vs iterative approach
- 📉 **75-93% fewer API calls** - 1 call vs 4-15 calls
- 👁️ **Full code visibility** - See complete program logic
- 🧠 **Better planning** - Holistic approach to complex tasks
- 💰 **Lower cost** - Significant token and API call savings

**Function Calling Advantages:**
- ⚡ **Faster on medium tasks** - No interpretation overhead for simple operations
- 🎯 **More predictable** - Exactly expected number of operations
- 🔄 **Easier debugging** - Step-by-step execution visibility
- 💪 **More reliable** - Handles errors gracefully with partial completion

## 🏗️ Architecture

```
godemode/
├── benchmark/
│   ├── agents/                   # CodeMode & FunctionCalling implementations
│   │   ├── codemode_agent.go
│   │   └── function_calling_agent.go
│   ├── systems/                  # Real systems (Email, DB, Graph, Logs, Config)
│   ├── tools/                    # 21 production tool implementations
│   ├── scenarios/                # 3 tasks with setup & verification
│   ├── runner/                   # Benchmark orchestration & reporting
│   ├── llm/                      # Claude API integration
│   └── cmd/main.go              # Main benchmark executable
├── pkg/
│   ├── compiler/                 # Code compilation (cached)
│   ├── validator/                # Safety validation
│   └── executor/                 # yaegi interpreter executor
└── examples/                     # Example programs
```

## 🔧 Integration with Claude API

### Set API Key
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Model Selection
```bash
# Use Sonnet 4 (default, recommended)
./godemode-benchmark

# Or specify model
CLAUDE_MODEL=claude-opus-4-20250514 ./godemode-benchmark
```

## 📝 How It Works

### Code Mode Flow
1. Claude generates complete Go program using task description
2. Code is validated for dangerous operations
3. yaegi interpreter executes the code
4. Tool calls are extracted and executed against real systems
5. Results are verified

### Function Calling Flow
1. Claude creates step-by-step plan
2. For each step, Claude decides which tool to call
3. Tool is executed against real systems
4. Result is fed back to Claude
5. Process repeats until task complete

## 🔒 Security Features

### Blocked by Validator:
- ❌ `os/exec` - Command execution
- ❌ `syscall` - System calls
- ❌ `unsafe` - Unsafe operations
- ❌ `net` - Network access
- ❌ `plugin` - Dynamic loading

### Execution Constraints:
- ⏱️ 30-second timeout per task
- 🔐 Interpreted execution (no system compilation)
- 📁 No direct file system access (only through provided APIs)

## 🧪 Testing

```bash
# Run full benchmark
./godemode-benchmark

# Run specific complexity level
TASK_FILTER=simple ./godemode-benchmark
TASK_FILTER=medium ./godemode-benchmark
TASK_FILTER=complex ./godemode-benchmark

# Run unit tests
go test ./...
```

## 🎯 Use Cases

### When to Use Code Mode
- ✅ Need to minimize API calls and tokens
- ✅ Complex workflows with loops/conditionals
- ✅ Cost optimization is priority
- ✅ Full code audit trail desired

### When to Use Function Calling
- ✅ Need predictable operation counts
- ✅ Real-time responses important
- ✅ Debugging visibility critical
- ✅ Simpler implementation preferred

## 🚧 Current Status

### Completed
- [x] yaegi interpreter-based execution
- [x] Source validation
- [x] 5 real systems with 21 production tools
- [x] 3 benchmark scenarios (simple, medium, complex)
- [x] Full verification for both modes
- [x] Claude API integration
- [x] Both agents passing 100% of tests
- [x] Comprehensive metrics collection

### Future Work
- [ ] Additional benchmark scenarios
- [ ] Performance optimizations
- [ ] Additional LLM provider support
- [ ] Enhanced security validations
- [ ] MCP (Model Context Protocol) integration

## 🤝 Contributing

Areas for contribution:
- Additional benchmark scenarios
- More tool implementations
- Performance optimizations
- Additional LLM providers
- Documentation improvements

## 📄 License

MIT License

## 🙏 Acknowledgments

- [yaegi](https://github.com/traefik/yaegi) - Go interpreter
- [Anthropic Claude](https://www.anthropic.com/) - LLM capabilities
- [SQLite](https://www.sqlite.org/) - Database
- [BadgerDB](https://github.com/dgraph-io/badger) - Knowledge graph storage

---

**Built with ❤️ using Go and Claude API**

*Production-ready benchmark framework for comparing agentic AI approaches*
