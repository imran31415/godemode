#!/bin/bash

# Excel MCP Benchmark Runner

set -e

# Check for API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "Error: ANTHROPIC_API_KEY environment variable not set"
    exit 1
fi

# Build
echo "Building benchmark..."
go build -o excel-benchmark ./simple-benchmark.go

# Run
echo "Running benchmark..."
./excel-benchmark
