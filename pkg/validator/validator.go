package validator

import (
	"fmt"
	"strings"
)

// Validator validates Go source code before compilation
type Validator struct {
	maxSourceSize     int      // Maximum source code size in bytes
	forbiddenImports  []string // Imports that are not allowed
	forbiddenKeywords []string // Keywords that indicate dangerous code
}

// ValidationError represents a validation failure
type ValidationError struct {
	Type    string // "import", "size", "keyword"
	Message string
	Detail  string // Additional details
}

func (e *ValidationError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("validation error (%s): %s - %s", e.Type, e.Message, e.Detail)
	}
	return fmt.Sprintf("validation error (%s): %s", e.Type, e.Message)
}

// NewValidator creates a new Validator with default settings
func NewValidator() *Validator {
	return &Validator{
		maxSourceSize: 1024 * 1024, // 1MB max source code
		forbiddenImports: []string{
			"os/exec",       // Command execution
			"syscall",       // Low-level system calls
			"unsafe",        // Unsafe memory operations
			"plugin",        // Dynamic plugin loading
			"net/http",      // HTTP server (can be added to tools instead)
			"net",           // Network access (can be added to tools)
			"os/signal",     // Signal handling
			"runtime/debug", // Debug capabilities
			// Note: os package for file I/O is blocked by WASM sandbox anyway
		},
		forbiddenKeywords: []string{
			"//go:linkname", // Linker directives
			"//go:noescape", // Compiler directives
			"//export",      // Export directives (could be used maliciously)
		},
	}
}

// NewValidatorWithConfig creates a Validator with custom configuration
func NewValidatorWithConfig(maxSourceSize int, forbiddenImports, forbiddenKeywords []string) *Validator {
	return &Validator{
		maxSourceSize:     maxSourceSize,
		forbiddenImports:  forbiddenImports,
		forbiddenKeywords: forbiddenKeywords,
	}
}

// Validate checks if source code is safe to compile and execute
func (v *Validator) Validate(sourceCode string) error {
	// Check source size
	if len(sourceCode) > v.maxSourceSize {
		return &ValidationError{
			Type:    "size",
			Message: "source code too large",
			Detail:  fmt.Sprintf("maximum %d bytes allowed, got %d bytes", v.maxSourceSize, len(sourceCode)),
		}
	}

	// Check for empty source
	if strings.TrimSpace(sourceCode) == "" {
		return &ValidationError{
			Type:    "empty",
			Message: "source code is empty",
		}
	}

	// Check for forbidden imports
	if err := v.checkImports(sourceCode); err != nil {
		return err
	}

	// Check for forbidden keywords/directives
	if err := v.checkKeywords(sourceCode); err != nil {
		return err
	}

	return nil
}

// checkImports validates that no forbidden packages are imported
func (v *Validator) checkImports(sourceCode string) error {
	for _, forbiddenImport := range v.forbiddenImports {
		// Check for both single and grouped imports
		patterns := []string{
			fmt.Sprintf(`import "%s"`, forbiddenImport),
			fmt.Sprintf(`import "%s"`, forbiddenImport),
			fmt.Sprintf(`"%s"`, forbiddenImport), // In import group
		}

		for _, pattern := range patterns {
			if strings.Contains(sourceCode, pattern) {
				return &ValidationError{
					Type:    "import",
					Message: "forbidden import detected",
					Detail:  fmt.Sprintf("package '%s' is not allowed", forbiddenImport),
				}
			}
		}
	}

	return nil
}

// checkKeywords validates that no forbidden keywords/directives are used
func (v *Validator) checkKeywords(sourceCode string) error {
	for _, keyword := range v.forbiddenKeywords {
		if strings.Contains(sourceCode, keyword) {
			return &ValidationError{
				Type:    "keyword",
				Message: "forbidden keyword or directive detected",
				Detail:  fmt.Sprintf("'%s' is not allowed", keyword),
			}
		}
	}

	return nil
}

// AddForbiddenImport adds a package to the forbidden imports list
func (v *Validator) AddForbiddenImport(packagePath string) {
	v.forbiddenImports = append(v.forbiddenImports, packagePath)
}

// AddForbiddenKeyword adds a keyword to the forbidden keywords list
func (v *Validator) AddForbiddenKeyword(keyword string) {
	v.forbiddenKeywords = append(v.forbiddenKeywords, keyword)
}

// SetMaxSourceSize sets the maximum allowed source code size
func (v *Validator) SetMaxSourceSize(size int) {
	v.maxSourceSize = size
}

// GetForbiddenImports returns the list of forbidden imports
func (v *Validator) GetForbiddenImports() []string {
	return append([]string{}, v.forbiddenImports...) // Return copy
}

// GetForbiddenKeywords returns the list of forbidden keywords
func (v *Validator) GetForbiddenKeywords() []string {
	return append([]string{}, v.forbiddenKeywords...) // Return copy
}
