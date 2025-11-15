# End-to-End Real-World Benchmark: E-Commerce Order Processing

## Overview

This benchmark demonstrates **three different approaches** to building an AI agent that processes a complete e-commerce order fulfillment workflow using **12 tools** across **4 systems**.

## The Three Approaches

### 1. Code Mode (GoDeMode)
**How it works:**
- Send 1 API call to Claude
- Claude generates a complete Go program
- Program uses tool registry to execute all 12 operations
- All tool calls happen locally without additional API calls

**Expected Performance:**
- **API Calls**: 1
- **Duration**: ~15-20s (1 API call + local tool execution)
- **Tokens**: ~3,000-4,000
- **Cost**: ~$0.05-0.08

**Advantages:**
- Minimal API latency
- Full code visibility
- Can use programming constructs (loops, conditionals, error handling)
- Best for complex workflows

### 2. Native Tool Calling (Anthropic Messages API)
**How it works:**
- Claude makes sequential tool_use calls
- For each tool: API call → tool execution → result → next API call
- Typically batches some tool calls together
- Continues until workflow complete

**Expected Performance:**
- **API Calls**: 3-5 (batching tool_use blocks)
- **Duration**: ~25-35s (multiple API roundtrips)
- **Tokens**: ~6,000-10,000
- **Cost**: ~$0.12-0.18

**Advantages:**
- Standard Anthropic approach
- Easy step-by-step debugging
- Built-in error recovery
- Official SDK support

### 3. Native MCP (Model Context Protocol)
**How it works:**
- MCP server exposes tools via JSON-RPC over HTTP
- Claude calls tools through MCP protocol
- Similar to Native Tool Calling but standardized
- Tools can be provided by any MCP-compatible server

**Expected Performance:**
- **API Calls**: 2-4 (can batch tool_use calls)
- **Duration**: ~20-30s (API calls + MCP server network overhead)
- **Tokens**: ~5,000-8,000
- **Cost**: ~$0.10-0.15

**Advantages:**
- Standardized protocol (MCP spec)
- Tool provider agnostic
- Can connect to any MCP server
- Growing ecosystem

## The Workflow: Complete Order Processing

```
1. validateCustomer     → Verify customer info, get loyalty tier
2. checkInventory       → Confirm all items in stock
3. calculateShipping    → Get shipping cost and delivery date
4. validateDiscount     → Check discount code, calculate savings
5. calculateTax         → Compute sales tax for location
6. processPayment       → Authorize payment
7. reserveInventory     → Lock inventory for order
8. createShippingLabel  → Generate tracking number and label
9. sendOrderConfirmation → Email customer
10. logTransaction      → Record in analytics
11. updateLoyaltyPoints → Award points to customer
12. createFulfillmentTask → Create warehouse picking task
```

## Performance Comparison Table

| Metric | Code Mode | Tool Calling | Native MCP | Winner |
|--------|-----------|--------------|------------|--------|
| **API Calls** | 1 | 3-5 | 2-4 | Code Mode |
| **Duration** | ~15-20s | ~25-35s | ~20-30s | Code Mode |
| **Tokens** | ~3,000-4,000 | ~6,000-10,000 | ~5,000-8,000 | Code Mode |
| **Cost** | ~$0.05-0.08 | ~$0.12-0.18 | ~$0.10-0.15 | Code Mode |
| **Debugging** | Code visible | Step-by-step | Step-by-step | Tool Calling |
| **Error Recovery** | Manual | Automatic | Automatic | Tool Calling/MCP |
| **Standardization** | Custom | Anthropic | MCP Spec | Native MCP |

## When to Use Each Approach

### Use Code Mode When:
- ✅ Complex workflows (10+ operations)
- ✅ Cost optimization is critical
- ✅ Performance matters (high volume)
- ✅ You have code execution capability
- ✅ Full visibility into logic is needed
- ✅ Workflow has conditional logic/loops

### Use Native Tool Calling When:
- ✅ Simple workflows (1-5 operations)
- ✅ Need official Anthropic support
- ✅ Debugging visibility is priority
- ✅ Error recovery is critical
- ✅ Don't want to manage code execution

