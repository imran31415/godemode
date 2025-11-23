# spec-to-godemode

Convert any MCP or OpenAPI specification into GoDeMode tool registries automatically.

## Overview

`spec-to-godemode` is a CLI tool that transforms MCP (Model Context Protocol) or OpenAPI specifications into ready-to-use GoDeMode tool registries. This enables instant integration of any API or tool collection into your Code Mode workflows.

## How It Works

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   MCP/OpenAPI   │ --> │ spec-to-godemode │ --> │  GoDeMode Code  │
│   Spec File     │     │   CLI Tool       │     │  (registry.go)  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

### Input
- **MCP Specification**: JSON file following the Model Context Protocol format
- **OpenAPI Specification**: JSON/YAML file following OpenAPI 3.x or Swagger 2.0

### Output
The tool generates three files:
1. **`registry.go`** - Complete tool registry with all tools registered
2. **`tools.go`** - Stub implementations for each tool (customize as needed)
3. **`README.md`** - Documentation for the generated tools

## Installation

```bash
# Build from source
cd godemode
go build -o spec-to-godemode ./cmd/spec-to-godemode/main.go

# Or install globally
go install github.com/imran31415/godemode/cmd/spec-to-godemode@latest
```

## Quick Start

### 1. Generate from MCP Specification

```bash
# Basic usage
./spec-to-godemode -spec my-mcp-server.json

# Custom output directory and package name
./spec-to-godemode -spec my-mcp-server.json -output ./mytools -package mytools
```

### 2. Generate from OpenAPI Specification

```bash
# From OpenAPI spec
./spec-to-godemode -spec api-spec.json -output ./api -package api

# From remote URL (download first)
curl -o api-spec.json https://api.example.com/openapi.json
./spec-to-godemode -spec api-spec.json -output ./api
```

### 3. Use Generated Code

```go
package main

import (
    "fmt"
    "mytools"
)

func main() {
    // Create the registry (tools are auto-registered)
    registry := mytools.NewRegistry()

    // Call a tool
    result, err := registry.Call("sendEmail", map[string]interface{}{
        "to": "user@example.com",
        "subject": "Hello",
        "body": "This is a test email",
    })

    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Result: %+v\n", result)
}
```

## CLI Options

```
Usage:
  spec-to-godemode -spec <file> [options]

Required Flags:
  -spec string
        Path to MCP or OpenAPI specification file

Optional Flags:
  -output string
        Output directory for generated code (default: ./generated)
  -package string
        Package name for generated code (default: tools)
  -version
        Show version and exit
  -help
        Show this help message
```

## Supported Specification Formats

### MCP (Model Context Protocol)

MCP specifications define tools that AI models can use:

```json
{
  "name": "email-server",
  "version": "1.0.0",
  "tools": [
    {
      "name": "sendEmail",
      "description": "Send an email to a recipient",
      "inputSchema": {
        "type": "object",
        "properties": {
          "to": {"type": "string", "description": "Recipient email"},
          "subject": {"type": "string", "description": "Email subject"},
          "body": {"type": "string", "description": "Email body"}
        },
        "required": ["to", "subject", "body"]
      }
    }
  ]
}
```

### OpenAPI 3.x

OpenAPI specifications are automatically converted to tool definitions:

```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "User API",
    "version": "1.0.0"
  },
  "paths": {
    "/users/{id}": {
      "get": {
        "operationId": "getUser",
        "summary": "Get a user by ID",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {"type": "string"}
          }
        ]
      }
    }
  }
}
```

## Converting Popular MCP Servers

### Example: SQLite MCP Server

```bash
# 1. Get the MCP specification
curl -o sqlite-mcp.json https://raw.githubusercontent.com/jparkerweb/mcp-sqlite/main/mcp-spec.json

# 2. Generate GoDeMode tools
./spec-to-godemode -spec sqlite-mcp.json -output ./sqlite-tools -package sqlitetools

# 3. Integrate into your Code Mode workflow
```

### Example: Filesystem MCP Server

```bash
# Generate from filesystem MCP spec
./spec-to-godemode -spec mcp-benchmark/specs/filesystem-server.json \
  -output ./fs-tools \
  -package fstools
```

## Generated Code Structure

After running `spec-to-godemode`, you'll get:

```
./mytools/
├── registry.go   # Main registry with all tools registered
├── tools.go      # Tool implementations (customize these)
└── README.md     # Auto-generated documentation
```

