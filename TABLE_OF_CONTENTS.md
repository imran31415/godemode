# GoDeMode Project Guide

**Quick Navigation**: A 5-minute overview of the GoDeMode project structure and key components.

---

## 🎯 What is GoDeMode?

GoDeMode is a benchmark comparing two approaches to LLM-powered task execution:
- **Code Mode**: LLM generates complete Go programs executed in a sandbox (63% fewer tokens)
- **Native Tool Calling**: LLM makes sequential function calls (more predictable)

**Key Result**: Both achieve 100% success, but Code Mode is 59% cheaper at scale.

---

## 📚 Documentation Map

### Start Here

**[README.md](./README.md)** - Project overview, installation, quick start
- What is GoDeMode and why it matters
- Installation (Go 1.21+, no TinyGo needed)
- Running your first benchmark: `./godemode-benchmark`
- Key benchmark results at a glance

### Deep Dives

**[RESEARCH.md](./RESEARCH.md)** - Comprehensive technical analysis (12K words)
- Complete implementation walkthrough with code snippets
- Practical examples: Simple task (3 ops) vs Complex task (15 ops)
- Actual LLM-generated code for both approaches
- Cost analysis: $20K/year savings at 1000 tasks/day
- When to use Code Mode vs Native Tool Calling
- 6 future research directions

**[BENCHMARK_REPORT.md](./BENCHMARK_REPORT.md)** - Detailed benchmark results
- All 3 tasks analyzed (Simple, Medium, Complex)
- Token usage, API calls, duration comparisons
- Verification details and metrics
- Lessons learned from implementation

**[BEST_PRACTICES_ANALYSIS.md](./BEST_PRACTICES_ANALYSIS.md)** - Implementation patterns
- Prompt engineering best practices
- Tool calling strategies (deduplication, exact identifiers)
- JSON response cleaning
- Task description guidelines

---

## 🏗️ Architecture Components

### 1. Go Sandbox Execution Environment

**Purpose**: Safely execute LLM-generated Go code

**Implementation**: [`benchmark/agents/codemode_agent.go`](./benchmark/agents/codemode_agent.go)

```
LLM generates Go code → yaegi interpreter → Execute with tools
                         ↓
                    Validation:
                    - No dangerous imports (os/exec, syscall)
                    - Size limits (100KB max)
                    - Execution timeout (30s)
```

**Key Features**:
- **yaegi Go interpreter**: 200x faster than WASM compilation (~15ms vs 2-3s)
- **Tool injection**: All tools available as native Go functions
- **Parameter extraction**: Parse generated code to verify LLM decisions
- **Security**: Sandboxed execution, no system access

**Entry point**: `codemode_agent.go:49-135` (RunTask method)

### 2. Benchmark Framework

**Purpose**: Compare Code Mode vs Native Tool Calling on real tasks

**Structure**:
```
benchmark/
├── agents/
│   ├── codemode_agent.go         # Code generation & execution
│   └── function_calling_agent.go # Sequential tool calling
├── scenarios/
│   └── support_simple.go         # Task definitions (3 tasks)
├── systems/
│   ├── database/   # SQLite ticket storage
│   ├── email/      # RFC 822 email handling  
│   ├── graph/      # BadgerDB knowledge graph
│   ├── filesystem/ # Logs & configs
│   └── security/   # Security event monitoring
└── cmd/
    └── main.go     # Benchmark runner
```

**Running benchmarks**:
```bash
# Build and run
go build -o godemode-benchmark benchmark/cmd/main.go
./godemode-benchmark

# With API key (uses real LLM)
export ANTHROPIC_API_KEY=sk-...
./godemode-benchmark
```

**Scenarios**: [`benchmark/scenarios/support_simple.go`](./benchmark/scenarios/support_simple.go)
- Simple (3 ops): Email → Ticket → Confirmation
- Medium (8 ops): Email → Logs → Graph → Ticket  
- Complex (15 ops): Full workflow with auto-resolution

### 3. Native Tool Calling Implementation

**Purpose**: Traditional sequential LLM function calling

**Implementation**: [`benchmark/agents/function_calling_agent.go`](./benchmark/agents/function_calling_agent.go)

**3-Phase Process**:

```go
// Phase 1: Planning (API Call #1)
plan := planTaskWithLLM(task)
// Returns: ["Read email", "Create ticket", "Send confirmation"]

// Phase 2: Tool Decision (API Call per step)
for step in plan.Steps {
    toolCall := decideToolCallWithLLM(step, task)
    // Returns: {"tool": "readEmail", "parameters": {"emailID": "support_001"}}
    
    // Phase 3: Execute
    result := executeToolCall(toolCall)
}
```

