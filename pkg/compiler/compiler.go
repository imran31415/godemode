package compiler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compiler handles compilation of Go source code to WebAssembly
type Compiler struct {
	cache *Cache
}

// NewCompiler creates a new Compiler instance with caching enabled
func NewCompiler() *Compiler {
	return &Compiler{
		cache: NewCache(),
	}
}

// CompilationError represents an error that occurred during compilation
type CompilationError struct {
	Message string
	Output  string // Full compiler output
}

func (e *CompilationError) Error() string {
	return fmt.Sprintf("compilation failed: %s", e.Message)
}

// CompileToWasm compiles Go source code to WebAssembly using TinyGo
// Returns the compiled WASM bytes or an error
func (c *Compiler) CompileToWasm(sourceCode string) ([]byte, error) {
	// Check cache first
	if wasmBytes, found := c.cache.Get(sourceCode); found {
		return wasmBytes, nil
	}

	// Perform compilation
	wasmBytes, err := c.compile(sourceCode)
	if err != nil {
		return nil, err
	}

	// Cache the result
	c.cache.Set(sourceCode, wasmBytes)

	return wasmBytes, nil
}

// compile performs the actual TinyGo compilation
func (c *Compiler) compile(sourceCode string) ([]byte, error) {
	// Create temporary directory for compilation
	tmpDir, err := os.MkdirTemp("", "godemode-compile-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write source code to main.go
	sourceFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	// Prepare output file path
	wasmFile := filepath.Join(tmpDir, "main.wasm")

	// Execute TinyGo compilation
	cmd := exec.Command("tinygo", "build",
		"-o", wasmFile,
		"-target", "wasi",
		"-no-debug", // Smaller binaries
		"-opt", "z", // Optimize for size
		sourceFile,
	)

	// Capture output for error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, &CompilationError{
			Message: c.parseCompilerError(string(output)),
			Output:  string(output),
		}
	}

	// Read compiled WASM
	wasmBytes, err := os.ReadFile(wasmFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read compiled WASM: %w", err)
	}

	return wasmBytes, nil
}

// parseCompilerError extracts a user-friendly error message from compiler output
func (c *Compiler) parseCompilerError(output string) string {
	lines := strings.Split(output, "\n")

	// Find the first line that looks like an error
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// TinyGo errors typically start with file path or "error:"
		if strings.Contains(line, "error:") || strings.Contains(line, ".go:") {
			return line
		}
	}

	// If no specific error found, return first non-empty line
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}

	return "unknown compilation error"
}

// ComputeHash computes SHA256 hash of source code
// This is used for cache keys
func ComputeHash(sourceCode string) string {
	hash := sha256.Sum256([]byte(sourceCode))
	return hex.EncodeToString(hash[:])
}

// Cache returns the compiler's cache instance
func (c *Compiler) Cache() *Cache {
	return c.cache
}

// CheckTinyGo verifies that TinyGo is installed and accessible
func CheckTinyGo() error {
	cmd := exec.Command("tinygo", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tinygo not found or not executable: %w\nOutput: %s", err, output)
	}
	return nil
}
