# Integration Guide: Using MCP Servers with GoDeMode

This guide shows you how to wrap existing MCP servers with GoDeMode, enabling you to use Code Mode with any MCP-compatible tool collection.

## Table of Contents
1. [Quick Start](#quick-start)
2. [Step-by-Step Guide](#step-by-step-guide)
3. [Real-World Example](#real-world-example)
4. [Advanced: Connecting to Running MCP Servers](#advanced-connecting-to-running-mcp-servers)
5. [Troubleshooting](#troubleshooting)

## Quick Start

**Goal**: Convert an MCP specification into a GoDeMode tool registry that can be used in Code Mode.

**5-Minute Setup:**
```bash
# 1. Get your MCP specification (JSON file)
# 2. Generate GoDeMode tools
./spec-to-godemode -spec your-mcp-spec.json -output ./mytools

# 3. Implement the tool functions in mytools/tools.go
# 4. Use in your GoDeMode application
```

## Step-by-Step Guide

### Step 1: Obtain Your MCP Specification

You need the JSON specification file from your MCP server. This typically includes:
- Tool names and descriptions
- Parameter definitions (types, required fields)
- Expected return types

**Example MCP Spec** (`email-server.json`):
```json
{
  "name": "email-server",
  "version": "1.0.0",
  "description": "MCP server for email operations",
  "tools": [
    {
      "name": "sendEmail",
      "description": "Send an email to a recipient",
      "inputSchema": {
        "type": "object",
        "properties": {
          "to": {
            "type": "string",
            "description": "Recipient email address"
          },
          "subject": {
            "type": "string",
            "description": "Email subject"
          },
          "body": {
            "type": "string",
            "description": "Email body content"
          }
        },
        "required": ["to", "subject", "body"]
      }
    },
    {
      "name": "readEmail",
      "description": "Read an email by ID",
      "inputSchema": {
        "type": "object",
        "properties": {
          "emailId": {
            "type": "string",
            "description": "Email ID to read"
          }
        },
        "required": ["emailId"]
      }
    }
  ]
}
```

### Step 2: Generate GoDeMode Tool Registry

```bash
cd /path/to/godemode

# Build the spec-to-godemode tool (if not already built)
go build -o spec-to-godemode ./cmd/spec-to-godemode/main.go

# Generate tools from your MCP spec
./spec-to-godemode -spec email-server.json -output ./emailtools -package emailtools
```

**Output:**
```
Detected spec format: mcp
Parsed 2 tools from MCP spec 'email-server'
Generating registry.go...
Generating tools.go...
Generating README.md...

Generated files:
  - ./emailtools/registry.go
  - ./emailtools/tools.go
  - ./emailtools/README.md
✓ Successfully generated GoDeMode code in ./emailtools
```

### Step 3: Implement Tool Functions

The generator creates **stub implementations**. You need to replace them with real logic.

**Generated stub** (`emailtools/tools.go`):
```go
func sendEmail(args map[string]interface{}) (interface{}, error) {
    // Required parameter: to (string)
    to, ok := args["to"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'to' not found or wrong type")
    }
    _ = to // TODO: Use this parameter

    // Required parameter: subject (string)
    subject, ok := args["subject"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'subject' not found or wrong type")
    }
    _ = subject // TODO: Use this parameter

    // TODO: Implement your business logic here
    return map[string]interface{}{
        "status":  "success",
        "message": "sendEmail executed",
    }, nil
}
```

**Your implementation** (replace the TODO section):
```go
func sendEmail(args map[string]interface{}) (interface{}, error) {
    // Parameter extraction (already generated)
    to, ok := args["to"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'to' not found or wrong type")
    }

    subject, ok := args["subject"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'subject' not found or wrong type")
    }

    body, ok := args["body"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'body' not found or wrong type")
    }

    // YOUR IMPLEMENTATION HERE
    // Option 1: Call your existing MCP server
    err := callMCPServer("email-server", "sendEmail", args)
    if err != nil {
        return nil, fmt.Errorf("failed to send email: %w", err)
    }

    // Option 2: Implement directly
    // err := smtp.SendMail(smtpServer, auth,
    //     fromEmail, []string{to}, []byte(formatEmail(subject, body)))

    return map[string]interface{}{
        "status":  "sent",
        "emailId": generateEmailID(),
        "to":      to,
        "subject": subject,
    }, nil
}
```

### Step 4: Use in GoDeMode Application

**Option A: Direct Usage (Manual Code)**
```go
package main

import (
    "fmt"
    "emailtools"
)

func main() {
    registry := emailtools.NewRegistry()

    // Call tools directly
    result, err := registry.Call("sendEmail", map[string]interface{}{
        "to":      "user@example.com",
        "subject": "Hello from GoDeMode",
        "body":    "This email was sent using GoDeMode!",
    })

    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Result: %+v\n", result)
}
```

**Option B: LLM-Generated Code (True Code Mode)**
```go
package main

import (
    "context"
    "fmt"
    "time"

    "emailtools"
    "github.com/imran31415/godemode/pkg/executor"
    "github.com/imran31415/godemode/benchmark/llm"
)

func main() {
    // 1. Create Claude client
    claude := llm.NewClaudeClient(os.Getenv("ANTHROPIC_API_KEY"))

    // 2. Describe tools to Claude
    registry := emailtools.NewRegistry()
    toolDescriptions := registry.GetToolDescriptions()

    // 3. Ask Claude to generate code
    prompt := fmt.Sprintf(`Generate Go code that:
1. Sends a welcome email to new@example.com
2. Sends a follow-up email
3. Returns summary of emails sent

Available tools:
%s

Generate complete Go code using the registry.`, toolDescriptions)

    response, err := claude.GenerateCode(prompt)
    if err != nil {
        panic(err)
    }

    generatedCode := response.Content[0].Text

    // 4. Execute the generated code safely
    exec := executor.NewInterpreterExecutor()
    ctx := context.Background()

    result, err := exec.Execute(ctx, generatedCode, 30*time.Second)
    if err != nil {
        fmt.Printf("Execution error: %v\n", err)
        return
    }

    fmt.Printf("Output:\n%s\n", result.Output)
    fmt.Printf("Duration: %v\n", result.Duration)
}
```

## Real-World Example

Let's wrap the filesystem MCP server we created:

### 1. Start with MCP Spec

Already exists: `specs/filesystem-server.json`

### 2. Generate Tools

```bash
./spec-to-godemode \
    -spec mcp-benchmark/specs/filesystem-server.json \
    -output ./my-fs-tools \
    -package myfs
```

### 3. Implement Real Filesystem Operations

```go
// my-fs-tools/tools.go
package myfs

import (
    "fmt"
    "os"
)

func readFile(args map[string]interface{}) (interface{}, error) {
    path, ok := args["path"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'path' not found")
    }

    // Real implementation
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    return map[string]interface{}{
        "content": string(content),
        "size":    len(content),
    }, nil
}

func writeFile(args map[string]interface{}) (interface{}, error) {
    path, ok := args["path"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'path' not found")
    }

    content, ok := args["content"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'content' not found")
    }

    // Real implementation
    err := os.WriteFile(path, []byte(content), 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to write file: %w", err)
    }

    return map[string]interface{}{
        "status": "success",
        "path":   path,
        "bytes":  len(content),
    }, nil
}

// ... implement other tools
```

### 4. Use with Code Mode

```go
package main

import (
    "context"
    "fmt"
    "myfs"
    "github.com/imran31415/godemode/pkg/executor"
)

func main() {
    // This code would be generated by Claude:
    sourceCode := `package main

import (
    "fmt"
    "myfs"
)

func main() {
    registry := myfs.NewRegistry()

    // Create backup directory
    registry.Call("createDirectory", map[string]interface{}{
        "path": "/tmp/backup",
    })

    // Copy important files
    files := []string{"config.json", "data.db", "settings.yaml"}
    for _, file := range files {
        content, _ := registry.Call("readFile", map[string]interface{}{
            "path": "/app/" + file,
        })

        registry.Call("writeFile", map[string]interface{}{
            "path":    "/tmp/backup/" + file,
            "content": content["content"],
        })
    }

    // List backup
    result, _ := registry.Call("listDirectory", map[string]interface{}{
        "path": "/tmp/backup",
    })

    fmt.Printf("Backup complete: %v\n", result)
}
`

    // Execute with GoDeMode
    exec := executor.NewInterpreterExecutor()
    result, err := exec.Execute(context.Background(), sourceCode, 30*time.Second)

    if err != nil {
        panic(err)
    }

    fmt.Printf("Result: %s\n", result.Output)
}
```

## Advanced: Connecting to Running MCP Servers

If you have an existing MCP server running, you can proxy calls through your GoDeMode tools:

```go
package emailtools

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// MCP server configuration
const mcpServerURL = "http://localhost:3000/mcp"

// Helper function to call MCP server
func callMCPServer(toolName string, args map[string]interface{}) (interface{}, error) {
    payload := map[string]interface{}{
        "jsonrpc": "2.0",
        "id":      1,
        "method":  "tools/call",
        "params": map[string]interface{}{
            "name":      toolName,
            "arguments": args,
        },
    }

    body, _ := json.Marshal(payload)
    resp, err := http.Post(mcpServerURL, "application/json", bytes.NewBuffer(body))
    if err != nil {
        return nil, fmt.Errorf("MCP call failed: %w", err)
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    if result["error"] != nil {
        return nil, fmt.Errorf("MCP error: %v", result["error"])
    }

    return result["result"], nil
}

// Use in your tool implementations:
func sendEmail(args map[string]interface{}) (interface{}, error) {
    // Validate parameters (generated code)
    to, ok := args["to"].(string)
    if !ok {
        return nil, fmt.Errorf("required parameter 'to' not found")
    }
    // ... validate other params

    // Proxy to MCP server
    return callMCPServer("sendEmail", args)
}
```

## Troubleshooting

### Issue: "Generated tools don't compile"

**Cause**: Missing imports or type mismatches

**Solution**: Check the generated `tools.go` file. You may need to add imports:
```go
import (
    "fmt"
    "os"           // Add if using filesystem
    "net/http"     // Add if making HTTP calls
    "encoding/json" // Add if parsing JSON
)
```

### Issue: "Tool execution fails"

**Cause**: Stub implementation not replaced

**Solution**: Replace the `// TODO: Implement your business logic here` section in each tool function.

### Issue: "Can't find MCP spec"

**Cause**: MCP server doesn't expose its specification

**Solutions**:
1. Check server documentation for spec endpoint (often `/mcp/spec` or `/.well-known/mcp.json`)
2. Manually create spec based on server's API documentation
3. Use OpenAPI spec if available: `./spec-to-godemode -spec openapi.json -output ./tools`

### Issue: "Performance not as expected"

**Validate with real benchmarks**: Our real MCP benchmark shows actual performance data. Real performance depends on:
- Actual LLM API latency (measured: ~7-8s for simple workflows)
- Network conditions (varies)
- Tool implementation complexity
- Code generation quality
- Number of tools involved (benefits scale with complexity)

For accurate measurements, run the real benchmark: `cd real-benchmark && ./real-benchmark`

## Next Steps

1. **Test Your Tools**: Write unit tests for each tool function
2. **Add Error Handling**: Implement robust error handling and logging
3. **Optimize**: Profile your tool implementations for performance
4. **Iterate**: Refine based on real-world usage patterns

## Additional Resources

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [GoDeMode Documentation](../README.md)
- [Example Specs](../examples/specs/)
- [Honest Comparison](./HONEST_COMPARISON.md)
