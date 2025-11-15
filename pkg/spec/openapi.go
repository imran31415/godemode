package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// OpenAPISpec represents a simplified OpenAPI 3.x specification
type OpenAPISpec struct {
	OpenAPI string                       `json:"openapi"`
	Info    OpenAPIInfo                  `json:"info"`
	Paths   map[string]OpenAPIPathItem   `json:"paths"`
	Servers []OpenAPIServer              `json:"servers,omitempty"`
}

// OpenAPIInfo contains API metadata
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// OpenAPIServer represents a server endpoint
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// OpenAPIPathItem represents operations on a path
type OpenAPIPathItem struct {
	Get    *OpenAPIOperation `json:"get,omitempty"`
	Post   *OpenAPIOperation `json:"post,omitempty"`
	Put    *OpenAPIOperation `json:"put,omitempty"`
	Delete *OpenAPIOperation `json:"delete,omitempty"`
	Patch  *OpenAPIOperation `json:"patch,omitempty"`
}

// OpenAPIOperation represents an API operation
type OpenAPIOperation struct {
	OperationID string                         `json:"operationId,omitempty"`
	Summary     string                         `json:"summary,omitempty"`
	Description string                         `json:"description,omitempty"`
	Parameters  []OpenAPIParameter             `json:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody            `json:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse     `json:"responses,omitempty"`
}

// OpenAPIParameter represents a parameter
type OpenAPIParameter struct {
	Name        string                 `json:"name"`
	In          string                 `json:"in"` // query, header, path, cookie
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required,omitempty"`
	Schema      *OpenAPISchema         `json:"schema,omitempty"`
}

// OpenAPIRequestBody represents a request body
type OpenAPIRequestBody struct {
	Description string                            `json:"description,omitempty"`
	Required    bool                              `json:"required,omitempty"`
	Content     map[string]OpenAPIMediaType       `json:"content,omitempty"`
}

// OpenAPIMediaType represents a media type object
type OpenAPIMediaType struct {
	Schema *OpenAPISchema `json:"schema,omitempty"`
}

// OpenAPIResponse represents a response
type OpenAPIResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty"`
}

// OpenAPISchema represents a schema object
type OpenAPISchema struct {
	Type       string                    `json:"type,omitempty"`
	Format     string                    `json:"format,omitempty"`
	Properties map[string]*OpenAPISchema `json:"properties,omitempty"`
	Items      *OpenAPISchema            `json:"items,omitempty"`
	Required   []string                  `json:"required,omitempty"`
	Enum       []interface{}             `json:"enum,omitempty"`
	Default    interface{}               `json:"default,omitempty"`
}

// ParseOpenAPISpec parses an OpenAPI specification from a file
func ParseOpenAPISpec(filePath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI spec file: %w", err)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	return &spec, nil
}

// ParseOpenAPISpecFromBytes parses an OpenAPI specification from bytes
func ParseOpenAPISpecFromBytes(data []byte) (*OpenAPISpec, error) {
	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	return &spec, nil
}

// ToToolDefinitions converts OpenAPI operations to ToolDefinitions
func (s *OpenAPISpec) ToToolDefinitions() []ToolDefinition {
	var tools []ToolDefinition

	for path, pathItem := range s.Paths {
		// Process each HTTP method
		if pathItem.Get != nil {
			tools = append(tools, s.operationToTool("GET", path, pathItem.Get))
		}
		if pathItem.Post != nil {
			tools = append(tools, s.operationToTool("POST", path, pathItem.Post))
		}
		if pathItem.Put != nil {
			tools = append(tools, s.operationToTool("PUT", path, pathItem.Put))
		}
		if pathItem.Delete != nil {
			tools = append(tools, s.operationToTool("DELETE", path, pathItem.Delete))
		}
		if pathItem.Patch != nil {
			tools = append(tools, s.operationToTool("PATCH", path, pathItem.Patch))
		}
	}

	return tools
}

// operationToTool converts an OpenAPI operation to a ToolDefinition
func (s *OpenAPISpec) operationToTool(method, path string, op *OpenAPIOperation) ToolDefinition {
	// Use operationId as name, or generate from method + path
	name := op.OperationID
	if name == "" {
		name = generateOperationName(method, path)
	}

	// Use summary or description
	description := op.Summary
	if description == "" {
		description = op.Description
	}
	if description == "" {
		description = fmt.Sprintf("%s %s", method, path)
	}

	// Extract parameters
	var params []Parameter

	// Add path/query/header parameters
	for _, param := range op.Parameters {
		params = append(params, Parameter{
			Name:        param.Name,
			Type:        mapOpenAPITypeToGo(param.Schema),
			Description: param.Description,
			Required:    param.Required || param.In == "path",
		})
	}

	// Add request body parameters if present
	if op.RequestBody != nil {
		params = append(params, extractRequestBodyParams(op.RequestBody)...)
	}

	return ToolDefinition{
		Name:        name,
		Description: description,
		Parameters:  params,
	}
}

// generateOperationName generates an operation name from method and path
func generateOperationName(method, path string) string {
	// Clean up path and convert to camelCase
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var nameParts []string

	nameParts = append(nameParts, strings.ToLower(method))

	for _, part := range parts {
		// Skip path parameters
		if strings.HasPrefix(part, "{") {
			continue
		}
		// Capitalize first letter
		if len(part) > 0 {
			nameParts = append(nameParts, strings.ToUpper(part[:1])+part[1:])
		}
	}

	return strings.Join(nameParts, "")
}

// extractRequestBodyParams extracts parameters from request body
func extractRequestBodyParams(body *OpenAPIRequestBody) []Parameter {
	var params []Parameter

	// Look for JSON content
	if jsonContent, ok := body.Content["application/json"]; ok && jsonContent.Schema != nil {
		if jsonContent.Schema.Properties != nil {
			for propName, propSchema := range jsonContent.Schema.Properties {
				required := false
				for _, req := range jsonContent.Schema.Required {
					if req == propName {
						required = true
						break
					}
				}

				params = append(params, Parameter{
					Name:        propName,
					Type:        mapOpenAPITypeToGo(propSchema),
					Description: "", // OpenAPI schemas don't have description at property level
					Required:    required,
				})
			}
		}
	}

	return params
}

// mapOpenAPITypeToGo maps OpenAPI types to Go types
func mapOpenAPITypeToGo(schema *OpenAPISchema) string {
	if schema == nil {
		return "interface{}"
	}

	switch schema.Type {
	case "string":
		if schema.Format == "date-time" {
			return "time.Time"
		}
		return "string"
	case "number":
		if schema.Format == "float" {
			return "float32"
		}
		return "float64"
	case "integer":
		if schema.Format == "int32" {
			return "int32"
		}
		return "int64"
	case "boolean":
		return "bool"
	case "array":
		if schema.Items != nil {
			return "[]" + mapOpenAPITypeToGo(schema.Items)
		}
		return "[]interface{}"
	case "object":
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}
