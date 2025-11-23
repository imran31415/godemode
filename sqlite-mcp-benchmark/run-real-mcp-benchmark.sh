#!/bin/bash

# Real MCP Benchmark Runner
# This script starts the MCP server and runs the benchmark

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 SQLite Real MCP Benchmark"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check for API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "❌ Error: ANTHROPIC_API_KEY environment variable not set"
    echo "   Run: export ANTHROPIC_API_KEY=your-key"
    exit 1
fi

cd "$(dirname "$0")"

# Clean up any existing database
rm -f mcp-benchmark.db codemode-benchmark.db

# Build MCP server
echo ""
echo "📦 Building MCP server..."
go build -o mcp-server-bin ./mcp-server.go

# Build benchmark
echo "📦 Building benchmark..."
go build -o real-mcp-benchmark-bin ./real-mcp-benchmark.go

# Start MCP server in background
echo ""
echo "🔌 Starting MCP server..."
./mcp-server-bin &
MCP_PID=$!

# Wait for server to start
sleep 2

# Check if server is running
if ! kill -0 $MCP_PID 2>/dev/null; then
    echo "❌ MCP server failed to start"
    exit 1
fi

echo "✅ MCP server running (PID: $MCP_PID)"

# Run benchmark
echo ""
echo "🏃 Running benchmark..."
echo ""
./real-mcp-benchmark-bin

# Cleanup
echo ""
echo "🧹 Cleaning up..."
kill $MCP_PID 2>/dev/null || true
rm -f mcp-benchmark.db codemode-benchmark.db

echo ""
echo "✅ Benchmark complete!"
