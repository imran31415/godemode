package spec

import (
	"encoding/json"
	"fmt"
	"os"
)

// MCPSpec represents an MCP (Model Context Protocol) specification
type MCPSpec struct {
	Name        string                `json:"name"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	Tools       []MCPTool             `json:"tools"`
	Resources   []MCPResource         `json:"resources,omitempty"`
	Prompts     []MCPPrompt           `json:"prompts,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MCPTool represents a tool in the MCP spec
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema MCPSchema              `json:"inputSchema"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MCPResource represents a resource in the MCP spec
type MCPResource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MCPPrompt represents a prompt template in the MCP spec
type MCPPrompt struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Arguments   []MCPArgument          `json:"arguments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MCPArgument represents an argument in a prompt
type MCPArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// MCPSchema represents a JSON schema for tool input
type MCPSchema struct {
	Type       string                  `json:"type"`
	Properties map[string]MCPProperty  `json:"properties,omitempty"`
	Required   []string                `json:"required,omitempty"`
	Items      *MCPSchema              `json:"items,omitempty"`
}

// MCPProperty represents a property in a JSON schema
type MCPProperty struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Enum        []interface{}          `json:"enum,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Items       *MCPSchema             `json:"items,omitempty"`
	Properties  map[string]MCPProperty `json:"properties,omitempty"`
}

// ParseMCPSpec parses an MCP specification from a file
func ParseMCPSpec(filePath string) (*MCPSpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP spec file: %w", err)
	}

	var spec MCPSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse MCP spec: %w", err)
	}

	return &spec, nil
}

// ParseMCPSpecFromBytes parses an MCP specification from bytes
func ParseMCPSpecFromBytes(data []byte) (*MCPSpec, error) {
	var spec MCPSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse MCP spec: %w", err)
	}

	return &spec, nil
}

// Validate validates the MCP specification
func (s *MCPSpec) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("MCP spec must have a name")
	}

	if len(s.Tools) == 0 {
		return fmt.Errorf("MCP spec must have at least one tool")
	}

	for i, tool := range s.Tools {
		if tool.Name == "" {
			return fmt.Errorf("tool at index %d must have a name", i)
		}
		if tool.Description == "" {
			return fmt.Errorf("tool %s must have a description", tool.Name)
		}
	}

	return nil
}

// ToToolDefinitions converts MCP tools to a common ToolDefinition format
func (s *MCPSpec) ToToolDefinitions() []ToolDefinition {
	tools := make([]ToolDefinition, len(s.Tools))

	for i, mcpTool := range s.Tools {
		tools[i] = ToolDefinition{
			Name:        mcpTool.Name,
			Description: mcpTool.Description,
			Parameters:  extractParameters(mcpTool.InputSchema),
		}
	}

	return tools
}

// extractParameters converts MCP schema properties to parameters
func extractParameters(schema MCPSchema) []Parameter {
	if schema.Properties == nil {
		return nil
	}

	params := make([]Parameter, 0, len(schema.Properties))
	for name, prop := range schema.Properties {
		required := false
		for _, req := range schema.Required {
			if req == name {
				required = true
				break
			}
		}

		params = append(params, Parameter{
			Name:        name,
			Type:        mapMCPTypeToGo(prop.Type),
			Description: prop.Description,
			Required:    required,
		})
	}

	return params
}

// mapMCPTypeToGo maps MCP/JSON Schema types to Go types
func mapMCPTypeToGo(mcpType string) string {
	switch mcpType {
	case "string":
		return "string"
	case "number":
		return "float64"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	case "array":
		return "[]interface{}"
	case "object":
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}
