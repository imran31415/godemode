package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigSystem handles reading and writing config files
type ConfigSystem struct {
	configPath string
}

// NewConfigSystem creates a new config system
func NewConfigSystem(configPath string) *ConfigSystem {
	return &ConfigSystem{configPath: configPath}
}

// ReadConfig reads a config file (supports JSON and YAML)
func (cs *ConfigSystem) ReadConfig(filename string) (map[string]interface{}, error) {
	fullPath := filepath.Join(cs.configPath, filename)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", filename, err)
	}

	var config map[string]interface{}

	// Determine format by extension
	ext := filepath.Ext(filename)

	switch ext {
	case ".json":
		err = json.Unmarshal(data, &config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported config format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// WriteConfig writes a config file
func (cs *ConfigSystem) WriteConfig(filename string, config map[string]interface{}) error {
	fullPath := filepath.Join(cs.configPath, filename)

	var data []byte
	var err error

	// Determine format by extension
	ext := filepath.Ext(filename)

	switch ext {
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported config format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// UpdateConfig updates specific fields in a config file
func (cs *ConfigSystem) UpdateConfig(filename string, updates map[string]interface{}) error {
	// Read existing config
	config, err := cs.ReadConfig(filename)
	if err != nil {
		return err
	}

	// Apply updates
	for key, value := range updates {
		config[key] = value
	}

	// Write back
	return cs.WriteConfig(filename, config)
}

// GetValue gets a specific value from a config file
func (cs *ConfigSystem) GetValue(filename, key string) (interface{}, error) {
	config, err := cs.ReadConfig(filename)
	if err != nil {
		return nil, err
	}

	value, exists := config[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// CheckFeatureFlag checks if a feature flag is enabled
func (cs *ConfigSystem) CheckFeatureFlag(flagName string) (bool, error) {
	config, err := cs.ReadConfig("feature_flags.json")
	if err != nil {
		return false, err
	}

	value, exists := config[flagName]
	if !exists {
		return false, nil  // Flag doesn't exist = disabled
	}

	enabled, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("feature flag %s is not a boolean", flagName)
	}

	return enabled, nil
}

// Reset clears all config files
func (cs *ConfigSystem) Reset() error {
	files, err := filepath.Glob(filepath.Join(cs.configPath, "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil || info.IsDir() {
			continue
		}
		os.Remove(file)
	}

	return nil
}
