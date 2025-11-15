#!/bin/bash

# Complete Benchmark Suite Runner
# Runs all three approaches and generates comparison

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 E-Commerce Order Processing Benchmark Suite"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check for API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "❌ Error: ANTHROPIC_API_KEY environment variable not set"
    echo "💡 Set it with: export ANTHROPIC_API_KEY=your-key-here"
    exit 1
fi

echo "✅ API key found"
echo ""

# Build all benchmarks
echo "🔨 Building benchmarks..."
go build -o codemode-benchmark codemode-benchmark.go
go build -o toolcalling-benchmark toolcalling-benchmark.go
go build -o mcp-benchmark mcp-benchmark.go
go build -o mcp-server mcp-server.go
echo "✅ Build complete"
echo ""

# Clean previous results
rm -f results-*.json

# Run benchmarks
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  Running Code Mode Benchmark"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
./codemode-benchmark
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  Running Tool Calling Benchmark"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
./toolcalling-benchmark
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  Running Native MCP Benchmark"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Starting MCP server..."
./mcp-server &
MCP_PID=$!
sleep 2  # Wait for server to start

./mcp-benchmark

# Stop MCP server
kill $MCP_PID 2>/dev/null || true
echo ""

# Generate comparison
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 COMPARISON RESULTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Parse and display results
if command -v jq &> /dev/null; then
    # If jq is available, use it for nice formatting
    echo "Approach          | Duration | API Calls | Tokens  | Cost"
    echo "----------------- | -------- | --------- | ------- | --------"

    CODEMODE_DUR=$(jq -r '.duration' results-codemode.json)
    CODEMODE_CALLS=$(jq -r '.apiCalls' results-codemode.json)
    CODEMODE_TOKENS=$(jq -r '.tokens' results-codemode.json)
    CODEMODE_COST=$(jq -r '.cost' results-codemode.json)
    echo "Code Mode         | ${CODEMODE_DUR}s | ${CODEMODE_CALLS}         | ${CODEMODE_TOKENS}   | \$${CODEMODE_COST}"

    TC_DUR=$(jq -r '.duration' results-toolcalling.json)
    TC_CALLS=$(jq -r '.apiCalls' results-toolcalling.json)
    TC_TOKENS=$(jq -r '.tokens' results-toolcalling.json)
    TC_COST=$(jq -r '.cost' results-toolcalling.json)
    echo "Tool Calling      | ${TC_DUR}s | ${TC_CALLS}         | ${TC_TOKENS}  | \$${TC_COST}"

    MCP_DUR=$(jq -r '.duration' results-mcp.json)
    MCP_CALLS=$(jq -r '.totalCalls' results-mcp.json)
    MCP_TOKENS=$(jq -r '.tokens' results-mcp.json)
    MCP_COST=$(jq -r '.cost' results-mcp.json)
    echo "Native MCP        | ${MCP_DUR}s | ${MCP_CALLS}        | ${MCP_TOKENS}   | \$${MCP_COST}"

    echo ""
    echo "📈 Performance vs Code Mode:"
    echo "  Tool Calling: $(echo "scale=1; ($TC_DUR / $CODEMODE_DUR - 1) * 100" | bc)% slower, $(echo "scale=1; ($TC_COST / $CODEMODE_COST - 1) * 100" | bc)% more expensive"
    echo "  Native MCP:   $(echo "scale=1; ($MCP_DUR / $CODEMODE_DUR - 1) * 100" | bc)% slower, $(echo "scale=1; ($MCP_COST / $CODEMODE_COST - 1) * 100" | bc)% more expensive"
else
    # Fallback to simple output
    echo "Code Mode results:"
    cat results-codemode.json
    echo ""
    echo "Tool Calling results:"
    cat results-toolcalling.json
    echo ""
    echo "Native MCP results:"
    cat results-mcp.json
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Benchmark complete! Results saved to results-*.json"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