### Use Native MCP When:
- ✅ Want standardized protocol
- ✅ Using multiple tool providers
- ✅ Need ecosystem compatibility
- ✅ Tool providers already support MCP
- ✅ Want future-proof solution

## Real-World Implications

### Cost at Scale (1,000 orders/day)

| Approach | Cost per Order | Daily Cost | Annual Cost |
|----------|---------------|------------|-------------|
| **Code Mode** | $0.06 | $60 | $21,900 |
| **Tool Calling** | $0.15 | $150 | $54,750 |
| **Native MCP** | $0.12 | $120 | $43,800 |

**Savings:** Code Mode saves **$32,850-65%** annually compared to alternatives.

### Performance at Scale

For a high-traffic e-commerce site processing 1,000 orders/hour:

- **Code Mode**: Can handle load efficiently, minimal API latency
- **Tool Calling**: May hit rate limits, higher latency per order
- **Native MCP**: Balanced, but network overhead accumulates

## Running the Benchmark

```bash
cd e2e-real-world-benchmark

# Ensure API key is set
export ANTHROPIC_API_KEY="your-key"

# Initialize Go module
go mod init e2e-benchmark
go mod tidy

# Run comprehensive benchmark
go run main.go

# View results
cat results.txt
```

## Architecture Diagrams

### Code Mode Flow
```
┌──────────────┐
│ User Request │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────┐
│ API Call 1: Generate Code    │
│ Claude writes complete Go    │
│ program with all 12 steps    │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Local Execution              │
│ registry.Call("validate...")│
│ registry.Call("checkInv...") │
│ ... (all 12 tools)          │
│ Total time: ~1-2 seconds    │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────┐
│ Final Result │
└──────────────┘

Total: 1 API call, 15-20s
```

### Native Tool Calling Flow
```
┌──────────────┐
│ User Request │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────┐
│ API Call 1: Plan + Tools 1-4 │
│ tool_use: validateCustomer   │
│ tool_use: checkInventory     │
│ tool_use: calculateShipping  │
│ tool_use: validateDiscount   │
└──────┬───────────────────────┘
       │ Execute tools, return results
       ▼
┌──────────────────────────────┐
│ API Call 2: Tools 5-8        │
│ tool_use: calculateTax       │
│ tool_use: processPayment     │
│ tool_use: reserveInventory   │
│ tool_use: createShipLabel    │
└──────┬───────────────────────┘
       │ Execute tools, return results
       ▼
┌──────────────────────────────┐
│ API Call 3: Tools 9-12       │
│ tool_use: sendConfirmation   │
│ tool_use: logTransaction     │
│ tool_use: updateLoyalty      │
│ tool_use: createFulfillment  │
└──────┬───────────────────────┘
       │ Execute tools, return results
       ▼
┌──────────────────────────────┐
│ API Call 4: Summary          │
│ Final order confirmation     │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────┐
│ Final Result │
└──────────────┘

Total: 3-5 API calls, 25-35s
```

### Native MCP Flow
```
┌──────────────┐
│ User Request │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────┐
│ API Call 1: tools/list       │
│ Query MCP server for tools   │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ API Call 2: Batch tool_use   │
│ Multiple tools/call requests │
│ to MCP server (JSON-RPC)     │
└──────┬───────────────────────┘
       │ Each tool: HTTP request to MCP
       ▼
┌──────────────────────────────┐
│ API Call 3: Continue + Summary│
│ More tools if needed         │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────┐
│ Final Result │
└──────────────┘

Total: 2-4 API calls, 20-30s
(+ MCP server HTTP overhead)
```

## Conclusion

For this real-world e-commerce scenario with 12 operations:

1. **Code Mode wins on performance and cost** (1 API call, fastest, cheapest)
2. **Tool Calling wins on ease of use** (standard API, best debugging)
3. **Native MCP wins on standardization** (protocol-based, ecosystem-friendly)

**Recommendation:** Use Code Mode for production systems with complex workflows where cost and performance matter. Use Tool Calling for simpler scenarios or when debugging/iteration is priority. Use Native MCP when working with standardized tool ecosystems or multiple providers.
