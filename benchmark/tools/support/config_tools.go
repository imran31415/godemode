package support

import (
	"fmt"

	"github.com/imran31415/godemode/benchmark/systems/filesystem"
	"github.com/imran31415/godemode/benchmark/tools"
)

// ConfigTools provides configuration file tools
type ConfigTools struct {
	configSystem *filesystem.ConfigSystem
}

// NewConfigTools creates config tools
func NewConfigTools(configSystem *filesystem.ConfigSystem) *ConfigTools {
	return &ConfigTools{configSystem: configSystem}
}

// RegisterTools registers all config tools with the registry
func (ct *ConfigTools) RegisterTools(registry *tools.Registry) error {
	// ReadConfig tool
	err := registry.Register(&tools.ToolInfo{
		Name:        "readConfig",
		Description: "Read a configuration file (JSON or YAML)",
		Parameters: []tools.ParamInfo{
			{Name: "filename", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			filename, ok := args["filename"]
			if !ok {
				return nil, fmt.Errorf("readConfig requires filename parameter")
			}

			filenameStr := fmt.Sprintf("%v", filename)
			return ct.configSystem.ReadConfig(filenameStr)
		},
	})
	if err != nil {
		return err
	}

	// CheckFeatureFlag tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "checkFeatureFlag",
		Description: "Check if a feature flag is enabled",
		Parameters: []tools.ParamInfo{
			{Name: "flagName", Type: "string", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			flagName, ok := args["flagName"]
			if !ok {
				return nil, fmt.Errorf("checkFeatureFlag requires flagName parameter")
			}

			flagNameStr := fmt.Sprintf("%v", flagName)
			return ct.configSystem.CheckFeatureFlag(flagNameStr)
		},
	})
	if err != nil {
		return err
	}

	// UpdateConfig tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "updateConfig",
		Description: "Update a configuration file",
		Parameters: []tools.ParamInfo{
			{Name: "filename", Type: "string", Required: true},
			{Name: "updates", Type: "map", Required: true},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			filename, ok := args["filename"]
			if !ok {
				return nil, fmt.Errorf("updateConfig requires filename parameter")
			}

			filenameStr := fmt.Sprintf("%v", filename)

			updates, ok := args["updates"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("updates must be a map")
			}

			err := ct.configSystem.UpdateConfig(filenameStr, updates)
			return err == nil, err
		},
	})

	return err
}
