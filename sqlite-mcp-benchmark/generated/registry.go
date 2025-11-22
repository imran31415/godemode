package sqlitetools

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
		Name:        "db_info",
		Description: "Get detailed information about the connected database including file path, size, and table count.",
		Parameters:  []ParamInfo{},
		Function:    db_info,
	})
	r.Register(&ToolInfo{
		Name:        "list_tables",
		Description: "List all tables in the database.",
		Parameters:  []ParamInfo{},
		Function:    list_tables,
	})
	r.Register(&ToolInfo{
		Name:        "get_table_schema",
		Description: "Get detailed information about a table's schema including column names, types, and constraints.",
		Parameters: []ParamInfo{
			{Name: "tableName", Type: "string", Required: true},
		},
		Function: get_table_schema,
	})
	r.Register(&ToolInfo{
		Name:        "create_record",
		Description: "Insert a new record into a table.",
		Parameters: []ParamInfo{
			{Name: "table", Type: "string", Required: true},
			{Name: "data", Type: "map[string]interface{}", Required: true},
		},
		Function: create_record,
	})
	r.Register(&ToolInfo{
		Name:        "read_records",
		Description: "Query records from a table with optional filtering, pagination, and sorting.",
		Parameters: []ParamInfo{
			{Name: "table", Type: "string", Required: true},
			{Name: "conditions", Type: "map[string]interface{}", Required: false},
			{Name: "limit", Type: "float64", Required: false},
			{Name: "offset", Type: "float64", Required: false},
		},
		Function: read_records,
	})
	r.Register(&ToolInfo{
		Name:        "update_records",
		Description: "Update records in a table that match specified conditions.",
		Parameters: []ParamInfo{
			{Name: "table", Type: "string", Required: true},
			{Name: "data", Type: "map[string]interface{}", Required: true},
			{Name: "conditions", Type: "map[string]interface{}", Required: true},
		},
		Function: update_records,
	})
	r.Register(&ToolInfo{
		Name:        "delete_records",
		Description: "Delete records from a table that match specified conditions.",
		Parameters: []ParamInfo{
			{Name: "conditions", Type: "map[string]interface{}", Required: true},
			{Name: "table", Type: "string", Required: true},
		},
		Function: delete_records,
	})
	r.Register(&ToolInfo{
		Name:        "query",
		Description: "Execute a custom SQL query against the connected SQLite database. Supports parameterized queries for security.",
		Parameters: []ParamInfo{
			{Name: "sql", Type: "string", Required: true},
			{Name: "values", Type: "[]interface{}", Required: false},
		},
		Function: query_exec,
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
