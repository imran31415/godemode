package codegen

import (
	"strings"
	"testing"

	"github.com/imran31415/godemode/pkg/spec"
)

func TestNewCodeGenerator(t *testing.T) {
	gen := NewCodeGenerator("testpkg")
	if gen == nil {
		t.Fatal("Expected code generator, got nil")
	}
	if gen.packageName != "testpkg" {
		t.Errorf("Expected package name 'testpkg', got '%s'", gen.packageName)
	}
}

func TestGenerateRegistry(t *testing.T) {
	gen := NewCodeGenerator("testtools")

	tools := []spec.ToolDefinition{
		{
			Name:        "getTool",
			Description: "Get a resource",
			Parameters: []spec.Parameter{
				{Name: "id", Type: "string", Required: true},
			},
		},
		{
			Name:        "createTool",
			Description: "Create a resource",
			Parameters: []spec.Parameter{
				{Name: "name", Type: "string", Required: true},
				{Name: "value", Type: "int", Required: false},
			},
		},
	}

	code, err := gen.GenerateRegistry(tools)
	if err != nil {
		t.Fatalf("Failed to generate registry: %v", err)
	}

	// Verify package declaration
	if !strings.Contains(code, "package testtools") {
		t.Error("Generated code should contain package declaration")
	}

	// Verify Registry struct exists
	if !strings.Contains(code, "type Registry struct") {
		t.Error("Generated code should contain Registry struct")
	}

	// Verify NewRegistry function exists
	if !strings.Contains(code, "func NewRegistry()") {
		t.Error("Generated code should contain NewRegistry function")
	}

	// Verify tools are registered
	if !strings.Contains(code, `Name:        "getTool"`) {
		t.Error("Generated code should register getTool")
	}

	if !strings.Contains(code, `Name:        "createTool"`) {
		t.Error("Generated code should register createTool")
	}

	// Verify parameters are included
	if !strings.Contains(code, `{Name: "id", Type: "string", Required: true}`) {
		t.Error("Generated code should include getTool parameters")
	}

	// Verify essential methods exist
	essentialMethods := []string{
		"func (r *Registry) Register(",
		"func (r *Registry) Get(",
		"func (r *Registry) Call(",
		"func (r *Registry) List()",
	}

	for _, method := range essentialMethods {
		if !strings.Contains(code, method) {
			t.Errorf("Generated code should contain method: %s", method)
		}
	}
}

func TestGenerateToolImplementation(t *testing.T) {
	gen := NewCodeGenerator("testtools")

	tool := spec.ToolDefinition{
		Name:        "sendEmail",
		Description: "Send an email",
		Parameters: []spec.Parameter{
			{Name: "to", Type: "string", Required: true, Description: "Recipient email"},
			{Name: "subject", Type: "string", Required: true, Description: "Email subject"},
			{Name: "body", Type: "string", Required: false, Description: "Email body"},
		},
	}

	code := gen.GenerateToolImplementation(tool)

	// Verify function signature
	if !strings.Contains(code, "func sendEmail(args map[string]interface{}) (interface{}, error)") {
		t.Error("Generated code should have correct function signature")
	}

	// Verify required parameters have validation
	if !strings.Contains(code, `args["to"]`) {
		t.Error("Generated code should extract 'to' parameter")
	}

	if !strings.Contains(code, `args["subject"]`) {
		t.Error("Generated code should extract 'subject' parameter")
	}

	// Verify required parameter validation
	if !strings.Contains(code, "required parameter 'to' not found") {
		t.Error("Generated code should validate required parameter 'to'")
	}

	// Verify TODO comment exists
	if !strings.Contains(code, "TODO: Implement your business logic here") {
		t.Error("Generated code should contain TODO comment")
	}

	// Verify stub return
	if !strings.Contains(code, `"status": "success"`) {
		t.Error("Generated code should return success status")
	}
}

func TestGenerateToolsFile(t *testing.T) {
	gen := NewCodeGenerator("mytools")

	tools := []spec.ToolDefinition{
		{
			Name:        "tool1",
			Description: "First tool",
			Parameters: []spec.Parameter{
				{Name: "param1", Type: "string", Required: true},
			},
		},
		{
			Name:        "tool2",
			Description: "Second tool",
			Parameters: []spec.Parameter{
				{Name: "param2", Type: "int", Required: false},
			},
		},
	}

	code, err := gen.GenerateToolsFile(tools)
	if err != nil {
		t.Fatalf("Failed to generate tools file: %v", err)
	}

	// Verify package declaration
	if !strings.Contains(code, "package mytools") {
		t.Error("Generated code should contain package declaration")
	}

	// Verify imports
	if !strings.Contains(code, `import`) {
		t.Error("Generated code should have imports")
	}

	// Verify both tool functions exist
	if !strings.Contains(code, "func tool1(args map[string]interface{}) (interface{}, error)") {
		t.Error("Generated code should contain tool1 function")
	}

	if !strings.Contains(code, "func tool2(args map[string]interface{}) (interface{}, error)") {
		t.Error("Generated code should contain tool2 function")
	}
}