**Key Design Decisions**:
- Full task context sent with EVERY decision
- Deduplication to prevent duplicate tickets
- JSON-only responses for reliable parsing
- Exact identifier matching (e.g., "error_report_001")

**Entry point**: `function_calling_agent.go:49-135` (RunTask method)

---

## 🚀 Using Code Mode in Your Project

### Basic Usage Pattern

```go
import (
    "github.com/imran31415/godemode/benchmark/agents"
    "github.com/imran31415/godemode/benchmark/scenarios"
)

// 1. Set up test environment with your tools
env := scenarios.NewTestEnvironment(
    emailSystem,
    database,
    graph,
    logSystem,
    configSystem,
)

// 2. Create Code Mode agent
agent := agents.NewCodeModeAgent(env)

// 3. Define your task
task := scenarios.Task{
    Name: "email-to-ticket",
    Description: "Read email 'support_001', create ticket, send confirmation",
    ExpectedOps: 3,
}

// 4. Run the task
metrics, err := agent.RunTask(ctx, task, env)

// 5. Check results
fmt.Printf("Success: %v\n", metrics.Success)
fmt.Printf("Tokens used: %d\n", metrics.TokensUsed)
fmt.Printf("API calls: %d\n", metrics.APICallCount)
```

### LLM Prompt Structure

**System Prompt** (from `codemode_agent.go:210-230`):
```
You are a Go code generator that creates complete, executable programs.

You have access to these tools (already imported and available):
- ReadEmail(emailID string) - Read email content
- CreateTicket(customerID, subject, description string, priority int, tags []string)
- SendEmail(to, subject, body string)
- SearchLogs(pattern string)
- FindSimilarIssues(description string, topK int)

CRITICAL: Generate ONLY executable Go code. No explanations, no markdown.
The code will be executed directly.
```

**User Prompt**:
```
Task: email-to-ticket

Description: Read email ID 'support_001' (a bug report about login failure),
create a ticket with high priority (3-4) since it's a bug, and send confirmation

Generate a complete Go program that accomplishes this task using the available tools.
Include proper error handling and return the results.
```

**LLM Response** (executed directly):
```go
package main

import "fmt"

func main() {
    email, err := ReadEmail("support_001")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    ticket, err := CreateTicket(
        "customer123",
        "Login Failure Bug",
        "Bug report from support_001",
        4, // High priority for bugs
        []string{"bug", "login"},
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    err = SendEmail(
        email.From,
        "Ticket Created",
        fmt.Sprintf("Ticket %s created", ticket["ticketID"]),
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
}
```

### Key Implementation Details

**Tool Registration** (from `codemode_agent.go:267-350`):
```go
// Register tools as Go functions in yaegi interpreter
interp.Use(yaegi.Symbols{
    "main": {
        "ReadEmail": reflect.ValueOf(env.EmailSystem.ReadEmail),
        "CreateTicket": reflect.ValueOf(createTicketWrapper),
        "SendEmail": reflect.ValueOf(env.EmailSystem.SendEmail),
        // ... register all your tools
    },
})
```

**Security Validation** (from `codemode_agent.go:252-265`):
```go
// Block dangerous imports
forbiddenImports := []string{
    "os/exec",   // Prevent command execution
    "net/http",  // Prevent network access
    "syscall",   // Prevent system calls
    "unsafe",    // Prevent unsafe operations
}

// Check code size
if len(code) > 100000 {
    return fmt.Errorf("generated code too large")
}
```

---

## 📊 Quick Decision Guide

**Use Code Mode when:**
- Task has 10+ operations
- Cost optimization is priority (63% token savings)
- Audit trail is valuable
- Batch processing acceptable

**Use Native Tool Calling when:**
- Real-time response critical
- Operation count must be exact
- Partial completion acceptable
- Simpler debugging needed

**Crossover point**: ~8-10 operations

---

## 🔗 Additional Resources

- **Frontend Demo**: [`frontend/README.md`](./frontend/README.md) - React Native walkthrough app
- **GitHub Issues**: [Report bugs or request features](https://github.com/imran31415/godemode/issues)
- **Model Used**: claude-sonnet-4-20250514

---

**Last Updated**: November 15, 2025  
**Status**: Production-ready, all tests passing ✅