### registry.go

```go
package mytools

import "github.com/imran31415/godemode/benchmark/tools"

func NewRegistry() *tools.Registry {
    registry := tools.NewRegistry()

    // Auto-register all tools
    registry.Register(&tools.ToolInfo{
        Name:        "sendEmail",
        Description: "Send an email to a recipient",
        Parameters:  []tools.ParamInfo{...},
        Function:    SendEmail,
    })

    return registry
}
```

### tools.go

```go
package mytools

// SendEmail - Send an email to a recipient
// TODO: Implement this function
func SendEmail(args map[string]interface{}) (interface{}, error) {
    to := args["to"].(string)
    subject := args["subject"].(string)
    body := args["body"].(string)

    // Your implementation here
    return map[string]interface{}{
        "success": true,
        "message": "Email sent",
    }, nil
}
```

## Best Practices

### 1. Customize Tool Implementations

The generated `tools.go` contains stub implementations. Replace these with your actual logic:

```go
// Before (stub)
func SendEmail(args map[string]interface{}) (interface{}, error) {
    return map[string]interface{}{"success": true}, nil
}

// After (real implementation)
func SendEmail(args map[string]interface{}) (interface{}, error) {
    to := args["to"].(string)
    subject := args["subject"].(string)
    body := args["body"].(string)

    // Use your email service
    err := emailService.Send(to, subject, body)
    if err != nil {
        return nil, fmt.Errorf("failed to send email: %w", err)
    }

    return map[string]interface{}{
        "success": true,
        "messageId": generateMessageID(),
    }, nil
}
```

### 2. Add Error Handling

```go
func GetUser(args map[string]interface{}) (interface{}, error) {
    id, ok := args["id"].(string)
    if !ok {
        return nil, fmt.Errorf("id must be a string")
    }

    user, err := database.GetUser(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    return user, nil
}
```

### 3. Multiple Registries

Combine tools from multiple sources:

```go
package main

import (
    "myapp/emailtools"
    "myapp/dbtools"
    "myapp/fstools"
)

func main() {
    // Create combined registry
    registry := tools.NewRegistry()

    // Register tools from different sources
    emailtools.RegisterTools(registry)
    dbtools.RegisterTools(registry)
    fstools.RegisterTools(registry)

    // Now use with Code Mode
}
```

## Integration with Code Mode

Once you have generated tools, integrate them with GoDeMode:

```go
package main

import (
    "context"
    "time"

    "github.com/imran31415/godemode/pkg/executor"
    "mytools"
)

func main() {
    // 1. Create registry with generated tools
    registry := mytools.NewRegistry()

    // 2. Create executor
    exec := executor.NewInterpreterExecutor()

    // 3. LLM generates code that uses your tools
    code := `package main

    func main() {
        result, _ := registry.Call("sendEmail", map[string]interface{}{
            "to": "user@example.com",
            "subject": "Hello",
            "body": "Test message",
        })
        fmt.Println(result)
    }`

    // 4. Execute in sandbox
    ctx := context.Background()
    result, err := exec.Execute(ctx, code, 30*time.Second)

    // Handle result
}
```

## Troubleshooting

### Unknown Spec Format

```
Error: unknown spec format - file must be MCP or OpenAPI specification
```

Ensure your file is valid JSON/YAML and contains the required fields:
- MCP: Must have `name` and `tools` fields
- OpenAPI: Must have `openapi` or `swagger` version field

### No Tools Found

```
Error: no tools found in specification
```

Check that your specification actually defines tools:
- MCP: Check the `tools` array
- OpenAPI: Check that paths have operations defined

### Invalid JSON

```
Error: failed to parse spec file: invalid character...
```

Validate your JSON at https://jsonlint.com/

## Examples

See the `examples/specs/` directory for example specifications:
- `example-mcp.json` - Email server with 3 tools
- `example-openapi.json` - User management API with 4 operations

## Related Documentation

- [Main GoDeMode README](../../README.md) - Complete project overview
- [MCP Benchmark Integration Guide](../../mcp-benchmark/INTEGRATION_GUIDE.md) - MCP integration details
- [MCP Summary](../../mcp-benchmark/SUMMARY.md) - MCP benchmark results

## Contributing

Areas for contribution:
- Support for additional spec formats (GraphQL, gRPC)
- Improved type inference for generated code
- Validation and testing utilities
- IDE integration

## License

MIT - Same as parent GoDeMode project
