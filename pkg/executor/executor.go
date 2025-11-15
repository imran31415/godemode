package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/imran31415/godemode/pkg/compiler"
	"github.com/imran31415/godemode/pkg/validator"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Executor handles the compilation and execution of Go code in a WASM sandbox
type Executor struct {
	compiler  *compiler.Compiler
	validator *validator.Validator
	// Runtime configuration
	memoryLimitPages uint32        // Memory limit in pages (64KB per page)
	defaultTimeout   time.Duration // Default execution timeout
}

// ExecutionResult contains the results of code execution
type ExecutionResult struct {
	Stdout   string        // Captured stdout
	Stderr   string        // Captured stderr
	Duration time.Duration // Execution duration
	Success  bool          // Whether execution completed successfully
	Error    string        // Error message if any
}

// ExecutionError represents an error that occurred during WASM execution
type ExecutionError struct {
	Type    string // "timeout", "memory", "trap", "panic"
	Message string
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("execution error (%s): %s", e.Type, e.Message)
}

// NewExecutor creates a new Executor with default settings
func NewExecutor() *Executor {
	return &Executor{
		compiler:         compiler.NewCompiler(),
		validator:        validator.NewValidator(),
		memoryLimitPages: 1024, // 64MB (1024 pages * 64KB)
		defaultTimeout:   30 * time.Second,
	}
}

// NewExecutorWithConfig creates an Executor with custom configuration
func NewExecutorWithConfig(memoryLimitMB uint32, defaultTimeout time.Duration) *Executor {
	// Convert MB to pages (1 page = 64KB)
	memoryPages := (memoryLimitMB * 1024 * 1024) / (64 * 1024)

	return &Executor{
		compiler:         compiler.NewCompiler(),
		validator:        validator.NewValidator(),
		memoryLimitPages: memoryPages,
		defaultTimeout:   defaultTimeout,
	}
}

// Execute compiles and runs Go source code in a WASM sandbox
// If timeout is 0, uses the default timeout
func (e *Executor) Execute(ctx context.Context, sourceCode string, timeout time.Duration) (*ExecutionResult, error) {
	startTime := time.Now()

	// Use default timeout if none specified
	if timeout == 0 {
		timeout = e.defaultTimeout
	}

	// Step 1: Validate source code
	if err := e.validator.Validate(sourceCode); err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("validation failed: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	// Step 2: Compile to WASM
	wasmBytes, err := e.compiler.CompileToWasm(sourceCode)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("compilation failed: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	// Step 3: Execute WASM
	result, err := e.executeWasm(ctx, wasmBytes, timeout)
	result.Duration = time.Since(startTime)

	return result, err
}

// executeWasm runs compiled WASM code in the wazero runtime
func (e *Executor) executeWasm(ctx context.Context, wasmBytes []byte, timeout time.Duration) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Success: false,
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create wazero runtime with configuration
	runtimeConfig := wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).            // Enable context cancellation
		WithMemoryLimitPages(e.memoryLimitPages) // Set memory limit

	r := wazero.NewRuntimeWithConfig(execCtx, runtimeConfig)
	defer r.Close(execCtx)

	// Instantiate WASI for basic I/O support
	if _, err := wasi_snapshot_preview1.Instantiate(execCtx, r); err != nil {
		return result, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Compile the WASM module
	compiledModule, err := r.CompileModule(execCtx, wasmBytes)
	if err != nil {
		return result, fmt.Errorf("failed to compile WASM module: %w", err)
	}
	defer compiledModule.Close(execCtx)

	// Setup stdout and stderr capture
	var stdout, stderr bytes.Buffer

	// Configure module with I/O
	config := wazero.NewModuleConfig().
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithStartFunctions("_start") // WASI entry point

	// Execute the module (this calls _start function)
	_, err = r.InstantiateModule(execCtx, compiledModule, config)

	// Capture output
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Handle execution errors
	if err != nil {
		execErr := e.classifyExecutionError(err)
		result.Error = execErr.Message
		return result, execErr
	}

	result.Success = true
	return result, nil
}

// classifyExecutionError categorizes execution errors for better error reporting
func (e *Executor) classifyExecutionError(err error) *ExecutionError {
	errMsg := err.Error()

	// Check for timeout
	if errors.Is(err, context.DeadlineExceeded) {
		return &ExecutionError{
			Type:    "timeout",
			Message: "execution timeout exceeded",
		}
	}

	// Check for context cancellation
	if errors.Is(err, context.Canceled) {
		return &ExecutionError{
			Type:    "canceled",
			Message: "execution was canceled",
		}
	}

	// Check for memory errors
	if strings.Contains(errMsg, "out of bounds") ||
		strings.Contains(errMsg, "memory") {
		return &ExecutionError{
			Type:    "memory",
			Message: "memory access violation or limit exceeded",
		}
	}

	// Check for panic/trap
	if strings.Contains(errMsg, "unreachable") ||
		strings.Contains(errMsg, "trap") {
		return &ExecutionError{
			Type:    "trap",
			Message: "WebAssembly trap occurred (panic or unreachable instruction)",
		}
	}

	// Generic execution error
	return &ExecutionError{
		Type:    "runtime",
		Message: errMsg,
	}
}

// ExecuteSimple is a convenience method that executes code with default settings
func (e *Executor) ExecuteSimple(sourceCode string) (*ExecutionResult, error) {
	return e.Execute(context.Background(), sourceCode, e.defaultTimeout)
}

// SetMemoryLimit sets the memory limit in megabytes
func (e *Executor) SetMemoryLimit(limitMB uint32) {
	e.memoryLimitPages = (limitMB * 1024 * 1024) / (64 * 1024)
}

// SetDefaultTimeout sets the default execution timeout
func (e *Executor) SetDefaultTimeout(timeout time.Duration) {
	e.defaultTimeout = timeout
}

// GetCacheSize returns the number of cached compiled modules
func (e *Executor) GetCacheSize() int {
	return e.compiler.Cache().Size()
}

// ClearCache clears the compilation cache
func (e *Executor) ClearCache() {
	e.compiler.Cache().Clear()
}