func TestGenerateTypes(t *testing.T) {
	gen := NewCodeGenerator("types")

	tools := []spec.ToolDefinition{
		{Name: "test", Description: "Test tool"},
	}

	code, err := gen.GenerateTypes(tools)
	if err != nil {
		t.Fatalf("Failed to generate types: %v", err)
	}

	// Verify package declaration
	if !strings.Contains(code, "package types") {
		t.Error("Generated code should contain package declaration")
	}

	// Verify ToolFunc type
	if !strings.Contains(code, "type ToolFunc func(args map[string]interface{}) (interface{}, error)") {
		t.Error("Generated code should define ToolFunc type")
	}

	// Verify ToolInfo struct
	if !strings.Contains(code, "type ToolInfo struct") {
		t.Error("Generated code should define ToolInfo struct")
	}

	// Verify ParamInfo struct
	if !strings.Contains(code, "type ParamInfo struct") {
		t.Error("Generated code should define ParamInfo struct")
	}
}

func TestMapTypeToGo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"int", "int"},
		{"int32", "int"},
		{"int64", "int"},
		{"integer", "int"},
		{"float32", "float64"},
		{"float64", "float64"},
		{"number", "float64"},
		{"bool", "bool"},
		{"boolean", "bool"},
		{"map", "map[string]interface{}"},
		{"map[string]interface{}", "map[string]interface{}"},
		{"object", "map[string]interface{}"},
		{"[]string", "[]string"},
		{"[]int", "[]int"},
		{"[]interface{}", "[]interface{}"},
		{"unknown", "interface{}"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapTypeToGo(tt.input)
			if result != tt.expected {
				t.Errorf("mapTypeToGo(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateREADME(t *testing.T) {
	gen := NewCodeGenerator("mypackage")

	tools := []spec.ToolDefinition{
		{
			Name:        "createUser",
			Description: "Create a new user",
			Parameters: []spec.Parameter{
				{Name: "username", Type: "string", Required: true, Description: "User's username"},
				{Name: "email", Type: "string", Required: true, Description: "User's email"},
				{Name: "age", Type: "int", Required: false, Description: "User's age"},
			},
		},
	}

	readme := gen.GenerateREADME(tools, spec.FormatMCP)

	// Verify title
	if !strings.Contains(readme, "# Generated GoDeMode Tools") {
		t.Error("README should contain title")
	}

	// Verify spec format mentioned
	if !strings.Contains(readme, "mcp specification") {
		t.Error("README should mention spec format")
	}

	// Verify tool count
	if !strings.Contains(readme, "## Tools (1 total)") {
		t.Error("README should show correct tool count")
	}

	// Verify tool documentation
	if !strings.Contains(readme, "### createUser") {
		t.Error("README should document createUser tool")
	}

	if !strings.Contains(readme, "Create a new user") {
		t.Error("README should include tool description")
	}

	// Verify parameters are documented
	if !strings.Contains(readme, "`username`") {
		t.Error("README should document username parameter")
	}

	if !strings.Contains(readme, "*(required)*") {
		t.Error("README should mark required parameters")
	}

	// Verify usage example exists
	if !strings.Contains(readme, "## Usage") {
		t.Error("README should include usage section")
	}

	if !strings.Contains(readme, "NewRegistry()") {
		t.Error("README should show usage example")
	}
}

func TestGenerateRegistryWithNoTools(t *testing.T) {
	gen := NewCodeGenerator("empty")

	tools := []spec.ToolDefinition{}

	code, err := gen.GenerateRegistry(tools)
	if err != nil {
		t.Fatalf("Failed to generate registry with no tools: %v", err)
	}

	// Should still generate valid code
	if !strings.Contains(code, "package empty") {
		t.Error("Should generate valid package even with no tools")
	}

	if !strings.Contains(code, "func NewRegistry()") {
		t.Error("Should still have NewRegistry function")
	}
}

func TestGenerateRegistryComplexTypes(t *testing.T) {
	gen := NewCodeGenerator("complex")

	tools := []spec.ToolDefinition{
		{
			Name:        "processData",
			Description: "Process complex data",
			Parameters: []spec.Parameter{
				{Name: "data", Type: "map[string]interface{}", Required: true},
				{Name: "tags", Type: "[]string", Required: false},
				{Name: "count", Type: "int64", Required: false},
			},
		},
	}

	code, err := gen.GenerateRegistry(tools)
	if err != nil {
		t.Fatalf("Failed to generate registry with complex types: %v", err)
	}

	if !strings.Contains(code, `Name:        "processData"`) {
		t.Error("Should register tool with complex types")
	}

	// Verify tool implementation handles complex types
	impl := gen.GenerateToolImplementation(tools[0])

	if !strings.Contains(impl, "func processData") {
		t.Error("Should generate implementation for tool with complex types")
	}
}
