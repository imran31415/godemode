package support

import (
	"fmt"
	"time"

	"github.com/imran31415/godemode/benchmark/systems/filesystem"
	"github.com/imran31415/godemode/benchmark/tools"
)

// LogTools provides log file tools
type LogTools struct {
	logSystem *filesystem.LogSystem
}

// NewLogTools creates log tools
func NewLogTools(logSystem *filesystem.LogSystem) *LogTools {
	return &LogTools{logSystem: logSystem}
}

// RegisterTools registers all log tools with the registry
func (lt *LogTools) RegisterTools(registry *tools.Registry) error {
	// SearchLogs tool
	err := registry.Register(&tools.ToolInfo{
		Name:        "searchLogs",
		Description: "Search log files for a pattern within a time window",
		Parameters: []tools.ParamInfo{
			{Name: "pattern", Type: "string", Required: true},
			{Name: "timeWindow", Type: "duration", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			pattern, ok := args["pattern"]
			if !ok {
				return nil, fmt.Errorf("searchLogs requires pattern parameter")
			}

			patternStr := fmt.Sprintf("%v", pattern)

			timeWindow := 24 * time.Hour // default
			if tw, ok := args["timeWindow"]; ok {
				if twDuration, ok := tw.(time.Duration); ok {
					timeWindow = twDuration
				}
			}

			return lt.logSystem.SearchLogs(patternStr, timeWindow)
		},
	})
	if err != nil {
		return err
	}

	// ExtractErrorContext tool
	err = registry.Register(&tools.ToolInfo{
		Name:        "extractErrorContext",
		Description: "Extract context lines around an error code in logs",
		Parameters: []tools.ParamInfo{
			{Name: "errorCode", Type: "string", Required: true},
			{Name: "contextLines", Type: "int", Required: false},
		},
		Function: func(args map[string]interface{}) (interface{}, error) {
			errorCode, ok := args["errorCode"]
			if !ok {
				return nil, fmt.Errorf("extractErrorContext requires errorCode parameter")
			}

			errorCodeStr := fmt.Sprintf("%v", errorCode)

			contextLines := 5 // default
			if cl, ok := args["contextLines"]; ok {
				if clInt, ok := cl.(int); ok {
					contextLines = clInt
				}
			}

			return lt.logSystem.ExtractErrorContext(errorCodeStr, contextLines)
		},
	})

	return err
}
