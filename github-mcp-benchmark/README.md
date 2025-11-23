# GitHub MCP Benchmark: CodeMode vs Tool Calling

This benchmark compares **CodeMode** (having Claude generate complete Go code) vs **Native Tool Calling** (sequential API calls with tool use) using read-only GitHub API tools.

## What This Proves

The benchmark tests whether CodeMode allows the LLM to achieve **greater complexity tasks** with the same prompt/inputs by:

1. Reducing API call overhead
2. Enabling local loop execution
3. Allowing complex conditional logic
4. Providing full auditability of operations

## Latest Results

### Simple Scenario (2-5 tool calls)

| Metric | CodeMode | Tool Calling | Improvement |
|--------|----------|--------------|-------------|
| **Duration** | 23.1s | 35.6s | **1.54x faster** |
| **API Calls** | 1 | 6 | 6x fewer |
| **Tokens** | 3,015 | 86,840 | **96.5% fewer** |
| **Cost** | $0.034 | $0.275 | **87.6% cheaper** |

### Complex Scenario (13-14 tool calls)

| Metric | CodeMode | Tool Calling | Improvement |
|--------|----------|--------------|-------------|
| **Duration** | 34.6s | 74.3s | **2.14x faster** |
| **API Calls** | 1 | 7 | 7x fewer |
| **Tokens** | 4,024 | 315,405 | **98.7% fewer** |
| **Cost** | $0.048 | $0.979 | **95.1% cheaper** |

## Tools Available (READ-ONLY)

This benchmark uses 12 read-only GitHub API tools for safety:

- `search_repositories` - Search for GitHub repositories
- `get_repository` - Get repository details
- `list_issues` - List and filter repository issues
- `get_issue` - Get the contents of an issue
- `list_issue_comments` - List comments on an issue
- `search_issues` - Search for issues and pull requests
- `list_pull_requests` - List and filter repository pull requests
- `get_pull_request` - Get details of a specific pull request
- `list_commits` - Get commits of a branch in a repository
- `get_file_contents` - Get contents of a file or directory
- `search_code` - Search for code across GitHub repositories
- `list_labels` - List labels for a repository

**Note**: Write operations (create_issue, update_issue, add_issue_comment) have been intentionally removed for safety.

## Running the Benchmark

```bash
# Set your API keys
export ANTHROPIC_API_KEY=your-anthropic-key
export GITHUB_TOKEN=your-github-token

# Run the benchmark
go build -o github-benchmark ./simple-benchmark.go
./github-benchmark
```

## Scenarios

### Simple Repository Analysis
1. Get repository details for cli/cli
2. List recent open issues (top 5)
3. Get details of the first issue
4. List comments on that issue
5. Search for related PRs

### Complex Multi-Repository Analysis
1. Search for repositories with "mcp server"
2. For each repository:
   - Get repository details
   - List open issues
   - List open PRs
3. Get details of first open issue
4. Search for bug-related issues
5. Get recent commits
6. Generate summary statistics

## Execution Architecture

CodeMode uses the Yaegi Go interpreter to execute LLM-generated code. The core implementation is in `pkg/executor/`:

- **`interpreter_executor.go`** - Main execution engine with timeout and output capture
- **`preprocessor.go`** - Handles markdown extraction and code transformation
- **`ExecuteGeneratedCode()`** - High-level API for running LLM-generated code with tool registries

## Why CodeMode Wins

### Token Efficiency

The GitHub API returns very large JSON responses (repository metadata, issue bodies, etc.). With Tool Calling, this context accumulates:

```
Call 1: prompt + tools = 1,700 tokens
Call 2: prompt + tools + repo data = 3,700 tokens
Call 3: prompt + tools + repo + issues = 18,900 tokens
...
Total: 315,000+ input tokens for complex scenarios
```

**CodeMode**: Single comprehensive prompt
```
Call 1: prompt + tool docs = 1,040 tokens
Total: 1,040 input tokens + 2,984 output tokens
```

### The Loop Advantage

For the complex scenario analyzing multiple repositories:
- **Tool Calling**: 7 API calls with growing context = $0.98 per run
- **CodeMode**: 1 API call + local execution of 13 operations = $0.05 per run

This is a **95% cost reduction** for the same analysis.

## Architecture

```
github-mcp-benchmark/
├── simple-benchmark.go     # Main benchmark code with audit logging
├── github-mcp-spec.json    # MCP specification for GitHub tools
├── generated/              # Generated from spec
│   ├── registry.go         # Tool registry
│   └── tools.go            # GitHub API implementations (read-only)
├── .env                    # API keys (not committed)
└── README.md               # This file
```

## Safety Considerations

This benchmark is designed to be safe:

1. **Read-only operations** - No write access to repositories
2. **Rate limiting** - Respects GitHub API rate limits
3. **No credentials in code** - Uses environment variables
4. **Minimal permissions** - Only needs `public_repo` scope for PAT
