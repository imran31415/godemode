package exceltools

import (
	"fmt"
	"sync"
)

// ToolParameter represents a parameter for a tool
type ToolParameter struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

// ToolInfo contains information about a tool
type ToolInfo struct {
	Name        string
	Description string
	Parameters  []ToolParameter
	Function    func(args map[string]interface{}) (interface{}, error)
}

// Registry holds all available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*ToolInfo
}

// NewRegistry creates a new tool registry with all Excel tools
func NewRegistry() *Registry {
	r := &Registry{
		tools: make(map[string]*ToolInfo),
	}
	r.registerTools()
	return r
}

// Register adds a tool to the registry
func (r *Registry) Register(tool *ToolInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*ToolInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, found := r.tools[name]
	return tool, found
}

// Call executes a tool by name with the given arguments
func (r *Registry) Call(name string, args map[string]interface{}) (interface{}, error) {
	tool, found := r.Get(name)
	if !found {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool.Function(args)
}

// ListTools returns all registered tools
func (r *Registry) ListTools() []*ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// registerTools registers all Excel tools
func (r *Registry) registerTools() {
	r.Register(&ToolInfo{
		Name:        "describe_sheets",
		Description: "List all sheet information of specified Excel file",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
		},
		Function: describeSheets,
	})

	r.Register(&ToolInfo{
		Name:        "read_sheet",
		Description: "Read values from Excel sheet with optional range",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
			{Name: "sheet_name", Type: "string", Required: true, Description: "Sheet name in the Excel file"},
			{Name: "range", Type: "string", Required: false, Description: "Cell range (e.g., 'A1:C10')"},
		},
		Function: readSheet,
	})

	r.Register(&ToolInfo{
		Name:        "write_to_sheet",
		Description: "Write values to the Excel sheet",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
			{Name: "sheet_name", Type: "string", Required: true, Description: "Sheet name in the Excel file"},
			{Name: "range", Type: "string", Required: true, Description: "Cell range (e.g., 'A1:C10')"},
			{Name: "values", Type: "array", Required: true, Description: "2D array of values to write"},
			{Name: "new_sheet", Type: "boolean", Required: false, Description: "Create new sheet if true"},
		},
		Function: writeToSheet,
	})

	r.Register(&ToolInfo{
		Name:        "create_sheet",
		Description: "Create a new sheet in the Excel file",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
			{Name: "sheet_name", Type: "string", Required: true, Description: "Name for the new sheet"},
		},
		Function: createSheet,
	})

	r.Register(&ToolInfo{
		Name:        "delete_sheet",
		Description: "Delete a sheet from the Excel file",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
			{Name: "sheet_name", Type: "string", Required: true, Description: "Sheet name to delete"},
		},
		Function: deleteSheet,
	})

	r.Register(&ToolInfo{
		Name:        "copy_sheet",
		Description: "Copy existing sheet to a new sheet",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
			{Name: "src_sheet", Type: "string", Required: true, Description: "Source sheet name"},
			{Name: "dst_sheet", Type: "string", Required: true, Description: "Destination sheet name"},
		},
		Function: copySheet,
	})

	r.Register(&ToolInfo{
		Name:        "get_cell",
		Description: "Get the value of a specific cell",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
			{Name: "sheet_name", Type: "string", Required: true, Description: "Sheet name in the Excel file"},
			{Name: "cell", Type: "string", Required: true, Description: "Cell reference (e.g., 'A1')"},
		},
		Function: getCell,
	})

	r.Register(&ToolInfo{
		Name:        "set_cell",
		Description: "Set the value of a specific cell",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
			{Name: "sheet_name", Type: "string", Required: true, Description: "Sheet name in the Excel file"},
			{Name: "cell", Type: "string", Required: true, Description: "Cell reference (e.g., 'A1')"},
			{Name: "value", Type: "any", Required: true, Description: "Value to set (string, number, or formula starting with '=')"},
		},
		Function: setCell,
	})

	r.Register(&ToolInfo{
		Name:        "set_formula",
		Description: "Set a formula in a specific cell",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path to the Excel file"},
			{Name: "sheet_name", Type: "string", Required: true, Description: "Sheet name in the Excel file"},
			{Name: "cell", Type: "string", Required: true, Description: "Cell reference (e.g., 'A1')"},
			{Name: "formula", Type: "string", Required: true, Description: "Formula (e.g., '=SUM(A1:A10)')"},
		},
		Function: setFormula,
	})

	r.Register(&ToolInfo{
		Name:        "create_workbook",
		Description: "Create a new Excel workbook",
		Parameters: []ToolParameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Absolute path for the new Excel file"},
		},
		Function: createWorkbook,
	})
}
