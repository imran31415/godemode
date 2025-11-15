package dataprocessing

import (
	"fmt"
)

type ToolInfo struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Function    func(map[string]interface{}) (interface{}, error)
}

type Registry struct {
	tools map[string]ToolInfo
}

func NewRegistry() *Registry {
	r := &Registry{
		tools: make(map[string]ToolInfo),
	}
	r.registerTools()
	return r
}

func (r *Registry) registerTools() {
	r.Register(ToolInfo{
		Name:        "filterArray",
		Description: "Filter an array based on a condition",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data":      map[string]interface{}{"type": "array", "description": "Array of numbers to filter"},
				"operation": map[string]interface{}{"type": "string", "description": "Operation: gt, lt, eq, gte, lte"},
				"value":     map[string]interface{}{"type": "number", "description": "Value to compare against"},
			},
			"required": []string{"data", "operation", "value"},
		},
		Function: filterArray,
	})

	r.Register(ToolInfo{
		Name:        "mapArray",
		Description: "Transform each element in an array",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data":      map[string]interface{}{"type": "array", "description": "Array to transform"},
				"operation": map[string]interface{}{"type": "string", "description": "Operation: double, square, negate"},
			},
			"required": []string{"data", "operation"},
		},
		Function: mapArray,
	})

	r.Register(ToolInfo{
		Name:        "reduceArray",
		Description: "Reduce an array to a single value",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data":      map[string]interface{}{"type": "array", "description": "Array of numbers"},
				"operation": map[string]interface{}{"type": "string", "description": "Operation: sum, product, max, min, avg"},
			},
			"required": []string{"data", "operation"},
		},
		Function: reduceArray,
	})

	r.Register(ToolInfo{
		Name:        "sortArray",
		Description: "Sort an array",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data":  map[string]interface{}{"type": "array", "description": "Array to sort"},
				"order": map[string]interface{}{"type": "string", "description": "Order: asc or desc"},
			},
			"required": []string{"data", "order"},
		},
		Function: sortArray,
	})

	r.Register(ToolInfo{
		Name:        "groupBy",
		Description: "Group array elements by a key",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{"type": "array", "description": "Array of objects"},
				"key":  map[string]interface{}{"type": "string", "description": "Key to group by"},
			},
			"required": []string{"data", "key"},
		},
		Function: groupBy,
	})

	r.Register(ToolInfo{
		Name:        "mergeArrays",
		Description: "Merge multiple arrays",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"arrays": map[string]interface{}{"type": "array", "description": "Arrays to merge"},
			},
			"required": []string{"arrays"},
		},
		Function: mergeArrays,
	})

	r.Register(ToolInfo{
		Name:        "uniqueValues",
		Description: "Get unique values from an array",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{"type": "array", "description": "Array with potential duplicates"},
			},
			"required": []string{"data"},
		},
		Function: uniqueValues,
	})
}

func (r *Registry) Register(tool ToolInfo) {
	r.tools[tool.Name] = tool
}

func (r *Registry) Call(name string, args map[string]interface{}) (interface{}, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}
	return tool.Function(args)
}

func (r *Registry) List() []ToolInfo {
	tools := make([]ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (r *Registry) GetToolDescriptions() string {
	var desc string
	for _, tool := range r.tools {
		desc += fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description)
	}
	return desc
}
