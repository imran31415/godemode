package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMCPSpecFromBytes(t *testing.T) {
	validSpec := []byte(`{
		"name": "example-server",
		"version": "1.0.0",
		"description": "An example MCP server",
		"tools": [
			{
				"name": "get_user",
				"description": "Get user information",
				"inputSchema": {
					"type": "object",
					"properties": {
						"user_id": {
							"type": "string",
							"description": "The user ID"
						},
						"include_posts": {
							"type": "boolean",
							"description": "Include user posts"
						}
					},
					"required": ["user_id"]
				}
			},
			{
				"name": "create_post",
				"description": "Create a new post",
				"inputSchema": {
					"type": "object",
					"properties": {
						"title": {
							"type": "string",
							"description": "Post title"
						},
						"content": {
							"type": "string",
							"description": "Post content"
						},
						"tags": {
							"type": "array",
							"description": "Post tags"
						}
					},
					"required": ["title", "content"]
				}
			}
		]
	}`)

	spec, err := ParseMCPSpecFromBytes(validSpec)
	if err != nil {
		t.Fatalf("Failed to parse valid MCP spec: %v", err)
	}

	if spec.Name != "example-server" {
		t.Errorf("Expected name 'example-server', got '%s'", spec.Name)
	}

	if spec.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", spec.Version)
	}

	if len(spec.Tools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(spec.Tools))
	}

	// Check first tool
	if spec.Tools[0].Name != "get_user" {
		t.Errorf("Expected first tool name 'get_user', got '%s'", spec.Tools[0].Name)
	}

	if spec.Tools[0].InputSchema.Properties == nil {
		t.Fatal("Expected tool to have properties")
	}

	if len(spec.Tools[0].InputSchema.Required) != 1 {
		t.Errorf("Expected 1 required field, got %d", len(spec.Tools[0].InputSchema.Required))
	}
}

func TestMCPSpecValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    MCPSpec
		wantErr bool
	}{
		{
			name: "Valid spec",
			spec: MCPSpec{
				Name:    "test-server",
				Version: "1.0.0",
				Tools: []MCPTool{
					{
						Name:        "test_tool",
						Description: "A test tool",
						InputSchema: MCPSchema{Type: "object"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			spec: MCPSpec{
				Version: "1.0.0",
				Tools: []MCPTool{
					{
						Name:        "test_tool",
						Description: "A test tool",
						InputSchema: MCPSchema{Type: "object"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "No tools",
			spec: MCPSpec{
				Name:    "test-server",
				Version: "1.0.0",
				Tools:   []MCPTool{},
			},
			wantErr: true,
		},
		{
			name: "Tool missing name",
			spec: MCPSpec{
				Name:    "test-server",
				Version: "1.0.0",
				Tools: []MCPTool{
					{
						Description: "A test tool",
						InputSchema: MCPSchema{Type: "object"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Tool missing description",
			spec: MCPSpec{
				Name:    "test-server",
				Version: "1.0.0",
				Tools: []MCPTool{
					{
						Name:        "test_tool",
						InputSchema: MCPSchema{Type: "object"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMCPToToolDefinitions(t *testing.T) {
	spec := MCPSpec{
		Name:    "test-server",
		Version: "1.0.0",
		Tools: []MCPTool{
			{
				Name:        "send_email",
				Description: "Send an email",
				InputSchema: MCPSchema{
					Type: "object",
					Properties: map[string]MCPProperty{
						"to": {
							Type:        "string",
							Description: "Recipient email",
						},
						"subject": {
							Type:        "string",
							Description: "Email subject",
						},
						"body": {
							Type:        "string",
							Description: "Email body",
						},
					},
					Required: []string{"to", "subject"},
				},
			},
		},
	}

	tools := spec.ToToolDefinitions()

	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool definition, got %d", len(tools))
	}

	tool := tools[0]
	if tool.Name != "send_email" {
		t.Errorf("Expected name 'send_email', got '%s'", tool.Name)
	}

	if tool.Description != "Send an email" {
		t.Errorf("Expected description 'Send an email', got '%s'", tool.Description)
	}

	if len(tool.Parameters) != 3 {
		t.Fatalf("Expected 3 parameters, got %d", len(tool.Parameters))
	}

	// Check that required fields are marked correctly
	requiredCount := 0
	for _, param := range tool.Parameters {
		if param.Required {
			requiredCount++
		}
	}

	if requiredCount != 2 {
		t.Errorf("Expected 2 required parameters, got %d", requiredCount)
	}
}

func TestMapMCPTypeToGo(t *testing.T) {
	tests := []struct {
		mcpType  string
		expected string
	}{
		{"string", "string"},
		{"number", "float64"},
		{"integer", "int"},
		{"boolean", "bool"},
		{"array", "[]interface{}"},
		{"object", "map[string]interface{}"},
		{"unknown", "interface{}"},
	}

	for _, tt := range tests {
		t.Run(tt.mcpType, func(t *testing.T) {
			result := mapMCPTypeToGo(tt.mcpType)
			if result != tt.expected {
				t.Errorf("mapMCPTypeToGo(%s) = %s, want %s", tt.mcpType, result, tt.expected)
			}
		})
	}
}

func TestParseMCPSpecFromFile(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-mcp.json")

	content := []byte(`{
		"name": "file-test",
		"version": "1.0.0",
		"description": "Test from file",
		"tools": [
			{
				"name": "test_tool",
				"description": "A test tool",
				"inputSchema": {
					"type": "object",
					"properties": {}
				}
			}
		]
	}`)

	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	spec, err := ParseMCPSpec(testFile)
	if err != nil {
		t.Fatalf("Failed to parse MCP spec from file: %v", err)
	}

	if spec.Name != "file-test" {
		t.Errorf("Expected name 'file-test', got '%s'", spec.Name)
	}

	// Test with non-existent file
	_, err = ParseMCPSpec("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}
