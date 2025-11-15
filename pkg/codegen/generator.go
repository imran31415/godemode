package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/imran31415/godemode/pkg/spec"
)

// CodeGenerator generates Go code from tool definitions
type CodeGenerator struct {
	packageName string
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator(packageName string) *CodeGenerator {
	return &CodeGenerator{
		packageName: packageName,
	}
}

// GenerateRegistry generates a tool registry file from tool definitions
func (g *CodeGenerator) GenerateRegistry(tools []spec.ToolDefinition) (string, error) {
	tmpl, err := template.New("registry").Parse(registryTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		PackageName string
		Tools       []spec.ToolDefinition
	}{
		PackageName: g.packageName,
		Tools:       tools,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format generated code: %w", err)
	}

	return string(formatted), nil
}

// GenerateToolImplementation generates stub implementation for a single tool
func (g *CodeGenerator) GenerateToolImplementation(tool spec.ToolDefinition) string {
	var buf bytes.Buffer

	// Generate function signature
	buf.WriteString(fmt.Sprintf("func %s(args map[string]interface{}) (interface{}, error) {\n", tool.Name))

	// Extract parameters
	for _, param := range tool.Parameters {
		if param.Required {
			buf.WriteString(fmt.Sprintf("\t// Required parameter: %s (%s)\n", param.Name, param.Type))
			buf.WriteString(fmt.Sprintf("\t%s, ok := args[\"%s\"]", param.Name, param.Name))
			buf.WriteString(".(")
			buf.WriteString(mapTypeToGo(param.Type))
			buf.WriteString(")\n")
			buf.WriteString("\tif !ok {\n")
			buf.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"required parameter '%s' not found or wrong type\")\n", param.Name))
			buf.WriteString("\t}\n")
			buf.WriteString(fmt.Sprintf("\t_ = %s // TODO: Use this parameter in your implementation\n\n", param.Name))
		} else {
			buf.WriteString(fmt.Sprintf("\t// Optional parameter: %s (%s)\n", param.Name, param.Type))
			buf.WriteString(fmt.Sprintf("\t%s, _ := args[\"%s\"]", param.Name, param.Name))
			buf.WriteString(".(")
			buf.WriteString(mapTypeToGo(param.Type))
			buf.WriteString(")\n")
			buf.WriteString(fmt.Sprintf("\t_ = %s // TODO: Use this parameter in your implementation\n\n", param.Name))
		}
	}

	// Add TODO comment for implementation
	buf.WriteString("\t// TODO: Implement your business logic here\n")
	buf.WriteString("\t// This is a stub implementation\n\n")

	// Return placeholder result
	buf.WriteString("\treturn map[string]interface{}{\n")
	buf.WriteString("\t\t\"status\": \"success\",\n")
	buf.WriteString(fmt.Sprintf("\t\t\"message\": \"%s executed\",\n", tool.Name))
	buf.WriteString("\t}, nil\n")

	buf.WriteString("}\n")

	return buf.String()
}

// GenerateToolsFile generates a complete tools.go file with all tool implementations
func (g *CodeGenerator) GenerateToolsFile(tools []spec.ToolDefinition) (string, error) {
	tmpl, err := template.New("tools").Parse(toolsFileTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Generate individual tool implementations
	implementations := make([]string, len(tools))
	for i, tool := range tools {
		implementations[i] = g.GenerateToolImplementation(tool)
	}

	data := struct {
		PackageName     string
		Implementations []string
	}{
		PackageName:     g.packageName,
		Implementations: implementations,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format generated code: %w", err)
	}

	return string(formatted), nil
}

// GenerateTypes generates type definitions from tool definitions
func (g *CodeGenerator) GenerateTypes(tools []spec.ToolDefinition) (string, error) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("package %s\n\n", g.packageName))
	buf.WriteString("// ToolFunc is a function signature for tools\n")
	buf.WriteString("type ToolFunc func(args map[string]interface{}) (interface{}, error)\n\n")

	buf.WriteString("// ToolInfo contains metadata about a tool\n")
	buf.WriteString("type ToolInfo struct {\n")
	buf.WriteString("\tName        string\n")
	buf.WriteString("\tDescription string\n")
	buf.WriteString("\tParameters  []ParamInfo\n")
	buf.WriteString("\tFunction    ToolFunc\n")
	buf.WriteString("}\n\n")

	buf.WriteString("// ParamInfo describes a parameter\n")
	buf.WriteString("type ParamInfo struct {\n")
	buf.WriteString("\tName     string\n")
	buf.WriteString("\tType     string\n")
	buf.WriteString("\tRequired bool\n")
	buf.WriteString("}\n")

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format generated code: %w", err)
	}

	return string(formatted), nil
}

// mapTypeToGo converts spec types to Go types for code generation
func mapTypeToGo(specType string) string {
	switch specType {
	case "string":
		return "string"
	case "int", "int32", "int64", "integer":
		return "int"
	case "float32", "float64", "number":
		return "float64"
	case "bool", "boolean":
		return "bool"
	case "map", "map[string]interface{}", "object":
		return "map[string]interface{}"
	case "[]string":
		return "[]string"
	case "[]int":
		return "[]int"
	case "[]interface{}":
		return "[]interface{}"
	default:
		// Try to handle array types
		if strings.HasPrefix(specType, "[]") {
			return specType
		}
		return "interface{}"
	}
}

// GenerateREADME generates a README file documenting the generated tools
func (g *CodeGenerator) GenerateREADME(tools []spec.ToolDefinition, specFormat spec.SpecFormat) string {
	var buf bytes.Buffer

	buf.WriteString("# Generated GoDeMode Tools\n\n")
	buf.WriteString(fmt.Sprintf("This package was auto-generated from a %s specification.\n\n", specFormat))
	buf.WriteString(fmt.Sprintf("## Tools (%d total)\n\n", len(tools)))

	for _, tool := range tools {
		buf.WriteString(fmt.Sprintf("### %s\n\n", tool.Name))
		buf.WriteString(fmt.Sprintf("%s\n\n", tool.Description))

		if len(tool.Parameters) > 0 {
			buf.WriteString("**Parameters:**\n\n")
			for _, param := range tool.Parameters {
				required := ""
				if param.Required {
					required = " *(required)*"
				}
				desc := param.Description
				if desc == "" {
					desc = "No description provided"
				}
				buf.WriteString(fmt.Sprintf("- `%s` (%s)%s - %s\n", param.Name, param.Type, required, desc))
			}
			buf.WriteString("\n")
		}
	}

	buf.WriteString("## Usage\n\n")
	buf.WriteString("```go\n")
	buf.WriteString("package main\n\n")
	buf.WriteString("import (\n")
	buf.WriteString(fmt.Sprintf("\t\"generated/%s\"\n", g.packageName))
	buf.WriteString(")\n\n")
	buf.WriteString("func main() {\n")
	buf.WriteString("\t// Create registry\n")
	buf.WriteString(fmt.Sprintf("\tregistry := %s.NewRegistry()\n\n", g.packageName))
	buf.WriteString("\t// Call a tool\n")
	buf.WriteString("\tresult, err := registry.Call(\"toolName\", map[string]interface{}{\n")
	buf.WriteString("\t\t\"param1\": \"value1\",\n")
	buf.WriteString("\t})\n")
	buf.WriteString("}\n")
	buf.WriteString("```\n")

	return buf.String()
}
