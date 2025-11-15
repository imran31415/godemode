package spec

import (
	"testing"
)

func TestDetectSpecFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected SpecFormat
	}{
		{
			name: "MCP format",
			input: []byte(`{
				"name": "test-server",
				"version": "1.0.0",
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
			}`),
			expected: FormatMCP,
		},
		{
			name: "OpenAPI 3.x format",
			input: []byte(`{
				"openapi": "3.0.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				},
				"paths": {}
			}`),
			expected: FormatOpenAPI,
		},
		{
			name: "Swagger 2.0 format",
			input: []byte(`{
				"swagger": "2.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				},
				"paths": {}
			}`),
			expected: FormatOpenAPI,
		},
		{
			name: "Unknown format",
			input: []byte(`{
				"random": "data",
				"nothing": "useful"
			}`),
			expected: FormatUnknown,
		},
		{
			name:     "Invalid JSON",
			input:    []byte(`{invalid json`),
			expected: FormatUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectSpecFormat(tt.input)
			if result != tt.expected {
				t.Errorf("DetectSpecFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToolDefinition(t *testing.T) {
	td := ToolDefinition{
		Name:        "testTool",
		Description: "A test tool",
		Parameters: []Parameter{
			{
				Name:        "param1",
				Type:        "string",
				Description: "First parameter",
				Required:    true,
			},
			{
				Name:        "param2",
				Type:        "int",
				Description: "Second parameter",
				Required:    false,
				Default:     42,
			},
		},
	}

	if td.Name != "testTool" {
		t.Errorf("Expected name 'testTool', got '%s'", td.Name)
	}

	if len(td.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(td.Parameters))
	}

	if td.Parameters[0].Required != true {
		t.Errorf("Expected param1 to be required")
	}

	if td.Parameters[1].Default != 42 {
		t.Errorf("Expected param2 default to be 42, got %v", td.Parameters[1].Default)
	}
}
