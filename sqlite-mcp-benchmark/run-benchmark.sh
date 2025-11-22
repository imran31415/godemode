#!/bin/bash

# SQLite MCP Benchmark: CodeMode vs Tool Calling
# Based on jparkerweb/mcp-sqlite

set -e

# Check for API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "Error: ANTHROPIC_API_KEY environment variable not set"
    echo ""
    echo "Usage:"
    echo "  export ANTHROPIC_API_KEY=your-api-key"
    echo "  ./run-benchmark.sh"
    exit 1
fi

# Build if needed
if [ ! -f sqlite-benchmark ]; then
    echo "Building benchmark..."
    go build -o sqlite-benchmark ./simple-benchmark.go
fi

# Run the benchmark
echo "Starting SQLite MCP Benchmark..."
echo ""
./sqlite-benchmark
