package utilitytools

import (
	"fmt"
	"sync"
)

// ToolFunc is a function signature for tools
type ToolFunc func(args map[string]interface{}) (interface{}, error)

// ToolInfo contains metadata about a tool
type ToolInfo struct {
	Name        string
	Description string
	Parameters  []ParamInfo
	Function    ToolFunc
}

// ParamInfo describes a parameter
type ParamInfo struct {
	Name     string
	Type     string
	Required bool
}

// Registry manages all available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*ToolInfo
}

// NewRegistry creates a new tool registry with all tools registered
func NewRegistry() *Registry {
	r := &Registry{
		tools: make(map[string]*ToolInfo),
	}
	r.registerTools()
	return r
}

// registerTools registers all generated tools
func (r *Registry) registerTools() {
	r.Register(&ToolInfo{
		Name:        "add",
		Description: "Add two numbers together",
		Parameters: []ParamInfo{
			{Name: "a", Type: "float64", Required: true},
			{Name: "b", Type: "float64", Required: true},
		},
		Function: add,
	})
	r.Register(&ToolInfo{
		Name:        "getCurrentTime",
		Description: "Get the current time in RFC3339 format",
		Parameters:  []ParamInfo{},
		Function:    getCurrentTime,
	})
	r.Register(&ToolInfo{
		Name:        "generateUUID",
		Description: "Generate a random UUID",
		Parameters:  []ParamInfo{},
		Function:    generateUUID,
	})
	r.Register(&ToolInfo{
		Name:        "concatenateStrings",
		Description: "Concatenate an array of strings with a separator",
		Parameters: []ParamInfo{
			{Name: "strings", Type: "[]interface{}", Required: true},
			{Name: "separator", Type: "string", Required: false},
		},
		Function: concatenateStrings,
	})
	r.Register(&ToolInfo{
		Name:        "reverseString",
		Description: "Reverse a string",
		Parameters: []ParamInfo{
			{Name: "text", Type: "string", Required: true},
		},
		Function: reverseString,
	})

}

// Register adds a tool to the registry
func (r *Registry) Register(info *ToolInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[info.Name]; exists {
		return fmt.Errorf("tool already registered: %s", info.Name)
	}

	r.tools[info.Name] = info
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*ToolInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, found := r.tools[name]
	return tool, found
}

// Call invokes a tool by name with arguments
func (r *Registry) Call(name string, args map[string]interface{}) (interface{}, error) {
	tool, found := r.Get(name)
	if !found {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Function(args)
}

// List returns all registered tool names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ListTools returns a list of all registered tools
func (r *Registry) ListTools() []*ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetDocumentation returns formatted documentation for all tools
func (r *Registry) GetDocumentation() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc := "Available Tools:\n\n"

	for _, tool := range r.tools {
		doc += fmt.Sprintf("## %s\n", tool.Name)
		doc += fmt.Sprintf("%s\n\n", tool.Description)

		if len(tool.Parameters) > 0 {
			doc += "Parameters:\n"
			for _, param := range tool.Parameters {
				required := ""
				if param.Required {
					required = " (required)"
				}
				doc += fmt.Sprintf("  - %s (%s)%s\n", param.Name, param.Type, required)
			}
			doc += "\n"
		}
	}

	return doc
}
