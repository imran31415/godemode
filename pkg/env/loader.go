package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load reads environment variables from a .env file
// Returns error if file doesn't exist or cannot be read
func Load(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		// If .env doesn't exist, that's okay - just use system env
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error opening .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format on line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("error setting %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
}

// LoadDefault loads from .env in the current directory
func LoadDefault() error {
	return Load(".env")
}

// MustLoad loads .env file and panics on error
func MustLoad(filepath string) {
	if err := Load(filepath); err != nil {
		panic(fmt.Sprintf("Failed to load .env file: %v", err))
	}
}
