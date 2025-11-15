package spec

import "encoding/json"

// ToolDefinition represents a normalized tool definition from any spec format
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  []Parameter
}

// Parameter represents a function parameter
type Parameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     interface{}
	Enum        []interface{}
}

// SpecFormat represents the format of a specification file
type SpecFormat string

const (
	FormatMCP     SpecFormat = "mcp"
	FormatOpenAPI SpecFormat = "openapi"
	FormatUnknown SpecFormat = "unknown"
)

// DetectSpecFormat detects the format of a specification file
func DetectSpecFormat(data []byte) SpecFormat {
	// Try to detect MCP format
	var mcpCheck struct {
		Tools []interface{} `json:"tools"`
	}
	if err := json.Unmarshal(data, &mcpCheck); err == nil && len(mcpCheck.Tools) > 0 {
		return FormatMCP
	}

	// Try to detect OpenAPI format
	var openAPICheck struct {
		OpenAPI string `json:"openapi"`
		Swagger string `json:"swagger"`
	}
	if err := json.Unmarshal(data, &openAPICheck); err == nil {
		if openAPICheck.OpenAPI != "" || openAPICheck.Swagger != "" {
			return FormatOpenAPI
		}
	}

	return FormatUnknown
}
