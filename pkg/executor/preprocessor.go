package executor

import (
	"regexp"
	"strings"
)

// CodePreprocessor handles preprocessing of generated Go code before execution
type CodePreprocessor struct{}

// NewCodePreprocessor creates a new code preprocessor
func NewCodePreprocessor() *CodePreprocessor {
	return &CodePreprocessor{}
}

// ExtractGoCode extracts Go code from markdown code blocks
// Handles various formats: ```go, ```, and raw code
func (p *CodePreprocessor) ExtractGoCode(text string) string {
	// Try to find ```go ... ``` blocks
	re := regexp.MustCompile("(?s)```go\\s*\n(.*?)```")
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try with just ``` markers
	re = regexp.MustCompile("(?s)```\\s*\n(.*?)```")
	matches = re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Strip any remaining markdown artifacts
	result := text

	// Remove opening code block markers
	result = regexp.MustCompile("^\\s*```go\\s*\n?").ReplaceAllString(result, "")
	result = regexp.MustCompile("^\\s*```\\s*\n?").ReplaceAllString(result, "")

	// Remove closing code block markers
	result = regexp.MustCompile("\n?```\\s*$").ReplaceAllString(result, "")

	return strings.TrimSpace(result)
}

// PrepareForExecution prepares code for execution with custom symbols
// This transforms registry.Call to the injected function name
func (p *CodePreprocessor) PrepareForExecution(code string, registryFuncName string) string {
	// Replace registry.Call with the custom function name
	modifiedCode := strings.Replace(code, "registry.Call", registryFuncName, -1)

	// Remove registry variable declarations that would conflict
	modifiedCode = regexp.MustCompile(`(?m)^var registry.*$`).ReplaceAllString(modifiedCode, "")
	modifiedCode = regexp.MustCompile(`(?ms)type Registry interface \{[^}]*\}`).ReplaceAllString(modifiedCode, "")

	// Add import for the custom symbols package
	if !strings.Contains(modifiedCode, `"main"`) && strings.Contains(modifiedCode, "package main") {
		modifiedCode = strings.Replace(modifiedCode, "package main", `package main

import . "main"`, 1)
	}

	return modifiedCode
}

// Process applies all preprocessing steps to prepare code for execution
// Returns the processed code ready for the interpreter
func (p *CodePreprocessor) Process(rawCode string, registryFuncName string) string {
	// Step 1: Extract code from markdown
	code := p.ExtractGoCode(rawCode)

	// Step 2: Prepare for execution with custom symbols
	code = p.PrepareForExecution(code, registryFuncName)

	return code
}

// ValidateBasicStructure performs basic validation on the code structure
// Returns an error message if issues are found, empty string otherwise
func (p *CodePreprocessor) ValidateBasicStructure(code string) string {
	// Check for package declaration
	if !strings.Contains(code, "package main") {
		return "missing package main declaration"
	}

	// Check for main function
	if !regexp.MustCompile(`func\s+main\s*\(`).MatchString(code) {
		return "missing main function"
	}

	return ""
}
